package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmbeddingInput struct {
	Raw json.RawMessage
}

func (ei *EmbeddingInput) UnmarshalJSON(data []byte) error {
	ei.Raw = make(json.RawMessage, len(data))
	copy(ei.Raw, data)
	return nil
}

func (ei *EmbeddingInput) MarshalJSON() ([]byte, error) {
	return ei.Raw, nil
}

type EmbeddingsRequest struct {
	Model          string         `json:"model"`
	Input          EmbeddingInput `json:"input"`
	EncodingFormat string         `json:"encoding_format,omitempty"`
	Dimensions     int            `json:"dimensions,omitempty"`
	User           string         `json:"user,omitempty"`
}

type EmbeddingsResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

type EmbeddingData struct {
	Object    string      `json:"object"`
	Embedding interface{} `json:"embedding"`
	Index     int         `json:"index"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func estimateEmbeddingTokens(input json.RawMessage) int64 {
	var parsed interface{}
	if err := json.Unmarshal(input, &parsed); err != nil {
		return int64(len(input) / 4)
	}

	switch v := parsed.(type) {
	case string:
		return int64(len(v) / 4)
	case []interface{}:
		total := 0
		for _, item := range v {
			switch itemVal := item.(type) {
			case string:
				total += len(itemVal) / 4
			case float64:
				total++
			default:
				total += 10
			}
		}
		return int64(total)
	default:
		return int64(len(input) / 4)
	}
}

func OpenAIEmbeddings(c *gin.Context) {
	startTime := time.Now()
	requestID := "embed_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	var req EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if len(req.Input.Raw) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input is required"})
		return
	}

	estimatedTokens := estimateEmbeddingTokens(req.Input.Raw)
	if estimatedTokens < 1 {
		estimatedTokens = 1
	}

	billingEngine := billing.GetBillingEngine()
	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeEmbeddings,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitToken,
		Quantity:     estimatedTokens,
		RequestID:    requestID,
		Reason:       "Embeddings pre-deduct",
	}
	if err := billingEngine.PreDeductBalanceV2(preDeductReq); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
		return
	}

	db := config.GetDB()

	providerCfg, decryptedKey, modelName, err := resolveResponseProvider(c, db, req.Model, userIDInt)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	upstreamURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeEmbeddings)
	if upstreamURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve endpoint URL"})
		return
	}

	reqBody := map[string]interface{}{
		"model": modelName,
		"input": req.Input.Raw,
	}
	if req.EncodingFormat != "" {
		reqBody["encoding_format"] = req.EncodingFormat
	}
	if req.Dimensions > 0 {
		reqBody["dimensions"] = req.Dimensions
	}
	if req.User != "" {
		reqBody["user"] = req.User
	}

	jsonBody, _ := json.Marshal(reqBody)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, bytes.NewReader(jsonBody))
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+decryptedKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream request failed: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read upstream response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.Data(resp.StatusCode, "application/json", respBody)
		return
	}

	var embedResp EmbeddingsResponse
	actualTokens := estimatedTokens
	if jsonErr := json.Unmarshal(respBody, &embedResp); jsonErr == nil && embedResp.Usage.TotalTokens > 0 {
		actualTokens = int64(embedResp.Usage.TotalTokens)
	}
	billingEngine.SettlePreDeductV2(userIDInt, requestID, actualTokens, services.EndpointTypeEmbeddings, providerCfg.Code, billing.BillingUnitToken)

	logBillingUsage(userIDInt, requestID, services.EndpointTypeEmbeddings, modelName, int(time.Since(startTime).Milliseconds()), actualTokens)

	c.Data(http.StatusOK, "application/json", respBody)
}
