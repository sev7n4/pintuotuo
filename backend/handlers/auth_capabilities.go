package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

// AuthCapabilities 描述当前环境已启用的扩展认证能力（供前端展示/灰度）。
type AuthCapabilities struct {
	SMS                  bool   `json:"sms"`
	EmailMagic           bool   `json:"email_magic"`
	WechatOAuth          bool   `json:"wechat_oauth"`
	GithubOAuth          bool   `json:"github_oauth"`
	AccountLinking       bool   `json:"account_linking"`
	MerchantRegisterMode string `json:"merchant_register_mode"` // open | invite_only | hidden
	AdminMFARequired     bool   `json:"admin_mfa_required"`
}

// GetAuthCapabilities GET /users/auth/capabilities（无需登录）
func GetAuthCapabilities(c *gin.Context) {
	mockSMS := os.Getenv("MOCK_SMS") == envTrue
	mode := merchantRegisterMode()
	cap := AuthCapabilities{
		SMS:                  os.Getenv("SMS_PROVIDER") != "" || mockSMS,
		EmailMagic:           os.Getenv("AUTH_MAGIC_LINK") == envTrue,
		WechatOAuth:          os.Getenv("WECHAT_OPEN_APP_ID") != "",
		GithubOAuth:          os.Getenv("GITHUB_OAUTH_CLIENT_ID") != "",
		AccountLinking:       os.Getenv("AUTH_ACCOUNT_LINKING") == envTrue || os.Getenv("GITHUB_OAUTH_CLIENT_ID") != "" || os.Getenv("WECHAT_OPEN_APP_ID") != "",
		MerchantRegisterMode: mode,
		AdminMFARequired:     adminMFARequired(),
	}
	c.JSON(http.StatusOK, cap)
}

// WechatOAuthStart GET /users/oauth/wechat/start — 跳转微信开放平台扫码页
func WechatOAuthStart(c *gin.Context) {
	appID := os.Getenv("WECHAT_OPEN_APP_ID")
	if appID == "" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "wechat_oauth_not_configured",
			"message": "未配置 WECHAT_OPEN_APP_ID",
		})
		return
	}
	redirectURI := os.Getenv("WECHAT_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WECHAT_OAUTH_REDIRECT_URI required"})
		return
	}
	state := c.Query("state")
	if state == "" {
		state = "pintuotuo"
	}
	u := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect",
		appID, url.QueryEscape(redirectURI), state,
	)
	c.Redirect(http.StatusFound, u)
}

// GithubOAuthStart GET /users/oauth/github/start — 跳转 GitHub 授权页
func GithubOAuthStart(c *gin.Context) {
	clientID := os.Getenv("GITHUB_OAUTH_CLIENT_ID")
	if clientID == "" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "github_oauth_not_configured",
			"message": "未配置 GITHUB_OAUTH_CLIENT_ID",
		})
		return
	}
	redirectURI := os.Getenv("GITHUB_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GITHUB_OAUTH_REDIRECT_URI required"})
		return
	}
	scope := os.Getenv("GITHUB_OAUTH_SCOPE")
	if scope == "" {
		scope = "read:user user:email"
	}
	u := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s",
		clientID, url.QueryEscape(redirectURI), url.QueryEscape(scope),
	)
	c.Redirect(http.StatusFound, u)
}

// UserIdentity 第三方绑定记录（仅展示，不含敏感 token）
type UserIdentity struct {
	Provider    string `json:"provider"`
	ExternalID  string `json:"external_id"`
	DisplayName string `json:"display_name,omitempty"`
}

// GetUserIdentities GET /users/me/identities
func GetUserIdentities(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userIDVal.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": []UserIdentity{}})
		return
	}
	rows, err := db.Query(
		`SELECT provider, external_id, COALESCE(display_name,'') FROM user_identity_links WHERE user_id = $1 ORDER BY id ASC`,
		userIDInt,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": []UserIdentity{}})
		return
	}
	defer rows.Close()
	var list []UserIdentity
	for rows.Next() {
		var id UserIdentity
		if err := rows.Scan(&id.Provider, &id.ExternalID, &id.DisplayName); err != nil {
			continue
		}
		list = append(list, id)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}
