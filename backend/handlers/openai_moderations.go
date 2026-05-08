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

type ModerationInput struct {
	Raw json.RawMessage
}

func (mi *ModerationInput) UnmarshalJSON(data []byte) error {
	mi.Raw = make(json.RawMessage, len(data))
	copy(mi.Raw, data)
	return nil
}

func (mi *ModerationInput) MarshalJSON() ([]byte, error) {
	return mi.Raw, nil
}

type ModerationsRequest struct {
	Model string          `json:"model,omitempty"`
	Input ModerationInput `json:"input"`
}

type ModerationsResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

type ModerationResult struct {
	Flagged                   bool                `json:"flagged"`
	Categories                map[string]bool     `json:"categories"`
	CategoryScores            map[string]float64  `json:"category_scores"`
	CategoryAppliedInputTypes map[string][]string `json:"category_applied_input_types,omitempty"`
	Index                     int                 `json:"index,omitempty"`
}

func OpenAIModerations(c *gin.Context) {
	startTime := time.Now()
	requestID := "mod_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	var req ModerationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if len(req.Input.Raw) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input is required"})
		return
	}

	if req.Model == "" {
		req.Model = "omni-moderation-latest"
	}

	billingEngine := billing.GetBillingEngine()
	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeModerations,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitRequest,
		Quantity:     1,
		RequestID:    requestID,
		Reason:       "Moderations pre-deduct",
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

	upstreamURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeModerations)
	if upstreamURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve endpoint URL"})
		return
	}

	reqBody := map[string]interface{}{
		"model": modelName,
		"input": req.Input.Raw,
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

	client := &http.Client{Timeout: 30 * time.Second}
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

	billingEngine.SettlePreDeductV2(userIDInt, requestID, 1, services.EndpointTypeModerations, providerCfg.Code, billing.BillingUnitRequest)

	logBillingUsage(userIDInt, requestID, services.EndpointTypeModerations, modelName, int(time.Since(startTime).Milliseconds()), 1)

	c.Data(http.StatusOK, "application/json", respBody)
}
