package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

func respondOpenAIError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "invalid_request_error",
		},
	})
}

// OpenAIChatCompletions accepts an OpenAI-compatible JSON body and reuses the platform proxy pipeline.
// 请求体除 model 由平台解析并重写为上游模型 id 外，整体原样转发至 OpenAI 兼容上游（含 tools、多模态 messages、stream 等）。
// LiteLLM 路由下由平台注入 user_config，客户端同名字段会被剥离后覆盖。
// Supports stream:true (SSE) for OpenAI-compatible providers; see deploy/litellm/README.md.
// Clients should set base URL to {API_ORIGIN}/api/v1/openai/v1 (OpenAI SDK: baseURL + "/chat/completions").
// Authentication: Bearer platform API key (ptd_* / ptt_*) or JWT（与 /openai/v1 其它路由一致）。
// 出站头：Openai-Beta、Openai-Organization 等默认白名单透传至上游（与 api_proxy 一致）；见 proxy_upstream_headers.go 环境变量说明。
func OpenAIChatCompletions(c *gin.Context) {
	bodyBytes, readErr := io.ReadAll(c.Request.Body)
	if readErr != nil {
		respondOpenAIError(c, http.StatusBadRequest, "Failed to read request body")
		return
	}

	var raw map[string]interface{}
	if unmarshalErr := json.Unmarshal(bodyBytes, &raw); unmarshalErr != nil {
		respondOpenAIError(c, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	modelVal, ok := raw["model"].(string)
	if !ok || strings.TrimSpace(modelVal) == "" {
		respondOpenAIError(c, http.StatusBadRequest, "model is required")
		return
	}
	msgs, ok := raw["messages"]
	if !ok {
		respondOpenAIError(c, http.StatusBadRequest, "messages is required")
		return
	}
	arr, ok := msgs.([]interface{})
	if !ok || len(arr) == 0 {
		respondOpenAIError(c, http.StatusBadRequest, "messages must be a non-empty array")
		return
	}

	stream := false
	if streamVal, streamOk := raw["stream"].(bool); streamOk {
		stream = streamVal
	}

	db := config.GetDB()
	provider, modelName := services.ResolveOpenAICompatModel(db, modelVal)
	if provider == "" || modelName == "" {
		respondOpenAIError(c, http.StatusBadRequest, "Could not resolve provider from model")
		return
	}

	rawBody := json.RawMessage(append(json.RawMessage(nil), bodyBytes...))
	req := APIProxyRequest{
		Provider:          provider,
		Model:             modelName,
		Messages:          nil,
		Stream:            stream,
		Options:           nil,
		RawOpenAIChatBody: rawBody,
	}

	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	startTime := time.Now()
	requestID := uuid.New().String()
	proxyAPIRequestCore(c, userIDInt, requestID, startTime, req, c.Request.URL.Path)
}

// OpenAIListModels exposes GET /v1/models for OpenAI SDK discovery (same base URL as chat/completions).
// When ENTITLEMENT_ENFORCEMENT=strict, the list is filtered to models the caller is entitled to (same scope as chat).
func OpenAIListModels(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	var resp *services.OpenAIModelsListResponse
	var listErr error
	if services.EntitlementEnforcementStrict() {
		userID, exists := c.Get("user_id")
		if !exists {
			middleware.RespondWithError(c, apperrors.ErrInvalidToken)
			return
		}
		userIDInt, ok := userID.(int)
		if !ok {
			middleware.RespondWithError(c, apperrors.ErrInvalidToken)
			return
		}
		resp, listErr = services.ListOpenAIModelsEntitledForUser(context.Background(), db, userIDInt)
	} else {
		resp, listErr = services.ListOpenAIModelsFromCatalog(context.Background(), db)
	}
	if listErr != nil {
		respondOpenAIError(c, http.StatusInternalServerError, "Failed to list models")
		return
	}
	c.JSON(http.StatusOK, resp)
}
