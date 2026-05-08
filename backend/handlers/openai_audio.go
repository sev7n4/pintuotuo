package handlers

import (
	"bytes"
	"encoding/json"
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

type AudioSpeechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

type AudioTranscriptionRequest struct {
	Model                  string   `json:"model"`
	Language               string   `json:"language,omitempty"`
	Prompt                 string   `json:"prompt,omitempty"`
	ResponseFormat         string   `json:"response_format,omitempty"`
	Temperature            float64  `json:"temperature,omitempty"`
	TimestampGranularities []string `json:"timestamp_granularities,omitempty"`
}

type AudioTranslationRequest struct {
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
}

type AudioTranscriptionResponse struct {
	Text     string         `json:"text"`
	Language string         `json:"language,omitempty"`
	Duration float64        `json:"duration,omitempty"`
	Segments []AudioSegment `json:"segments,omitempty"`
	Words    []AudioWord    `json:"words,omitempty"`
}

type AudioTranslationResponse struct {
	Text     string  `json:"text"`
	Language string  `json:"language,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

type AudioSegment struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Tokens           []int   `json:"tokens"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
}

type AudioWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

var audioFormatContentType = map[string]string{
	"mp3":  "audio/mpeg",
	"opus": "audio/opus",
	"aac":  "audio/aac",
	"flac": "audio/flac",
	"wav":  "audio/wav",
	"pcm":  "audio/pcm",
}

func getAudioContentType(format string) string {
	if ct, ok := audioFormatContentType[format]; ok {
		return ct
	}
	return "audio/mpeg"
}

func OpenAIAudioSpeech(c *gin.Context) {
	startTime := time.Now()
	requestID := "audio_speech_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	var req AudioSpeechRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.Input == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "input is required"})
		return
	}
	if req.Voice == "" {
		req.Voice = "alloy"
	}

	charCount := int64(len(req.Input))
	if charCount < 1 {
		charCount = 1
	}

	billingEngine := billing.GetBillingEngine()
	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeAudioSpeech,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitCharacter,
		Quantity:     charCount,
		RequestID:    requestID,
		Reason:       "Audio speech pre-deduct",
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

	upstreamURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeAudioSpeech)
	if upstreamURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve endpoint URL"})
		return
	}

	reqBody := map[string]interface{}{
		"model": modelName,
		"input": req.Input,
		"voice": req.Voice,
	}
	if req.ResponseFormat != "" {
		reqBody["response_format"] = req.ResponseFormat
	}
	if req.Speed > 0 {
		reqBody["speed"] = req.Speed
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

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.Data(resp.StatusCode, "application/json", respBody)
		return
	}

	contentType := getAudioContentType(req.ResponseFormat)
	billingEngine.SettlePreDeductV2(userIDInt, requestID, charCount, services.EndpointTypeAudioSpeech, providerCfg.Code, billing.BillingUnitCharacter)

	logBillingUsage(userIDInt, requestID, services.EndpointTypeAudioSpeech, modelName, int(time.Since(startTime).Milliseconds()), charCount)

	services.StreamBinaryResponse(c, resp, contentType)
}

func OpenAIAudioTranscriptions(c *gin.Context) {
	startTime := time.Now()
	requestID := "audio_trans_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	fileData, _, err := services.ParseFileFieldWithMIMEValidation(c, "file", "audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model := c.PostForm("model")
	if model == "" {
		model = "whisper-1"
	}
	language := c.PostForm("language")
	prompt := c.PostForm("prompt")
	responseFormat := c.PostForm("response_format")
	temperature := c.PostForm("temperature")

	estimatedSeconds := int64(len(fileData) / 32000)
	if estimatedSeconds < 1 {
		estimatedSeconds = 1
	}

	billingEngine := billing.GetBillingEngine()
	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeAudioTranscriptions,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitSecond,
		Quantity:     estimatedSeconds,
		RequestID:    requestID,
		Reason:       "Audio transcriptions pre-deduct",
	}
	if preDeductErr := billingEngine.PreDeductBalanceV2(preDeductReq); preDeductErr != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
		return
	}

	db := config.GetDB()

	providerCfg, decryptedKey, modelName, resolveErr := resolveResponseProvider(c, db, model, userIDInt)
	if resolveErr != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": resolveErr.Error()})
		return
	}

	upstreamURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeAudioTranscriptions)
	if upstreamURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve endpoint URL"})
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, partErr := writer.CreateFormFile("file", "audio.mp3")
	if partErr != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create multipart form"})
		return
	}
	part.Write(fileData)

	writer.WriteField("model", modelName)
	if language != "" {
		writer.WriteField("language", language)
	}
	if prompt != "" {
		writer.WriteField("prompt", prompt)
	}
	if responseFormat != "" {
		writer.WriteField("response_format", responseFormat)
	}
	if temperature != "" {
		writer.WriteField("temperature", temperature)
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

	var transResp AudioTranscriptionResponse
	if jsonErr := json.Unmarshal(respBody, &transResp); jsonErr == nil && transResp.Duration > 0 {
		actualSeconds := int64(transResp.Duration)
		if actualSeconds < 1 {
			actualSeconds = 1
		}
		billingEngine.SettlePreDeductV2(userIDInt, requestID, actualSeconds, services.EndpointTypeAudioTranscriptions, providerCfg.Code, billing.BillingUnitSecond)
		logBillingUsage(userIDInt, requestID, services.EndpointTypeAudioTranscriptions, modelName, int(time.Since(startTime).Milliseconds()), actualSeconds)
	} else {
		billingEngine.SettlePreDeductV2(userIDInt, requestID, estimatedSeconds, services.EndpointTypeAudioTranscriptions, providerCfg.Code, billing.BillingUnitSecond)
		logBillingUsage(userIDInt, requestID, services.EndpointTypeAudioTranscriptions, modelName, int(time.Since(startTime).Milliseconds()), estimatedSeconds)
	}

	c.Data(http.StatusOK, "application/json", respBody)
}

