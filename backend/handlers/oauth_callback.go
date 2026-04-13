package handlers

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

const (
	oauthProviderGitHub = "github"
	oauthProviderWechat = "wechat"
)

func oauthFrontendBase() string {
	b := strings.TrimSpace(os.Getenv("FRONTEND_URL"))
	if b == "" {
		return "http://localhost:5173"
	}
	return strings.TrimRight(b, "/")
}

func redirectOAuthSuccess(c *gin.Context, jwt string) {
	u := fmt.Sprintf("%s/login?oauth=1&token=%s", oauthFrontendBase(), url.QueryEscape(jwt))
	c.Redirect(http.StatusFound, u)
}

func redirectOAuthError(c *gin.Context, msg string) {
	u := fmt.Sprintf("%s/login?oauth_error=%s", oauthFrontendBase(), url.QueryEscape(msg))
	c.Redirect(http.StatusFound, u)
}

func oauthHTTPClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}

func randomOAuthPassword() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("oauth-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}

// GithubOAuthCallback GET /users/oauth/github/callback — code 换 token，登录或注册后重定向前端并附带 JWT
func GithubOAuthCallback(c *gin.Context) {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		redirectOAuthError(c, "missing_code")
		return
	}
	clientID := os.Getenv("GITHUB_OAUTH_CLIENT_ID")
	secret := os.Getenv("GITHUB_OAUTH_CLIENT_SECRET")
	redir := os.Getenv("GITHUB_OAUTH_REDIRECT_URI")
	if clientID == "" || secret == "" || redir == "" {
		redirectOAuthError(c, "github_oauth_not_configured")
		return
	}

	body := map[string]string{
		"client_id":     clientID,
		"client_secret": secret,
		"code":          code,
		"redirect_uri":  redir,
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewReader(raw))
	if err != nil {
		redirectOAuthError(c, "github_token_request")
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauthHTTPClient().Do(req)
	if err != nil {
		redirectOAuthError(c, "github_token_network")
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		redirectOAuthError(c, "github_token_read")
		return
	}

	var tr struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.Unmarshal(respBody, &tr); err != nil || tr.AccessToken == "" {
		redirectOAuthError(c, "github_token_invalid")
		return
	}

	reqUser, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		redirectOAuthError(c, "github_user_request")
		return
	}
	reqUser.Header.Set("Authorization", "Bearer "+tr.AccessToken)
	reqUser.Header.Set("Accept", "application/vnd.github+json")

	respU, err := oauthHTTPClient().Do(reqUser)
	if err != nil {
		redirectOAuthError(c, "github_user_network")
		return
	}
	defer respU.Body.Close()
	userBody, err := io.ReadAll(respU.Body)
	if err != nil || respU.StatusCode != http.StatusOK {
		redirectOAuthError(c, "github_user_invalid")
		return
	}

	var gu struct {
		ID    int64   `json:"id"`
		Login string  `json:"login"`
		Name  *string `json:"name"`
	}
	if err := json.Unmarshal(userBody, &gu); err != nil || gu.ID == 0 {
		redirectOAuthError(c, "github_user_parse")
		return
	}

	extID := strconv.FormatInt(gu.ID, 10)
	display := gu.Login
	if gu.Name != nil && strings.TrimSpace(*gu.Name) != "" {
		display = strings.TrimSpace(*gu.Name)
	} else if display == "" {
		display = "GitHub User"
	}

	syntheticEmail := fmt.Sprintf("gh-%s@oauth.pintuotuo.local", extID)

	db := config.GetDB()
	if db == nil {
		redirectOAuthError(c, "database_unavailable")
		return
	}

	user, jwt, err := upsertOAuthUser(db, oauthProviderGitHub, extID, syntheticEmail, display)
	if err != nil {
		redirectOAuthError(c, "oauth_user_failed")
		return
	}
	_ = user
	redirectOAuthSuccess(c, jwt)
}

