package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImagesGenerationsRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Style          string `json:"style,omitempty"`
	User           string `json:"user,omitempty"`
}

type ImagesVariationsRequest struct {
	Model          string `json:"model,omitempty"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

type ImagesEditsRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

type ImagesResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Data    []ImageData `json:"data"`
	Created int64       `json:"created"`
}

type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

var imageSizePricing = map[string]float64{
	"256x256":   5.0,
	"512x512":   10.0,
	"1024x1024": 20.0,
	"1024x1792": 30.0,
	"1792x1024": 30.0,
	"1792x1792": 40.0,
}

const defaultImagePrice = 20.0
const defaultImageSize = "1024x1024"

func getImagePrice(size string) float64 {
	if price, ok := imageSizePricing[size]; ok {
		return price
	}
	return defaultImagePrice
}

func OpenAIImagesGenerations(c *gin.Context) {
	startTime := time.Now()
	requestID := "img_gen_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	var req ImagesGenerationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.N <= 0 {
		req.N = 1
	}
	if req.Size == "" {
		req.Size = defaultImageSize
	}

	pricePerImage := getImagePrice(req.Size)
	totalPrice := pricePerImage * float64(req.N)

	billingEngine := billing.GetBillingEngine()
	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeImagesGenerations,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitImage,
		Quantity:     int64(totalPrice / getImagePrice(defaultImageSize)),
		RequestID:    requestID,
		Reason:       "Images generations pre-deduct",
	}
	if preDeductErr := billingEngine.PreDeductBalanceV2(preDeductReq); preDeductErr != nil {
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

	upstreamURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeImagesGenerations)
	if upstreamURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve endpoint URL"})
		return
	}

	reqBody := map[string]interface{}{
		"model":           modelName,
		"prompt":          req.Prompt,
		"n":               req.N,
		"size":            req.Size,
		"response_format": req.ResponseFormat,
	}
	if req.Quality != "" {
		reqBody["quality"] = req.Quality
	}
	if req.Style != "" {
		reqBody["style"] = req.Style
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

	client := &http.Client{Timeout: 120 * time.Second}
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

	actualQuantity := int64(float64(req.N) * pricePerImage / getImagePrice(defaultImageSize))
	billingEngine.SettlePreDeductV2(userIDInt, requestID, actualQuantity, services.EndpointTypeImagesGenerations, providerCfg.Code, billing.BillingUnitImage)

	logBillingUsage(userIDInt, requestID, services.EndpointTypeImagesGenerations, modelName, int(time.Since(startTime).Milliseconds()), actualQuantity)
	c.Data(http.StatusOK, "application/json", respBody)
}

func OpenAIImagesVariations(c *gin.Context) {
	startTime := time.Now()
	requestID := "img_var_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	imageData, _, err := services.ParseFileFieldWithMIMEValidation(c, "image", "image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	n := services.ParseFormInt(c, "n", 1)
	size := c.PostForm("size")
	if size == "" {
		size = defaultImageSize
	}
	responseFormat := c.PostForm("response_format")
	model := c.PostForm("model")

	pricePerImage := getImagePrice(size)
	totalPrice := pricePerImage * float64(n)

	billingEngine := billing.GetBillingEngine()
	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeImagesVariations,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitImage,
		Quantity:     int64(totalPrice / getImagePrice(defaultImageSize)),
		RequestID:    requestID,
		Reason:       "Images variations pre-deduct",
	}
	if preDeductErr := billingEngine.PreDeductBalanceV2(preDeductReq); preDeductErr != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
		return
	}

	db := config.GetDB()

	if model == "" {
		model = "dall-e-2"
	}
	providerCfg, decryptedKey, modelName, err := resolveResponseProvider(c, db, model, userIDInt)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	upstreamURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeImagesVariations)
	if upstreamURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve endpoint URL"})
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("image", "image.png")
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create multipart form"})
		return
	}
	part.Write(imageData)

	writer.WriteField("model", modelName)
	writer.WriteField("n", fmt.Sprintf("%d", n))
	if size != "" {
		writer.WriteField("size", size)
	}
	if responseFormat != "" {
		writer.WriteField("response_format", responseFormat)
	}
	writer.Close()

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, &buf)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
		return
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+decryptedKey)

	client := &http.Client{Timeout: 120 * time.Second}
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

	actualQuantity := int64(float64(n) * pricePerImage / getImagePrice(defaultImageSize))
	billingEngine.SettlePreDeductV2(userIDInt, requestID, actualQuantity, services.EndpointTypeImagesVariations, providerCfg.Code, billing.BillingUnitImage)

	logBillingUsage(userIDInt, requestID, services.EndpointTypeImagesVariations, modelName, int(time.Since(startTime).Milliseconds()), actualQuantity)

	c.Data(http.StatusOK, "application/json", respBody)
}

func OpenAIImagesEdits(c *gin.Context) {
	startTime := time.Now()
	requestID := "img_edit_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	imageData, _, err := services.ParseFileFieldWithMIMEValidation(c, "image", "image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prompt := c.PostForm("prompt")
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	n := services.ParseFormInt(c, "n", 1)
	size := c.PostForm("size")
	if size == "" {
		size = defaultImageSize
	}
	responseFormat := c.PostForm("response_format")
	model := c.PostForm("model")

	pricePerImage := getImagePrice(size)
	totalPrice := pricePerImage * float64(n)

	billingEngine := billing.GetBillingEngine()
	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeImagesEdits,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitImage,
		Quantity:     int64(totalPrice / getImagePrice(defaultImageSize)),
		RequestID:    requestID,
		Reason:       "Images edits pre-deduct",
	}
	if preDeductErr := billingEngine.PreDeductBalanceV2(preDeductReq); preDeductErr != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
		return
	}

	db := config.GetDB()

	if model == "" {
		model = "dall-e-2"
	}
	providerCfg, decryptedKey, modelName, err := resolveResponseProvider(c, db, model, userIDInt)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	upstreamURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeImagesEdits)
	if upstreamURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve endpoint URL"})
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("image", "image.png")
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create multipart form"})
		return
	}
	part.Write(imageData)

	maskData, maskFilename, maskErr := services.ParseFileField(c, "mask")
	if maskErr == nil && maskData != nil {
		maskPart, _ := writer.CreateFormFile("mask", maskFilename)
		maskPart.Write(maskData)
	}

	writer.WriteField("model", modelName)
	writer.WriteField("prompt", prompt)
	writer.WriteField("n", fmt.Sprintf("%d", n))
	if size != "" {
		writer.WriteField("size", size)
	}
	if responseFormat != "" {
		writer.WriteField("response_format", responseFormat)
	}
	writer.Close()

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, &buf)
	if err != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
		return
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+decryptedKey)

	client := &http.Client{Timeout: 120 * time.Second}
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

	actualQuantity := int64(float64(n) * pricePerImage / getImagePrice(defaultImageSize))
	billingEngine.SettlePreDeductV2(userIDInt, requestID, actualQuantity, services.EndpointTypeImagesEdits, providerCfg.Code, billing.BillingUnitImage)

	logBillingUsage(userIDInt, requestID, services.EndpointTypeImagesEdits, modelName, int(time.Since(startTime).Milliseconds()), actualQuantity)

	c.Data(http.StatusOK, "application/json", respBody)
}

func logBillingUsage(userID int, requestID, endpointType, model string, latencyMs int, quantity int64) {
	_ = userID
	_ = requestID
	_ = endpointType
	_ = model
	_ = latencyMs
	_ = quantity
}