func OpenAIAudioTranslations(c *gin.Context) {
	startTime := time.Now()
	requestID := "audio_transl_" + uuid.New().String()[:24]

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userIDInt, _ := userIDStr.(int)

	fileData, _, err := services.ParseFileFieldWithMIMEValidation(c, "file", "audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model := c.PostForm("model")
	if model == "" {
		model = "whisper-1"
	}
	prompt := c.PostForm("prompt")
	responseFormat := c.PostForm("response_format")
	temperature := c.PostForm("temperature")

	estimatedSeconds := int64(len(fileData) / 32000)
	if estimatedSeconds < 1 {
		estimatedSeconds = 1
	}

	billingEngine := billing.GetBillingEngine()
	preDeductReq := billing.BillingRequest{
		UserID:       userIDInt,
		EndpointType: services.EndpointTypeAudioTranslations,
		ProviderCode: "openai",
		UnitType:     billing.BillingUnitSecond,
		Quantity:     estimatedSeconds,
		RequestID:    requestID,
		Reason:       "Audio translations pre-deduct",
	}
	if preDeductErr := billingEngine.PreDeductBalanceV2(preDeductReq); preDeductErr != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "insufficient balance"})
		return
	}

	db := config.GetDB()

	providerCfg, decryptedKey, modelName, resolveErr := resolveResponseProvider(c, db, model, userIDInt)
	if resolveErr != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": resolveErr.Error()})
		return
	}

	upstreamURL := services.ResolveEndpointByType(providerCfg, services.EndpointTypeAudioTranslations)
	if upstreamURL == "" {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve endpoint URL"})
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, partErr := writer.CreateFormFile("file", "audio.mp3")
	if partErr != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create multipart form"})
		return
	}
	part.Write(fileData)

	writer.WriteField("model", modelName)
	if prompt != "" {
		writer.WriteField("prompt", prompt)
	}
	if responseFormat != "" {
		writer.WriteField("response_format", responseFormat)
	}
	if temperature != "" {
		writer.WriteField("temperature", temperature)
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

	var translResp AudioTranslationResponse
	if jsonErr := json.Unmarshal(respBody, &translResp); jsonErr == nil && translResp.Duration > 0 {
		actualSeconds := int64(translResp.Duration)
		if actualSeconds < 1 {
			actualSeconds = 1
		}
		billingEngine.SettlePreDeductV2(userIDInt, requestID, actualSeconds, services.EndpointTypeAudioTranslations, providerCfg.Code, billing.BillingUnitSecond)
		logBillingUsage(userIDInt, requestID, services.EndpointTypeAudioTranslations, modelName, int(time.Since(startTime).Milliseconds()), actualSeconds)
	} else {
		billingEngine.SettlePreDeductV2(userIDInt, requestID, estimatedSeconds, services.EndpointTypeAudioTranslations, providerCfg.Code, billing.BillingUnitSecond)
		logBillingUsage(userIDInt, requestID, services.EndpointTypeAudioTranslations, modelName, int(time.Since(startTime).Milliseconds()), estimatedSeconds)
	}

	c.Data(http.StatusOK, "application/json", respBody)
}
