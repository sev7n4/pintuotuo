package handlers

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

var phoneCN = regexp.MustCompile(`^1[3-9]\d{9}$`)

type otpRecord struct {
	code string
	exp  time.Time
}

var smsOTPStore sync.Map // key: normalized phone

func smsAuthEnabled() bool {
	return os.Getenv("SMS_PROVIDER") != "" || os.Getenv("MOCK_SMS") == envTrue
}

func normalizePhone(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func storeOTP(phone, code string) {
	smsOTPStore.Store(phone, otpRecord{code: code, exp: time.Now().Add(5 * time.Minute)})
}

func verifyAndConsumeOTP(phone, code string) bool {
	v, ok := smsOTPStore.Load(phone)
	if !ok {
		return false
	}
	r := v.(otpRecord)
	if time.Now().After(r.exp) {
		smsOTPStore.Delete(phone)
		return false
	}
	if r.code != code {
		return false
	}
	smsOTPStore.Delete(phone)
	return true
}

func randomDigits6() string {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "123456"
	}
	return fmt.Sprintf("%06d", 100000+n.Int64())
}

type smsSendBody struct {
	Phone string `json:"phone" binding:"required"`
	Scene string `json:"scene"`
}

// SendSMSCode POST /users/sms/send
func SendSMSCode(c *gin.Context) {
	var body smsSendBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	if !smsAuthEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "sms_not_configured",
			"message": "请设置 MOCK_SMS=true（开发）或配置 SMS_PROVIDER（生产短信网关）",
		})
		return
	}
	phone := normalizePhone(body.Phone)
	if !phoneCN.MatchString(phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_phone", "message": "请输入有效中国大陆手机号"})
		return
	}
	var code string
	if os.Getenv("MOCK_SMS") == envTrue {
		code = "123456"
		log.Printf("[MOCK_SMS] OTP for %s is fixed 123456 (dev only)", phone)
	} else {
		code = randomDigits6()
		// 生产环境：在此调用 SMS_PROVIDER 对应 SDK 发送 code
	}
	storeOTP(phone, code)
	resp := gin.H{"message": "sent", "expires_in": 300}
	if os.Getenv("MOCK_SMS") == envTrue {
		resp["debug_code"] = code
	}
	c.JSON(http.StatusOK, resp)
}

type smsRegisterBody struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

// RegisterWithSMS POST /users/sms/register — 验证短信后注册并登录（与邮箱注册一致返回 token）
func RegisterWithSMS(c *gin.Context) {
	var req smsRegisterBody
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	if !smsAuthEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "sms_not_configured"})
		return
	}
	phone := normalizePhone(req.Phone)
	if !phoneCN.MatchString(phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_phone"})
		return
	}
	if !verifyAndConsumeOTP(phone, strings.TrimSpace(req.Code)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code", "message": "验证码无效或已过期"})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var block int
	if err := db.QueryRow(`SELECT id FROM users WHERE phone = $1`, phone).Scan(&block); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "phone_exists", "message": "该手机号已注册，请直接登录"})
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	email := fmt.Sprintf("p%s@phone.pintuotuo.local", phone)
	if err := db.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&block); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user_exists"})
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	role := "user"
	if req.Role == "merchant" {
		role = "merchant"
	}
	displayName := phone
	hash := hashPassword(req.Password)

	var user models.User
	err := db.QueryRow(
		`INSERT INTO users (email, name, password_hash, role, status, phone) VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, email, name, role, created_at, updated_at`,
		email, displayName, hash, role, "active", phone,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"USER_CREATION_FAILED",
			"Failed to create user",
			http.StatusInternalServerError,
			err,
		))
		return
	}
	user.Phone = &phone

	if _, err := db.Exec(`INSERT INTO tokens (user_id, balance) VALUES ($1, $2)`, user.ID, 0); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if role == "merchant" {
		_, _ = db.Exec(`INSERT INTO merchants (user_id, company_name, status) VALUES ($1, $2, $3)`, user.ID, displayName, "pending")
	}

	token := generateToken(user.ID, user.Email, user.Role)
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})
}

type smsLoginBody struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// LoginWithSMS POST /users/sms/login — 验证码登录（无需密码）
func LoginWithSMS(c *gin.Context) {
	var req smsLoginBody
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	if !smsAuthEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "sms_not_configured"})
		return
	}
	phone := normalizePhone(req.Phone)
	if !phoneCN.MatchString(phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_phone"})
		return
	}
	if !verifyAndConsumeOTP(phone, strings.TrimSpace(req.Code)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code", "message": "验证码无效或已过期"})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var user models.User
	var phoneCol sql.NullString
	err := db.QueryRow(
		`SELECT id, email, name, role, created_at, updated_at, phone FROM users WHERE phone = $1`,
		phone,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt, &phoneCol)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			middleware.RespondWithError(c, apperrors.ErrInvalidCredentials)
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if phoneCol.Valid && phoneCol.String != "" {
		user.Phone = &phoneCol.String
	}

	token := generateToken(user.ID, user.Email, user.Role)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"user":  user,
			"token": token,
		},
	})
}
