package handlers

import (
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/utils"
)

// PostAdminTOTPSetup POST /users/me/mfa/totp/setup — 仅 admin，生成密钥（未启用 MFA 前可重复调用覆盖待确认密钥）
func PostAdminTOTPSetup(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError("FORBIDDEN", "仅运营账号可配置 MFA", http.StatusForbidden, nil))
		return
	}
	userIDVal, ok := c.Get("user_id")
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userID, ok := userIDVal.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	secret := strings.TrimRight(base32.StdEncoding.EncodeToString(secretBytes), "=")
	enc, err := utils.Encrypt(secret)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	_, err = db.Exec(`UPDATE users SET mfa_totp_secret_enc = $1, mfa_enabled = false WHERE id = $2`, enc, userID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	issuer := strings.TrimSpace(os.Getenv("MFA_TOTP_ISSUER"))
	if issuer == "" {
		issuer = "Pintuotuo"
	}
	email := ""
	_ = db.QueryRow(`SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
	uri := "otpauth://totp/" + issuer + ":" + email + "?secret=" + secret + "&issuer=" + issuer
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"secret":      secret,
			"otpauth_uri": uri,
		},
	})
}

type totpConfirmBody struct {
	Code string `json:"code" binding:"required"`
}

// PostAdminTOTPConfirm POST /users/me/mfa/totp/confirm — 校验一次 TOTP 后启用 MFA
func PostAdminTOTPConfirm(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError("FORBIDDEN", "仅运营账号可配置 MFA", http.StatusForbidden, nil))
		return
	}
	var body totpConfirmBody
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	userIDVal, ok := c.Get("user_id")
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userID, ok := userIDVal.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	var enc string
	err := db.QueryRow(`SELECT COALESCE(mfa_totp_secret_enc,'') FROM users WHERE id = $1`, userID).Scan(&enc)
	if err != nil || enc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "MFA_NOT_SETUP", "message": "请先调用 MFA 初始化接口"})
		return
	}
	secret, err := utils.Decrypt(enc)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !utils.ValidateTOTP(strings.TrimSpace(body.Code), secret) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_TOTP", "message": "验证码不正确"})
		return
	}
	_, err = db.Exec(`UPDATE users SET mfa_enabled = true WHERE id = $1`, userID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "MFA 已启用"})
}