// WechatOAuthCallback GET /users/oauth/wechat/callback — 微信开放平台扫码登录回调
func WechatOAuthCallback(c *gin.Context) {
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		redirectOAuthError(c, "missing_code")
		return
	}
	appID := os.Getenv("WECHAT_OPEN_APP_ID")
	secret := os.Getenv("WECHAT_OPEN_APP_SECRET")
	if appID == "" || secret == "" {
		redirectOAuthError(c, "wechat_oauth_not_configured")
		return
	}

	u := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		url.QueryEscape(appID),
		url.QueryEscape(secret),
		url.QueryEscape(code),
	)
	resp, err := oauthHTTPClient().Get(u)
	if err != nil {
		redirectOAuthError(c, "wechat_token_network")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		redirectOAuthError(c, "wechat_token_read")
		return
	}

	var wxTok struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		OpenID       string `json:"openid"`
		Scope        string `json:"scope"`
		Unionid      string `json:"unionid"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &wxTok); err != nil {
		redirectOAuthError(c, "wechat_token_parse")
		return
	}
	if wxTok.ErrCode != 0 || wxTok.AccessToken == "" || wxTok.OpenID == "" {
		redirectOAuthError(c, "wechat_token_invalid")
		return
	}

	extID := wxTok.Unionid
	if extID == "" {
		extID = wxTok.OpenID
	}
	syntheticEmail := fmt.Sprintf("wx-%s@oauth.pintuotuo.local", sanitizeEmailLocal(extID))

	infoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN",
		url.QueryEscape(wxTok.AccessToken),
		url.QueryEscape(wxTok.OpenID),
	)
	respI, err := oauthHTTPClient().Get(infoURL)
	if err != nil {
		redirectOAuthError(c, "wechat_userinfo_network")
		return
	}
	defer respI.Body.Close()
	infoBody, err := io.ReadAll(respI.Body)
	if err != nil {
		redirectOAuthError(c, "wechat_userinfo_read")
		return
	}

	var wxUser struct {
		Nickname string `json:"nickname"`
		OpenID   string `json:"openid"`
		Unionid  string `json:"unionid"`
		ErrCode  int    `json:"errcode"`
		ErrMsg   string `json:"errmsg"`
	}
	if err := json.Unmarshal(infoBody, &wxUser); err != nil {
		redirectOAuthError(c, "wechat_userinfo_parse")
		return
	}
	if wxUser.ErrCode != 0 {
		redirectOAuthError(c, "wechat_userinfo_error")
		return
	}
	display := strings.TrimSpace(wxUser.Nickname)
	if display == "" {
		display = "微信用户"
	}

	db := config.GetDB()
	if db == nil {
		redirectOAuthError(c, "database_unavailable")
		return
	}

	user, jwt, err := upsertOAuthUser(db, oauthProviderWechat, extID, syntheticEmail, display)
	if err != nil {
		redirectOAuthError(c, "oauth_user_failed")
		return
	}
	_ = user
	redirectOAuthSuccess(c, jwt)
}

func sanitizeEmailLocal(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "id"
	}
	return out
}

func upsertOAuthUser(db *sql.DB, provider, externalID, email, displayName string) (*models.User, string, error) {

	var u models.User
	err := db.QueryRow(`
		SELECT u.id, u.email, u.name, u.role, u.created_at, u.updated_at
		FROM users u
		INNER JOIN user_identity_links l ON l.user_id = u.id
		WHERE l.provider = $1 AND l.external_id = $2
	`, provider, externalID).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		return &u, generateToken(u.ID, u.Email, u.Role), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, "", err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = tx.Rollback() }()

	pwd := hashPassword(randomOAuthPassword())
	err = tx.QueryRow(`
		INSERT INTO users (email, name, password_hash, role, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, name, role, created_at, updated_at`,
		email, displayName, pwd, roleUser, "active",
	).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, "", err
	}

	if _, err := tx.Exec(`INSERT INTO tokens (user_id, balance) VALUES ($1, $2)`, u.ID, 0); err != nil {
		return nil, "", err
	}

	if _, err := tx.Exec(
		`INSERT INTO user_identity_links (user_id, provider, external_id, display_name) VALUES ($1, $2, $3, $4)`,
		u.ID, provider, externalID, displayName,
	); err != nil {
		return nil, "", err
	}

	if err := tx.Commit(); err != nil {
		return nil, "", err
	}

	return &u, generateToken(u.ID, u.Email, u.Role), nil
}
