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
// Supports stream:true (SSE) for OpenAI-compatible providers; see deploy/litellm/README.md.
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

	db := config.GetDB()
	provider, modelName := services.ResolveOpenAICompatModel(db, modelVal)
	if provider == "" || modelName == "" {
		respondOpenAIError(c, http.StatusBadRequest, "Could not resolve provider from model")
		return
	}

	req := APIProxyRequest{
		Provider: provider,
		Model:    modelName,
		Messages: messages,
		Stream:   stream,
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
