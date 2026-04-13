package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/models"
)

type magicLinkRecord struct {
	Email string
	Exp   time.Time
}

var magicLinkStore sync.Map // key: token, val: magicLinkRecord

const (
	magicTokenTTL = 15 * time.Minute
)

func magicLinkEnabled() bool {
	return os.Getenv("AUTH_MAGIC_LINK") == envTrue
}

func magicLinkMockEnabled() bool {
	return os.Getenv("EMAIL_MAGIC_MOCK") == envTrue
}

func emailSenderConfigured() bool {
	return strings.TrimSpace(os.Getenv("SMTP_HOST")) != "" &&
		strings.TrimSpace(os.Getenv("SMTP_PORT")) != "" &&
		strings.TrimSpace(os.Getenv("SMTP_FROM")) != ""
}

func randomTokenURLSafe(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func sendMagicLinkEmail(to, link string) error {
	to = strings.TrimSpace(to)
	if to == "" || strings.ContainsAny(to, "\r\n") {
		return fmt.Errorf("invalid recipient")
	}
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := os.Getenv("SMTP_PASS")
	if pass == "" {
		pass = os.Getenv("SMTP_PASSWORD")
	}
	if host == "" || port == "" || from == "" {
		return fmt.Errorf("smtp not configured")
	}

	subject := "拼脱脱登录链接"
	body := fmt.Sprintf("请在 15 分钟内点击登录：\n\n%s\n\n若非本人操作请忽略。", link)
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body)

	addr := host + ":" + port
	var auth smtp.Auth
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	// to 已校验无换行；发件人来自环境变量（运维配置）
	return smtp.SendMail(addr, auth, from, []string{to}, msg) // #nosec G707 -- recipient sanitized, no header injection
}

func createOrGetEmailUser(db *sql.DB, email string) (*models.User, error) {
	var u models.User
	err := db.QueryRow(
		`SELECT id, email, name, role, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		return &u, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	displayName := registrationDisplayName(email, "")
	pwd := hashPassword(randomOAuthPassword())
	err = tx.QueryRow(
		`INSERT INTO users (email, name, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, email, name, role, created_at, updated_at`,
		email, displayName, pwd, roleUser, "active",
	).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`INSERT INTO tokens (user_id, balance) VALUES ($1, $2)`, u.ID, 0); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &u, nil
}

type emailMagicSendBody struct {
	Email string `json:"email" binding:"required,email"`
}

// SendEmailMagicLink POST /users/email/magic/send
func SendEmailMagicLink(c *gin.Context) {
	var req emailMagicSendBody
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	if !magicLinkEnabled() {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "magic_link_disabled",
			"message": "AUTH_MAGIC_LINK 未开启",
		})
		return
	}
	addr, err := mail.ParseAddress(strings.TrimSpace(req.Email))
	if err != nil || addr.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_email", "message": "邮箱格式不正确"})
		return
	}

	token := randomTokenURLSafe(32)
	magicLinkStore.Store(token, magicLinkRecord{
		Email: strings.ToLower(addr.Address),
		Exp:   time.Now().Add(magicTokenTTL),
	})

	// 回调应请求后端，避免暴露 JWT 签发逻辑到前端路由。
	var verifyURL string
	apiBase := strings.TrimSpace(os.Getenv("PUBLIC_API_BASE_URL"))
	if apiBase != "" {
		verifyURL = fmt.Sprintf("%s/api/v1/users/email/magic/verify?token=%s", strings.TrimRight(apiBase, "/"), token)
	} else {
		apiBase = strings.TrimRight(getEnv("FRONTEND_URL", "http://localhost:5173"), "/")
		// 当 FRONTEND_URL 指向前端站点时，默认约定 API 同域 /api/v1（Nginx 反代）。
		verifyURL = fmt.Sprintf("%s/api/v1/users/email/magic/verify?token=%s", apiBase, token)
	}

	if emailSenderConfigured() {
		if err := sendMagicLinkEmail(addr.Address, verifyURL); err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"EMAIL_SEND_FAILED",
				"发送邮件失败",
				http.StatusInternalServerError,
				err,
			))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "magic_link_sent"})
		return
	}

	if magicLinkMockEnabled() {
		log.Printf("[EMAIL_MAGIC_MOCK] Magic link for %s: %s", addr.Address, verifyURL)
		c.JSON(http.StatusOK, gin.H{
			"message":    "magic_link_sent",
			"debug_link": verifyURL,
		})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "email_sender_not_configured",
		"message": "请配置 SMTP_* 或 EMAIL_MAGIC_MOCK=true",
	})
}

// VerifyEmailMagicLink GET /users/email/magic/verify?token=...
func VerifyEmailMagicLink(c *gin.Context) {
	if !magicLinkEnabled() {
		redirectOAuthError(c, "magic_link_disabled")
		return
	}
	tok := strings.TrimSpace(c.Query("token"))
	if tok == "" {
		redirectOAuthError(c, "missing_magic_token")
		return
	}

	v, ok := magicLinkStore.Load(tok)
	if !ok {
		redirectOAuthError(c, "invalid_magic_token")
		return
	}
	rec := v.(magicLinkRecord)
	magicLinkStore.Delete(tok)
	if time.Now().After(rec.Exp) {
		redirectOAuthError(c, "expired_magic_token")
		return
	}

	db := config.GetDB()
	if db == nil {
		redirectOAuthError(c, "database_unavailable")
		return
	}
	u, err := createOrGetEmailUser(db, rec.Email)
	if err != nil {
		redirectOAuthError(c, "magic_link_user_failed")
		return
	}
	redirectOAuthSuccess(c, generateToken(u.ID, u.Email, u.Role))
}
