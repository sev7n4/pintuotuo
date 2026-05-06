package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ResponseInput struct {
	Text  string `json:"-"`
	Parts []ResponseInputMessagePart `json:"-"`
}

type ResponseInputMessagePart struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

func (ri *ResponseInput) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		ri.Text = str
		ri.Parts = nil
		return nil
	}
	var parts []ResponseInputMessagePart
	if err := json.Unmarshal(data, &parts); err == nil {
		ri.Text = ""
		ri.Parts = parts
		return nil
	}
	return fmt.Errorf("input must be a string or an array of messages")
}

func (ri ResponseInput) MarshalJSON() ([]byte, error) {
	if ri.Parts != nil {
		return json.Marshal(ri.Parts)
	}
	return json.Marshal(ri.Text)
}

type ToolConfig struct {
	Type        string          `json:"type" binding:"required"`
	Name        string          `json:"name,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
	ServerURL   string          `json:"server_url,omitempty"`
	ServerLabel string          `json:"server_label,omitempty"`
}

type ResponseRequest struct {
	Model              string        `json:"model" binding:"required"`
	Input              ResponseInput `json:"input" binding:"required"`
	Instructions       string        `json:"instructions,omitempty"`
	PreviousResponseID string        `json:"previous_response_id,omitempty"`
	Tools              []ToolConfig  `json:"tools,omitempty"`
	ToolChoice         interface{}   `json:"tool_choice,omitempty"`
	Temperature        *float64      `json:"temperature,omitempty"`
	MaxOutputTokens    *int          `json:"max_output_tokens,omitempty"`
	Metadata           interface{}   `json:"metadata,omitempty"`
	Stream             bool          `json:"stream,omitempty"`
	Background         bool          `json:"background,omitempty"`
	Truncate           *int          `json:"truncate,omitempty"`
	Reasoning          interface{}   `json:"reasoning,omitempty"`
}

type ReasoningOutput struct {
	Type             string          `json:"type"`
	Summary          json.RawMessage `json:"summary,omitempty"`
	EncryptedContent string          `json:"encrypted_content,omitempty"`
}

type ResponseOutputItem struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Status  string          `json:"status,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
	Name    string          `json:"name,omitempty"`
	CallID  string          `json:"call_id,omitempty"`
}

type ResponseUsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type ResponseAPIResponse struct {
	ID        string               `json:"id"`
	Object    string               `json:"object"`
	Model     string               `json:"model"`
	Output    []ResponseOutputItem `json:"output"`
	Usage     ResponseUsageInfo    `json:"usage"`
	Status    string               `json:"status"`
	CreatedAt int64                `json:"created_at"`
	Error     interface{}          `json:"error,omitempty"`
	Metadata  interface{}          `json:"metadata,omitempty"`
}

type BackgroundJobResponse struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Status string `json:"status"`
	Model  string `json:"model"`
}

