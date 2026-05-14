package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyProxyWhitelistClientHeaders(t *testing.T) {
	t.Setenv(envProxyForwardClientHeaders, "")
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/anthropic/v1/messages", nil)
	req.Header.Set("Anthropic-Beta", "tools-2024-04-04")
	req.Header.Set("Openai-Organization", "org-from-client")
	req.Header.Set("Authorization", "Bearer client-should-not-copy")
	c.Request = req

	hreq, err := http.NewRequest(http.MethodPost, "https://example.com/v1/messages", nil)
	require.NoError(t, err)
	hreq.Header.Set("X-Api-Key", "upstream-secret")

	applyProxyWhitelistClientHeaders(c, hreq)

	assert.Equal(t, "upstream-secret", hreq.Header.Get("X-Api-Key"))
	assert.Equal(t, "tools-2024-04-04", hreq.Header.Get("Anthropic-Beta"))
	assert.Equal(t, "org-from-client", hreq.Header.Get("Openai-Organization"))
	assert.Empty(t, hreq.Header.Get("Authorization"))
}

func TestApplyProxyWhitelistClientHeaders_Disabled(t *testing.T) {
	t.Setenv(envProxyForwardClientHeaders, "false")
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Anthropic-Beta", "x")
	c.Request = req

	hreq, err := http.NewRequest(http.MethodPost, "https://u/m", nil)
	require.NoError(t, err)
	applyProxyWhitelistClientHeaders(c, hreq)
	assert.Empty(t, hreq.Header.Get("Anthropic-Beta"))
}

func TestApplyProxyOutboundAuthHeaders_AnthropicDefaults(t *testing.T) {
	t.Setenv(envProxyForwardClientHeaders, "")
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	hreq, err := http.NewRequest(http.MethodPost, "https://glm/messages", nil)
	require.NoError(t, err)
	pk := models.MerchantAPIKey{RouteMode: "direct"}

	applyProxyOutboundAuthHeaders(c, hreq, "req-1", true, "merchant-key", &pk, false)

	assert.Equal(t, "merchant-key", hreq.Header.Get("X-Api-Key"))
	assert.Equal(t, "2023-06-01", hreq.Header.Get("Anthropic-Version"))
}

func TestApplyProxyOutboundAuthHeaders_AnthropicVersionFromClient(t *testing.T) {
	t.Setenv(envProxyForwardClientHeaders, "")
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Anthropic-Version", "2024-11-01")
	c.Request = req

	hreq, err := http.NewRequest(http.MethodPost, "https://glm/messages", nil)
	require.NoError(t, err)
	pk := models.MerchantAPIKey{RouteMode: "direct"}

	applyProxyOutboundAuthHeaders(c, hreq, "req-2", true, "mk", &pk, false)

	assert.Equal(t, "2024-11-01", hreq.Header.Get("Anthropic-Version"))
}

func TestApplyProxyOutboundAuthHeaders_OpenAIUsesBearer(t *testing.T) {
	t.Setenv(envProxyForwardClientHeaders, "")
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Openai-Beta", "assistants=v2")
	c.Request = req

	hreq, err := http.NewRequest(http.MethodPost, "https://oai/v1/chat/completions", nil)
	require.NoError(t, err)
	pk := models.MerchantAPIKey{RouteMode: "direct"}

	applyProxyOutboundAuthHeaders(c, hreq, "r", false, "sk-upstream", &pk, false)

	assert.Contains(t, hreq.Header.Get("Authorization"), "sk-upstream")
	assert.Equal(t, "assistants=v2", hreq.Header.Get("Openai-Beta"))
	assert.Empty(t, hreq.Header.Get("Anthropic-Version"))
}

func TestApplyProxyWhitelistClientHeaders_ExtraFromEnv(t *testing.T) {
	t.Setenv(envProxyForwardClientHeaders, "")
	t.Setenv(envProxyForwardExtraHeaders, "X-Gateway-Stage")
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Gateway-Stage", "canary")
	c.Request = req

	hreq, err := http.NewRequest(http.MethodPost, "https://u/m", nil)
	require.NoError(t, err)
	applyProxyWhitelistClientHeaders(c, hreq)
	assert.Equal(t, "canary", hreq.Header.Get("X-Gateway-Stage"))
}

func TestApplyProxyWhitelistClientHeaders_ExtraEnvDoesNotBypassBlocklist(t *testing.T) {
	t.Setenv(envProxyForwardClientHeaders, "")
	t.Setenv(envProxyForwardExtraHeaders, "Authorization")
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer evil")
	c.Request = req

	hreq, err := http.NewRequest(http.MethodPost, "https://u/m", nil)
	require.NoError(t, err)
	hreq.Header.Set("Authorization", "Bearer upstream")
	applyProxyWhitelistClientHeaders(c, hreq)
	assert.Equal(t, "Bearer upstream", hreq.Header.Get("Authorization"))
}
