package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

// resolveOpenAICompatModel maps an OpenAI-style model field to provider + upstream model name.
// If model contains a slash (e.g. "openai/gpt-4o"), the prefix is the provider code.
// Otherwise provider is inferred from the model id (e.g. claude-* -> anthropic).
func resolveOpenAICompatModel(model string) (provider string, modelName string) {
	model = strings.TrimSpace(model)
	if model == "" {
		return "", ""
	}
	if idx := strings.Index(model, "/"); idx > 0 {
		return strings.ToLower(strings.TrimSpace(model[:idx])), strings.TrimSpace(model[idx+1:])
	}
	m := strings.ToLower(model)
	switch {
	case strings.HasPrefix(m, "claude"):
		return "anthropic", model
	case strings.HasPrefix(m, "gemini"):
		return "google", model
	case strings.HasPrefix(m, "glm-") || strings.HasPrefix(m, "chatglm") || strings.HasPrefix(m, "cog-"):
		return "zhipu", model
	default:
		return "openai", model
	}
}

func respondOpenAIError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "invalid_request_error",
		},
	})
}

// OpenAIChatCompletions accepts an OpenAI-compatible JSON body and reuses the platform proxy pipeline.
// Clients should set base URL to {API_ORIGIN}/api/v1/openai/v1 (OpenAI SDK: baseURL + "/chat/completions").
// Authentication: Bearer platform API key (ptd_* / ptt_*) or existing JWT (same as /proxy/chat).
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
	msgsRaw, ok := raw["messages"]
	if !ok {
		respondOpenAIError(c, http.StatusBadRequest, "messages is required")
		return
	}
	messagesJSON, marshalErr := json.Marshal(msgsRaw)
	if marshalErr != nil {
		respondOpenAIError(c, http.StatusBadRequest, "messages is invalid")
		return
	}
	var messages []ChatMessage
	if msgErr := json.Unmarshal(messagesJSON, &messages); msgErr != nil || len(messages) == 0 {
		respondOpenAIError(c, http.StatusBadRequest, "messages must be a non-empty array")
		return
	}

	stream := false
	if streamVal, streamOk := raw["stream"].(bool); streamOk {
		stream = streamVal
	}
	if stream {
		respondOpenAIError(c, http.StatusBadRequest, "Streaming is not supported yet; set stream to false")
		return
	}

	delete(raw, "model")
	delete(raw, "messages")
	delete(raw, "stream")
	var options json.RawMessage
	if len(raw) > 0 {
		optBytes, optErr := json.Marshal(raw)
		if optErr != nil {
			respondOpenAIError(c, http.StatusBadRequest, "Invalid optional parameters")
			return
		}
		options = optBytes
	}

	provider, modelName := resolveOpenAICompatModel(modelVal)
	if provider == "" || modelName == "" {
		respondOpenAIError(c, http.StatusBadRequest, "Could not resolve provider from model")
		return
	}

	req := APIProxyRequest{
		Provider: provider,
		Model:    modelName,
		Messages: messages,
		Stream:   false,
		Options:  options,
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
