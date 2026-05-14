package handlers

import (
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

// Anthropic Messages API stop_reason 常用值（与 goconst / 客户端约定一致）
const (
	anthropicStopEndTurn   = "end_turn"
	anthropicStopMaxTokens = "max_tokens"
)

// ctxKeyAnthropicCompat 标记请求来自 Anthropic Messages HTTP 客户端（用于流式分支与响应形态；原生上游时仍透传 SSE/JSON）。
const ctxKeyAnthropicCompat = "ptd_anthropic_compat"

type anthropicCompatCtx struct {
	ClientModel string
}

func anthropicCompatFromContext(c *gin.Context) (anthropicCompatCtx, bool) {
	v, ok := c.Get(ctxKeyAnthropicCompat)
	if !ok {
		return anthropicCompatCtx{}, false
	}
	ac, ok := v.(anthropicCompatCtx)
	return ac, ok
}

// AnthropicMessages 接受 Anthropic Messages API 原始 JSON，在 api_format=anthropic 的上游原样转发（仅改写 model；LiteLLM 注入 user_config）。
// 若上游为 OpenAI 兼容而客户端走本路由，请改用 POST …/openai/v1/chat/completions。
// Base URL：{API_ORIGIN}/api/v1/anthropic/v1；鉴权：Authorization: Bearer ptd_* 或 x-api-key: ptd_*。
// 出站头：Anthropic-Beta、Anthropic-Version（及 Openai-* 等）默认白名单透传至上游，见 applyProxyOutboundAuthHeaders；可设 API_PROXY_FORWARD_EXTRA_HEADERS 追加、API_PROXY_FORWARD_CLIENT_HEADERS=false 全关。
func AnthropicMessages(c *gin.Context) {
	bodyBytes, readErr := io.ReadAll(c.Request.Body)
	if readErr != nil {
		respondAnthropicInvalidRequest(c, "Failed to read request body")
		return
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		respondAnthropicInvalidRequest(c, "Invalid JSON body")
		return
	}
	modelVal, ok := raw["model"].(string)
	if !ok || strings.TrimSpace(modelVal) == "" {
		respondAnthropicInvalidRequest(c, "model is required")
		return
	}
	maxTok, ok := raw["max_tokens"]
	if !ok {
		respondAnthropicInvalidRequest(c, "max_tokens is required and must be positive")
		return
	}
	maxN, ok := anthropicJSONNumberToInt(maxTok)
	if !ok || maxN <= 0 {
		respondAnthropicInvalidRequest(c, "max_tokens is required and must be positive")
		return
	}
	_ = maxN // 校验存在即可，完整负载在 RawAnthropicMessagesBody 中透传
	msgs, ok := raw["messages"]
	if !ok {
		respondAnthropicInvalidRequest(c, "messages is required")
		return
	}
	arr, ok := msgs.([]interface{})
	if !ok || len(arr) == 0 {
		respondAnthropicInvalidRequest(c, "messages must be a non-empty array")
		return
	}

	stream := false
	if s, streamOk := raw["stream"].(bool); streamOk {
		stream = s
	}

	db := config.GetDB()
	provider, modelName := services.ResolveOpenAICompatModel(db, modelVal)
	if provider == "" || modelName == "" {
		respondAnthropicInvalidRequest(c, "Could not resolve provider from model")
		return
	}

	rawBody := json.RawMessage(append(json.RawMessage(nil), bodyBytes...))
	req := APIProxyRequest{
		Provider:                 provider,
		Model:                    modelName,
		Messages:                 nil,
		Stream:                   stream,
		Options:                  nil,
		RawAnthropicMessagesBody: rawBody,
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

	c.Set(ctxKeyAnthropicCompat, anthropicCompatCtx{ClientModel: modelVal})
	startTime := time.Now()
	requestID := uuid.New().String()
	proxyAPIRequestCore(c, userIDInt, requestID, startTime, req, c.Request.URL.Path)
}

func respondAnthropicInvalidRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"type": "error",
		"error": gin.H{
			"type":    "invalid_request_error",
			"message": msg,
		},
	})
}

// AnthropicListModels 提供与 Anthropic「模型列表」相近的 JSON 形态，id 与 GET /openai/v1/models 一致（如 alibaba/qwen-turbo）。
func AnthropicListModels(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		respondAnthropicInvalidRequest(c, "database unavailable")
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
		resp, listErr = services.ListOpenAIModelsEntitledForUser(c.Request.Context(), db, userIDInt)
	} else {
		resp, listErr = services.ListOpenAIModelsFromCatalog(c.Request.Context(), db)
	}
	if listErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "api_error",
				"message": "Failed to list models",
			},
		})
		return
	}
	items := make([]gin.H, 0, len(resp.Data))
	for _, it := range resp.Data {
		row := gin.H{
			"id":   it.ID,
			"type": "model",
		}
		if it.OwnedBy != "" {
			row["display_name"] = it.ID
		}
		items = append(items, row)
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "has_more": false})
}

func anthropicJSONNumberToInt(v interface{}) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	case int64:
		return int(x), true
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}

func openAICompletionToAnthropicMessage(body []byte, clientModel string) (map[string]interface{}, error) {
	var apiResp APIProxyResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}
	text := ""
	if len(apiResp.Choices) > 0 && apiResp.Choices[0].Message != nil {
		text = apiResp.Choices[0].Message.Content
	}
	stopReason := anthropicStopEndTurn
	if len(apiResp.Choices) > 0 {
		switch apiResp.Choices[0].FinishReason {
		case "length":
			stopReason = anthropicStopMaxTokens
		case "stop", "":
			stopReason = anthropicStopEndTurn
		default:
			stopReason = apiResp.Choices[0].FinishReason
		}
	}
	return map[string]interface{}{
		"id":            "msg_" + strings.ReplaceAll(uuid.New().String(), "-", ""),
		"type":          "message",
		"role":          "assistant",
		"model":         clientModel,
		"content":       []map[string]interface{}{{"type": "text", "text": text}},
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage": map[string]interface{}{
			"input_tokens":  apiResp.Usage.PromptTokens,
			"output_tokens": apiResp.Usage.CompletionTokens,
		},
	}, nil
}