type BackgroundStatusResponse struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Status     string `json:"status"`
	ResponseID string `json:"response_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

func OpenAIResponses(c *gin.Context) {
	startTime := time.Now()
	requestID := "resp_req_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	var req ResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	db := config.GetDB()
	billingEngine := billing.GetBillingEngine()
	storageSvc := services.NewResponseStorageService(db)

	estimatedTokens := estimateResponseTokens(&req)
	merchantID := getMerchantIDFromContext(c)

	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeResponses,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitToken,
		Quantity:     int64(estimatedTokens),
		RequestID:    requestID,
		Reason:       "Response API pre-deduct",
	}
	if err := billingEngine.PreDeductBalanceV2(preDeductReq); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
		return
	}

	providerCfg, decryptedKey, billModel, err := resolveResponseProvider(c, db, req.Model, userIDInt)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var previousOutput []byte
	if req.PreviousResponseID != "" {
		prevResp, err := storageSvc.GetResponse(c.Request.Context(), req.PreviousResponseID)
		if err != nil || prevResp == nil {
			billingEngine.CancelPreDeduct(userIDInt, requestID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "previous_response_id not found"})
			return
		}
		previousOutput = prevResp.Output
	}

	upstreamBody := buildUpstreamRequestBody(&req, previousOutput)

	if req.Background {
		handleBackgroundRequest(c, db, billingEngine, storageSvc, userIDInt, merchantID, requestID, req, providerCfg, decryptedKey, billModel, upstreamBody, startTime)
		return
	}

	if req.Stream {
		handleStreamResponse(c, db, billingEngine, storageSvc, userIDInt, merchantID, requestID, req, providerCfg, decryptedKey, billModel, upstreamBody, startTime)
		return
	}

	handleSyncResponse(c, db, billingEngine, storageSvc, userIDInt, merchantID, requestID, req, providerCfg, decryptedKey, billModel, upstreamBody, startTime)
}

func OpenAIResponsesGet(c *gin.Context) {
	responseID := c.Param("id")
	db := config.GetDB()
	storageSvc := services.NewResponseStorageService(db)

	resp, err := storageSvc.GetResponse(c.Request.Context(), responseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "response not found"})
		return
	}

	var output []ResponseOutputItem
	if resp.Output != nil {
		json.Unmarshal(resp.Output, &output)
	}
	var usage ResponseUsageInfo
	if resp.Usage != nil {
		json.Unmarshal(resp.Usage, &usage)
	}

	c.JSON(http.StatusOK, ResponseAPIResponse{
		ID:        resp.ResponseID,
		Object:    "response",
		Model:     resp.Model,
		Output:    output,
		Usage:     usage,
		Status:    resp.Status,
		CreatedAt: resp.CreatedAt.Unix(),
	})
}

func OpenAIResponsesDelete(c *gin.Context) {
	responseID := c.Param("id")
	db := config.GetDB()
	storageSvc := services.NewResponseStorageService(db)

	err := storageSvc.DeleteResponse(c.Request.Context(), responseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      responseID,
		"object":  "response",
		"deleted": true,
	})
}

func OpenAIResponsesStatus(c *gin.Context) {
	responseID := c.Param("id")
	db := config.GetDB()
	storageSvc := services.NewResponseStorageService(db)

	resp, err := storageSvc.GetResponse(c.Request.Context(), responseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "response not found"})
		return
	}

	statusResp := BackgroundStatusResponse{
		ID:     resp.ResponseID,
		Object: "response",
		Status: resp.Status,
	}
	if resp.Status == "completed" {
		statusResp.ResponseID = resp.ResponseID
	}

	c.JSON(http.StatusOK, statusResp)
}

func handleSyncResponse(c *gin.Context, db *sql.DB, billingEngine *billing.BillingEngine, storageSvc *services.ResponseStorageService, userID int, merchantID int, requestID string, req ResponseRequest, providerCfg *services.ExecutionProviderConfig, decryptedKey string, billModel string, upstreamBody map[string]interface{}, startTime time.Time) {
	endpointURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeResponses)

	upstreamJSON, _ := json.Marshal(upstreamBody)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", endpointURL, strings.NewReader(string(upstreamJSON)))
	if err != nil {
		billingEngine.CancelPreDeduct(userID, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+decryptedKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		billingEngine.CancelPreDeduct(userID, requestID)
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream error: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		billingEngine.CancelPreDeduct(userID, requestID)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read upstream response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		billingEngine.CancelPreDeduct(userID, requestID)
		c.JSON(resp.StatusCode, json.RawMessage(body))
		return
	}

	var apiResp ResponseAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		billingEngine.CancelPreDeduct(userID, requestID)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to parse upstream response"})
		return
	}

	actualTokens := apiResp.Usage.TotalTokens
	if actualTokens == 0 {
		actualTokens = apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens
	}
	billingEngine.SettlePreDeductV2(userID, requestID, int64(actualTokens), services.EndpointTypeResponses, "openai")

	toolBilling := calculateToolBilling(apiResp.Output)
	if toolBilling > 0 {
		toolReq := billing.BillingRequest{
			UserID:       userID,
			EndpointType: services.EndpointTypeResponses,
			ProviderCode: "openai",
			UnitType:     billing.BillingUnitRequest,
			Quantity:     int64(toolBilling),
			RequestID:    requestID + "_tools",
			Reason:       "Response API tool calls",
		}
		billingEngine.PreDeductBalanceV2(toolReq)
		billingEngine.SettlePreDeductV2(userID, requestID+"_tools", int64(toolBilling), services.EndpointTypeResponses, "openai")
	}

	inputJSON, _ := json.Marshal(req.Input)
	outputJSON, _ := json.Marshal(apiResp.Output)
	usageJSON, _ := json.Marshal(apiResp.Usage)

	storedResp := &models.StoredResponse{
		ResponseID: apiResp.ID,
		UserID:     userID,
		MerchantID: merchantID,
		Model:      req.Model,
		Input:      inputJSON,
		Output:     outputJSON,
		Usage:      usageJSON,
		Status:     "completed",
	}
	if storeErr := storageSvc.StoreResponse(c.Request.Context(), storedResp); storeErr != nil {
		log.Printf("failed to store response %s: %v", apiResp.ID, storeErr)
	}

	c.JSON(http.StatusOK, apiResp)
}

func handleStreamResponse(c *gin.Context, db *sql.DB, billingEngine *billing.BillingEngine, storageSvc *services.ResponseStorageService, userID int, merchantID int, requestID string, req ResponseRequest, providerCfg *services.ExecutionProviderConfig, decryptedKey string, billModel string, upstreamBody map[string]interface{}, startTime time.Time) {
	upstreamBody["stream"] = true
	endpointURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeResponses)

	upstreamJSON, _ := json.Marshal(upstreamBody)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", endpointURL, strings.NewReader(string(upstreamJSON)))
	if err != nil {
		billingEngine.CancelPreDeduct(userID, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+decryptedKey)

	client := &http.Client{Timeout: 15 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		billingEngine.CancelPreDeduct(userID, requestID)
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream error: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		billingEngine.CancelPreDeduct(userID, requestID)
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, json.RawMessage(body))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var totalOutputTokens int
	var responseID string
	var outputItems []ResponseOutputItem

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
			c.Writer.(http.Flusher).Flush()
			break
		}

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if id, ok := event["response_id"].(string); ok && id != "" {
			responseID = id
		}

		if eventType, ok := event["type"].(string); ok {
			if strings.Contains(eventType, "delta") || strings.Contains(eventType, "text") {
				totalOutputTokens++
			}
			if eventType == "response.completed" {
				if respObj, ok := event["response"].(map[string]interface{}); ok {
					if usage, ok := respObj["usage"].(map[string]interface{}); ok {
						if ot, ok := usage["output_tokens"].(float64); ok {
							totalOutputTokens = int(ot)
						}
					}
					if id, ok := respObj["id"].(string); ok {
						responseID = id
					}
					if output, ok := respObj["output"].([]interface{}); ok {
						outputJSON, _ := json.Marshal(output)
						json.Unmarshal(outputJSON, &outputItems)
					}
				}
			}
		}

		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		c.Writer.(http.Flusher).Flush()
	}

	estimatedInputTokens := estimateResponseTokens(&req)
	totalTokens := estimatedInputTokens + totalOutputTokens
	billingEngine.SettlePreDeductV2(userID, requestID, int64(totalTokens), services.EndpointTypeResponses, "openai")

	if responseID != "" {
		inputJSON, _ := json.Marshal(req.Input)
		outputJSON, _ := json.Marshal(outputItems)
		usageJSON, _ := json.Marshal(ResponseUsageInfo{
			InputTokens:  estimatedInputTokens,
			OutputTokens: totalOutputTokens,
			TotalTokens:  totalTokens,
		})
		storedResp := &models.StoredResponse{
			ResponseID: responseID,
			UserID:     userID,
			MerchantID: merchantID,
			Model:      req.Model,
			Input:      inputJSON,
			Output:     outputJSON,
			Usage:      usageJSON,
			Status:     "completed",
		}
		if storeErr := storageSvc.StoreResponse(c.Request.Context(), storedResp); storeErr != nil {
			log.Printf("failed to store stream response %s: %v", responseID, storeErr)
		}
	}
}

func handleBackgroundRequest(c *gin.Context, db *sql.DB, billingEngine *billing.BillingEngine, storageSvc *services.ResponseStorageService, userID int, merchantID int, requestID string, req ResponseRequest, providerCfg *services.ExecutionProviderConfig, decryptedKey string, billModel string, upstreamBody map[string]interface{}, startTime time.Time) {
	jobID := "job_" + uuid.New().String()[:24]
	responseID := "resp_" + uuid.New().String()[:24]

	inputJSON, _ := json.Marshal(req.Input)
	storedResp := &models.StoredResponse{
		ResponseID:      responseID,
		UserID:          userID,
		MerchantID:      merchantID,
		Model:           req.Model,
		Input:           inputJSON,
		Status:          "queued",
		BackgroundJobID: sql.NullString{String: jobID, Valid: true},
	}
	if err := storageSvc.StoreResponse(c.Request.Context(), storedResp); err != nil {
		billingEngine.CancelPreDeduct(userID, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create background job"})
		return
	}

	go func() {
		ctx := context.Background()
		storageSvc.UpdateStatus(ctx, responseID, "running")

		endpointURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeResponses)
		upstreamJSON, _ := json.Marshal(upstreamBody)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", endpointURL, strings.NewReader(string(upstreamJSON)))
		if err != nil {
			storageSvc.UpdateStatus(ctx, responseID, "failed")
			billingEngine.CancelPreDeduct(userID, requestID)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+decryptedKey)

		client := &http.Client{Timeout: 300 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			storageSvc.UpdateStatus(ctx, responseID, "failed")
			billingEngine.CancelPreDeduct(userID, requestID)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil || resp.StatusCode != http.StatusOK {
			storageSvc.UpdateStatus(ctx, responseID, "failed")
			billingEngine.CancelPreDeduct(userID, requestID)
			return
		}

		var apiResp ResponseAPIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			storageSvc.UpdateStatus(ctx, responseID, "failed")
			billingEngine.CancelPreDeduct(userID, requestID)
			return
		}

		actualTokens := apiResp.Usage.TotalTokens
		if actualTokens == 0 {
			actualTokens = apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens
		}
		billingEngine.SettlePreDeductV2(userID, requestID, int64(actualTokens), services.EndpointTypeResponses, "openai")

		outputJSON, _ := json.Marshal(apiResp.Output)
		usageJSON, _ := json.Marshal(apiResp.Usage)
		storageSvc.UpdateOutput(ctx, responseID, outputJSON, usageJSON)
	}()

	c.JSON(http.StatusAccepted, BackgroundJobResponse{
		ID:     responseID,
		Object: "response",
		Status: "queued",
		Model:  req.Model,
	})
}

func buildUpstreamRequestBody(req *ResponseRequest, previousOutput []byte) map[string]interface{} {
	body := map[string]interface{}{
		"model": req.Model,
	}

	if req.Input.Text != "" {
		body["input"] = req.Input.Text
	} else if req.Input.Parts != nil {
		body["input"] = req.Input.Parts
	}

	if req.Instructions != "" {
		body["instructions"] = req.Instructions
	}
	if req.PreviousResponseID != "" {
		body["previous_response_id"] = req.PreviousResponseID
	}
	if len(req.Tools) > 0 {
		body["tools"] = req.Tools
	}
	if req.ToolChoice != nil {
		body["tool_choice"] = req.ToolChoice
	}
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.MaxOutputTokens != nil {
		body["max_output_tokens"] = *req.MaxOutputTokens
	}
	if req.Metadata != nil {
		body["metadata"] = req.Metadata
	}
	if req.Stream {
		body["stream"] = true
	}
	if req.Reasoning != nil {
		body["reasoning"] = req.Reasoning
	}
	if req.Truncate != nil {
		body["truncate"] = *req.Truncate
	}

	return body
}

func estimateResponseTokens(req *ResponseRequest) int {
	tokens := 0
	if req.Input.Text != "" {
		tokens += len(req.Input.Text) / 4
	}
	for _, part := range req.Input.Parts {
		tokens += len(part.Content) / 4
	}
	if req.Instructions != "" {
		tokens += len(req.Instructions) / 4
	}
	if tokens == 0 {
		tokens = 100
	}
	return tokens
}

func calculateToolBilling(output []ResponseOutputItem) int {
	toolBillingMap := map[string]billing.BillingUnit{
		"web_search_call":  billing.BillingUnitRequest,
		"file_search_call": billing.BillingUnitRequest,
		"computer_call":    billing.BillingUnitRequest,
		"code_interpreter": billing.BillingUnitRequest,
		"image_generation": billing.BillingUnitImage,
		"mcp_call":         billing.BillingUnitRequest,
		"function_call":    billing.BillingUnitToken,
	}

	count := 0
	for _, item := range output {
		if _, ok := toolBillingMap[item.Type]; ok {
			if toolBillingMap[item.Type] != billing.BillingUnitToken {
				count++
			}
		}
	}
	return count
}

func resolveResponseProvider(c *gin.Context, db *sql.DB, model string, userID int) (*services.ExecutionProviderConfig, string, string, error) {
	provider, modelName := services.ResolveOpenAICompatModel(db, model)
	if provider == "" || modelName == "" {
		return nil, "", "", fmt.Errorf("could not resolve provider from model")
	}

	cfg, err := getProviderRuntimeConfig(db, provider)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get provider config: %w", err)
	}

	execCfg := &services.ExecutionProviderConfig{
		Code:           cfg.Code,
		Name:           cfg.Name,
		APIBaseURL:     cfg.APIBaseURL,
		APIFormat:      cfg.APIFormat,
		ProviderRegion: cfg.ProviderRegion,
		RouteStrategy:  cfg.RouteStrategy,
		Endpoints:      cfg.Endpoints,
	}
	execCfg.GatewayMode = services.ResolveRouteModeWithProvider("", cfg.ProviderRegion)

	endpointURL := services.ResolveEndpoint(execCfg)
	if endpointURL != "" {
		execCfg.APIBaseURL = endpointURL
	}

	decryptedKey, err := getDecryptedAPIKeyForProvider(db, provider, userID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get API key: %w", err)
	}

	return execCfg, decryptedKey, modelName, nil
}

func getDecryptedAPIKeyForProvider(db *sql.DB, provider string, userID int) (string, error) {
	var encryptedKey string
	err := db.QueryRow(
		`SELECT ak.encrypted_key FROM merchant_api_keys ak
		 JOIN model_providers mp ON ak.provider_id = mp.id
		 WHERE mp.code = $1 AND ak.status = 'active'
		 ORDER BY ak.is_default DESC LIMIT 1`,
		provider,
	).Scan(&encryptedKey)
	if err != nil {
		return "", fmt.Errorf("no API key found for provider %s: %w", provider, err)
	}
	decrypted, err := utils.Decrypt(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt API key: %w", err)
	}
	return decrypted, nil
}

func getMerchantIDFromContext(c *gin.Context) int {
	if merchantID, exists := c.Get("merchant_id"); exists {
		if id, ok := merchantID.(int); ok {
			return id
		}
		if id, ok := merchantID.(string); ok {
			if intID, err := strconv.Atoi(id); err == nil {
				return intID
			}
		}
	}
	return 0
}
