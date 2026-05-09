package handlers

import (
	"encoding/json"
	"fmt"
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

// ctxKeyAnthropicCompat 标记本请求需将上游 OpenAI 兼容结果转为 Anthropic Messages API 形态。
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

// --- Anthropic POST /v1/messages 请求体（子集，满足 Claude Code / SDK 常见字段） ---

type anthropicMessagesRequest struct {
	Model         string          `json:"model"`
	MaxTokens     int             `json:"max_tokens"`
	Messages      json.RawMessage `json:"messages"`
	System        json.RawMessage `json:"system,omitempty"`
	Stream        bool            `json:"stream"`
	Temperature   *float64        `json:"temperature,omitempty"`
	TopP          *float64        `json:"top_p,omitempty"`
	TopK          *int            `json:"top_k,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
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

// AnthropicMessages 将 Anthropic Messages 请求转为内部 OpenAI 兼容代理链路（含流式 SSE 的协议转换）。
// Base URL：{API_ORIGIN}/api/v1/anthropic/v1；鉴权：Authorization: Bearer ptd_* 或 x-api-key: ptd_*。
func AnthropicMessages(c *gin.Context) {
	bodyBytes, readErr := io.ReadAll(c.Request.Body)
	if readErr != nil {
		respondAnthropicInvalidRequest(c, "Failed to read request body")
		return
	}
	var in anthropicMessagesRequest
	if err := json.Unmarshal(bodyBytes, &in); err != nil {
		respondAnthropicInvalidRequest(c, "Invalid JSON body")
		return
	}
	modelVal := strings.TrimSpace(in.Model)
	if modelVal == "" {
		respondAnthropicInvalidRequest(c, "model is required")
		return
	}
	if in.MaxTokens <= 0 {
		respondAnthropicInvalidRequest(c, "max_tokens is required and must be positive")
		return
	}
	msgs, err := anthropicToChatMessages(in.System, in.Messages)
	if err != nil {
		respondAnthropicInvalidRequest(c, err.Error())
		return
	}
	if len(msgs) == 0 {
		respondAnthropicInvalidRequest(c, "messages must be a non-empty array")
		return
	}

	db := config.GetDB()
	provider, modelName := services.ResolveOpenAICompatModel(db, modelVal)
	if provider == "" || modelName == "" {
		respondAnthropicInvalidRequest(c, "Could not resolve provider from model")
		return
	}

	opts := map[string]interface{}{"max_tokens": in.MaxTokens}
	if in.Temperature != nil {
		opts["temperature"] = *in.Temperature
	}
	if in.TopP != nil {
		opts["top_p"] = *in.TopP
	}
	if in.TopK != nil {
		opts["top_k"] = *in.TopK
	}
	if len(in.StopSequences) > 0 {
		opts["stop"] = in.StopSequences
	}
	optBytes, _ := json.Marshal(opts)

	req := APIProxyRequest{
		Provider: provider,
		Model:    modelName,
		Messages: msgs,
		Stream:   in.Stream,
		Options:  optBytes,
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

func anthropicToChatMessages(system json.RawMessage, messagesJSON json.RawMessage) ([]ChatMessage, error) {
	var out []ChatMessage
	if len(system) > 0 {
		s, err := anthropicContentToString(system)
		if err != nil {
			return nil, fmt.Errorf("invalid system: %w", err)
		}
		if strings.TrimSpace(s) != "" {
			out = append(out, ChatMessage{Role: "system", Content: s})
		}
	}
	var rawMsgs []json.RawMessage
	if err := json.Unmarshal(messagesJSON, &rawMsgs); err != nil {
		return nil, fmt.Errorf("messages must be an array")
	}
	for _, rm := range rawMsgs {
		var wrap map[string]json.RawMessage
		if err := json.Unmarshal(rm, &wrap); err != nil {
			return nil, fmt.Errorf("invalid message object")
		}
		roleBytes, ok := wrap["role"]
		if !ok {
			return nil, fmt.Errorf("each message must have role")
		}
		var role string
		if err := json.Unmarshal(roleBytes, &role); err != nil {
			return nil, fmt.Errorf("invalid message role")
		}
		role = strings.TrimSpace(strings.ToLower(role))
		if role != "user" && role != "assistant" {
			return nil, fmt.Errorf("unsupported message role %q (only user, assistant)", role)
		}
		contentRaw, ok := wrap["content"]
		if !ok {
			return nil, fmt.Errorf("each message must have content")
		}
		text, err := anthropicContentToString(contentRaw)
		if err != nil {
			return nil, err
		}
		out = append(out, ChatMessage{Role: role, Content: text})
	}
	return out, nil
}

func anthropicContentToString(raw json.RawMessage) (string, error) {
	raw = bytesTrimSpaceJSON(raw)
	if len(raw) == 0 {
		return "", nil
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", fmt.Errorf("invalid string content")
		}
		return s, nil
	}
	if raw[0] == '[' {
		var blocks []map[string]interface{}
		if err := json.Unmarshal(raw, &blocks); err != nil {
			return "", fmt.Errorf("invalid content blocks")
		}
		var b strings.Builder
		for _, blk := range blocks {
			t, _ := blk["type"].(string)
			if t == "text" {
				if tx, ok := blk["text"].(string); ok {
					b.WriteString(tx)
				}
			}
		}
		return b.String(), nil
	}
	return "", fmt.Errorf("unsupported content shape")
}

func bytesTrimSpaceJSON(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
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
