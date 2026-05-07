package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseInput_UnmarshalJSON_String(t *testing.T) {
	var input ResponseInput
	err := json.Unmarshal([]byte(`"Hello"`), &input)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", input.Text)
	assert.Nil(t, input.Parts)
}

func TestResponseInput_UnmarshalJSON_Array(t *testing.T) {
	var input ResponseInput
	err := json.Unmarshal([]byte(`[{"role":"user","content":"Hello"}]`), &input)
	assert.NoError(t, err)
	assert.Equal(t, "", input.Text)
	assert.Len(t, input.Parts, 1)
	assert.Equal(t, "user", input.Parts[0].Role)
}

func TestResponseInput_UnmarshalJSON_Invalid(t *testing.T) {
	var input ResponseInput
	err := json.Unmarshal([]byte(`123`), &input)
	assert.Error(t, err)
}

func TestResponseInput_MarshalJSON_String(t *testing.T) {
	input := ResponseInput{Text: "Hello"}
	data, err := json.Marshal(input)
	assert.NoError(t, err)
	assert.Equal(t, `"Hello"`, string(data))
}

func TestResponseInput_MarshalJSON_Array(t *testing.T) {
	input := ResponseInput{
		Parts: []ResponseInputMessagePart{
			{Role: "user", Content: json.RawMessage(`"Hello"`)},
		},
	}
	data, err := json.Marshal(input)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"role"`)
	assert.Contains(t, string(data), `"user"`)
}

func TestResponseRequest_ParseComplete(t *testing.T) {
	raw := `{
		"model": "gpt-4o",
		"input": "Hello",
		"instructions": "Be helpful",
		"previous_response_id": "resp_abc123",
		"stream": false,
		"background": false,
		"temperature": 0.7,
		"max_output_tokens": 1024
	}`
	var req ResponseRequest
	err := json.Unmarshal([]byte(raw), &req)
	assert.NoError(t, err)
	assert.Equal(t, "gpt-4o", req.Model)
	assert.Equal(t, "Hello", req.Input.Text)
	assert.Equal(t, "Be helpful", req.Instructions)
	assert.Equal(t, "resp_abc123", req.PreviousResponseID)
	assert.False(t, req.Stream)
	assert.False(t, req.Background)
	assert.NotNil(t, req.Temperature)
	assert.Equal(t, 0.7, *req.Temperature)
	assert.NotNil(t, req.MaxOutputTokens)
	assert.Equal(t, 1024, *req.MaxOutputTokens)
}

func TestResponseRequest_ParseWithTools(t *testing.T) {
	raw := `{
		"model": "gpt-4o",
		"input": "Search for weather",
		"tools": [
			{"type": "web_search_preview"},
			{"type": "function", "name": "get_weather", "parameters": {"city": "string"}}
		]
	}`
	var req ResponseRequest
	err := json.Unmarshal([]byte(raw), &req)
	assert.NoError(t, err)
	assert.Len(t, req.Tools, 2)
	assert.Equal(t, "web_search_preview", req.Tools[0].Type)
	assert.Equal(t, "function", req.Tools[1].Type)
	assert.Equal(t, "get_weather", req.Tools[1].Name)
}

func TestResponseRequest_ParseWithReasoning(t *testing.T) {
	raw := `{
		"model": "o3-mini",
		"input": "Solve this puzzle",
		"reasoning": {"effort": "high"}
	}`
	var req ResponseRequest
	err := json.Unmarshal([]byte(raw), &req)
	assert.NoError(t, err)
	assert.NotNil(t, req.Reasoning)
}

func TestBuildUpstreamRequestBody(t *testing.T) {
	temp := 0.7
	maxTokens := 1024
	req := &ResponseRequest{
		Model:           "gpt-4o",
		Input:           ResponseInput{Text: "Hello"},
		Instructions:    "Be helpful",
		Temperature:     &temp,
		MaxOutputTokens: &maxTokens,
		Stream:          true,
	}

	body := buildUpstreamRequestBody(req, nil)
	assert.Equal(t, "gpt-4o", body["model"])
	assert.Equal(t, "Hello", body["input"])
	assert.Equal(t, "Be helpful", body["instructions"])
	assert.Equal(t, 0.7, body["temperature"])
	assert.Equal(t, 1024, body["max_output_tokens"])
	assert.Equal(t, true, body["stream"])
}

func TestBuildUpstreamRequestBody_WithPreviousOutput(t *testing.T) {
	req := &ResponseRequest{
		Model:              "gpt-4o",
		Input:              ResponseInput{Text: "Follow up"},
		PreviousResponseID: "resp_abc123",
	}

	previousOutput := []byte(`[{"type":"message","content":"Previous answer"}]`)
	body := buildUpstreamRequestBody(req, previousOutput)
	assert.Equal(t, "Follow up", body["input"])
	assert.Equal(t, "resp_abc123", body["previous_response_id"])
}

func TestEstimateResponseTokens(t *testing.T) {
	req := &ResponseRequest{
		Input:        ResponseInput{Text: "Hello world"},
		Instructions: "Be helpful",
	}
	tokens := estimateResponseTokens(req)
	assert.Greater(t, tokens, 0)
}

func TestEstimateResponseTokens_EmptyInput(t *testing.T) {
	req := &ResponseRequest{
		Input: ResponseInput{},
	}
	tokens := estimateResponseTokens(req)
	assert.Equal(t, 100, tokens)
}

func TestCalculateToolBilling(t *testing.T) {
	output := []ResponseOutputItem{
		{Type: "message", Status: "completed"},
		{Type: "web_search_call", Status: "completed"},
		{Type: "image_generation", Status: "completed"},
		{Type: "function_call", Status: "completed"},
	}
	entries := calculateToolBilling(output)
	assert.Len(t, entries, 3)

	entryMap := map[billing.BillingUnit]int{}
	for _, e := range entries {
		entryMap[e.UnitType] = e.Count
	}
	assert.Equal(t, 1, entryMap[billing.BillingUnitRequest])
	assert.Equal(t, 1, entryMap[billing.BillingUnitImage])
	assert.Equal(t, 1, entryMap[billing.BillingUnitToken])
}

func TestCalculateToolBilling_NoTools(t *testing.T) {
	output := []ResponseOutputItem{
		{Type: "message", Status: "completed"},
	}
	entries := calculateToolBilling(output)
	assert.Len(t, entries, 0)
}

func TestCalculateToolBilling_AllToolTypes(t *testing.T) {
	output := []ResponseOutputItem{
		{Type: "web_search_call", Status: "completed"},
		{Type: "file_search_call", Status: "completed"},
		{Type: "computer_call", Status: "completed"},
		{Type: "code_interpreter", Status: "completed"},
		{Type: "mcp_call", Status: "completed"},
		{Type: "image_generation", Status: "completed"},
		{Type: "function_call", Status: "completed"},
	}
	entries := calculateToolBilling(output)
	assert.Len(t, entries, 3)

	entryMap := map[billing.BillingUnit]int{}
	for _, e := range entries {
		entryMap[e.UnitType] = e.Count
	}
	assert.Equal(t, 5, entryMap[billing.BillingUnitRequest])
	assert.Equal(t, 1, entryMap[billing.BillingUnitImage])
	assert.Equal(t, 1, entryMap[billing.BillingUnitToken])
}

func TestCalculateToolBilling_MultipleSameType(t *testing.T) {
	output := []ResponseOutputItem{
		{Type: "web_search_call", Status: "completed"},
		{Type: "web_search_call", Status: "completed"},
		{Type: "image_generation", Status: "completed"},
		{Type: "image_generation", Status: "completed"},
		{Type: "image_generation", Status: "completed"},
	}
	entries := calculateToolBilling(output)
	assert.Len(t, entries, 2)

	entryMap := map[billing.BillingUnit]int{}
	for _, e := range entries {
		entryMap[e.UnitType] = e.Count
	}
	assert.Equal(t, 2, entryMap[billing.BillingUnitRequest])
	assert.Equal(t, 3, entryMap[billing.BillingUnitImage])
}

func TestReasoningOutput_Parse(t *testing.T) {
	raw := `{
		"type": "reasoning",
		"summary": [{"type":"summary_text","text":"Thinking..."}],
		"encrypted_content": "abc123def456"
	}`
	var reasoning ReasoningOutput
	err := json.Unmarshal([]byte(raw), &reasoning)
	assert.NoError(t, err)
	assert.Equal(t, "reasoning", reasoning.Type)
	assert.Equal(t, "abc123def456", reasoning.EncryptedContent)
}

func TestResponseAPIResponse_Parse(t *testing.T) {
	raw := `{
		"id": "resp_001",
		"object": "response",
		"model": "gpt-4o",
		"output": [
			{"type": "message", "status": "completed", "content": "Hello!"}
		],
		"usage": {"input_tokens": 10, "output_tokens": 5, "total_tokens": 15},
		"status": "completed",
		"created_at": 1234567890
	}`
	var resp ResponseAPIResponse
	err := json.Unmarshal([]byte(raw), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "resp_001", resp.ID)
	assert.Equal(t, "response", resp.Object)
	assert.Equal(t, "gpt-4o", resp.Model)
	assert.Len(t, resp.Output, 1)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
	assert.Equal(t, "completed", resp.Status)
}

func TestBackgroundStatusResponse_WithErrorMessage(t *testing.T) {
	resp := &models.StoredResponse{
		ResponseID:   "resp_001",
		Status:       "failed",
		ErrorMessage: sql.NullString{String: "upstream timeout", Valid: true},
	}

	statusResp := BackgroundStatusResponse{
		ID:     resp.ResponseID,
		Object: "response",
		Status: resp.Status,
	}
	if resp.Status == "completed" {
		statusResp.ResponseID = resp.ResponseID
	}
	if resp.Status == "failed" && resp.ErrorMessage.Valid {
		statusResp.Error = resp.ErrorMessage.String
	}

	assert.Equal(t, "resp_001", statusResp.ID)
	assert.Equal(t, "failed", statusResp.Status)
	assert.Equal(t, "upstream timeout", statusResp.Error)
	assert.Empty(t, statusResp.ResponseID)
}

func TestBackgroundStatusResponse_Completed(t *testing.T) {
	resp := &models.StoredResponse{
		ResponseID: "resp_001",
		Status:     "completed",
	}

	statusResp := BackgroundStatusResponse{
		ID:     resp.ResponseID,
		Object: "response",
		Status: resp.Status,
	}
	if resp.Status == "completed" {
		statusResp.ResponseID = resp.ResponseID
	}

	assert.Equal(t, "completed", statusResp.Status)
	assert.Equal(t, "resp_001", statusResp.ResponseID)
}

func TestResponseAPIResponse_WithError(t *testing.T) {
	resp := &models.StoredResponse{
		ResponseID:   "resp_001",
		Model:        "gpt-4o",
		Status:       "failed",
		ErrorMessage: sql.NullString{String: "rate limit exceeded", Valid: true},
	}

	apiResp := ResponseAPIResponse{
		ID:        resp.ResponseID,
		Object:    "response",
		Model:     resp.Model,
		Status:    resp.Status,
	}
	if resp.ErrorMessage.Valid && resp.Status == "failed" {
		apiResp.Error = map[string]string{
			"message": resp.ErrorMessage.String,
			"type":    "upstream_error",
		}
	}

	assert.Equal(t, "failed", apiResp.Status)
	errObj, ok := apiResp.Error.(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, "rate limit exceeded", errObj["message"])
	assert.Equal(t, "upstream_error", errObj["type"])
}

func TestBuildUpstreamRequestBody_WithReasoning(t *testing.T) {
	req := &ResponseRequest{
		Model:     "o3-mini",
		Input:     ResponseInput{Text: "Solve this"},
		Reasoning: map[string]interface{}{"effort": "high"},
	}

	body := buildUpstreamRequestBody(req, nil)
	assert.Equal(t, "o3-mini", body["model"])
	assert.Equal(t, "Solve this", body["input"])
	reasoning, ok := body["reasoning"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "high", reasoning["effort"])
}

func TestBuildUpstreamRequestBody_WithTruncate(t *testing.T) {
	truncate := 100
	req := &ResponseRequest{
		Model:    "gpt-4o",
		Input:    ResponseInput{Text: "Hello"},
		Truncate: &truncate,
	}

	body := buildUpstreamRequestBody(req, nil)
	assert.Equal(t, 100, body["truncate"])
}

func TestBuildUpstreamRequestBody_WithMetadata(t *testing.T) {
	req := &ResponseRequest{
		Model:    "gpt-4o",
		Input:    ResponseInput{Text: "Hello"},
		Metadata: map[string]interface{}{"user_id": "123"},
	}

	body := buildUpstreamRequestBody(req, nil)
	metadata, ok := body["metadata"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "123", metadata["user_id"])
}

func TestBuildUpstreamRequestBody_InputArray(t *testing.T) {
	req := &ResponseRequest{
		Model: "gpt-4o",
		Input: ResponseInput{
			Parts: []ResponseInputMessagePart{
				{Role: "user", Content: json.RawMessage(`"Hello"`)},
				{Role: "assistant", Content: json.RawMessage(`"Hi there"`)},
			},
		},
	}

	body := buildUpstreamRequestBody(req, nil)
	assert.Equal(t, "gpt-4o", body["model"])
	parts, ok := body["input"].([]ResponseInputMessagePart)
	assert.True(t, ok)
	assert.Len(t, parts, 2)
}

func TestEstimateResponseTokens_WithParts(t *testing.T) {
	req := &ResponseRequest{
		Input: ResponseInput{
			Parts: []ResponseInputMessagePart{
				{Role: "user", Content: json.RawMessage(`"This is a longer message for token estimation"`)},
			},
		},
	}
	tokens := estimateResponseTokens(req)
	assert.Greater(t, tokens, 0)
}

func setupResponseTestRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	origDB := config.DB
	config.DB = db
	t.Cleanup(func() {
		config.DB = origDB
		db.Close()
	})
	return r, mock
}

func TestOpenAIResponsesGet_NotFound(t *testing.T) {
	r, mock := setupResponseTestRouter(t)

	mock.ExpectQuery(`SELECT .+ FROM stored_responses WHERE response_id = .+ AND deleted_at IS NULL`).
		WithArgs("resp_nonexist").
		WillReturnError(sql.ErrNoRows)

	r.GET("/v1/responses/:id", func(c *gin.Context) {
		OpenAIResponsesGet(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/responses/resp_nonexist", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOpenAIResponsesDelete_NotFound(t *testing.T) {
	r, mock := setupResponseTestRouter(t)
	

	mock.ExpectQuery(`SELECT .+ FROM stored_responses WHERE response_id = .+ AND deleted_at IS NULL`).
		WithArgs("resp_nonexist").
		WillReturnError(sql.ErrNoRows)

	r.DELETE("/v1/responses/:id", func(c *gin.Context) {
		OpenAIResponsesDelete(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/responses/resp_nonexist", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOpenAIResponsesStatus_NotFound(t *testing.T) {
	r, mock := setupResponseTestRouter(t)
	

	mock.ExpectQuery(`SELECT .+ FROM stored_responses WHERE response_id = .+ AND deleted_at IS NULL`).
		WithArgs("resp_nonexist").
		WillReturnError(sql.ErrNoRows)

	r.GET("/v1/responses/:id/status", func(c *gin.Context) {
		OpenAIResponsesStatus(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/responses/resp_nonexist/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOpenAIResponsesGet_DBError(t *testing.T) {
	r, mock := setupResponseTestRouter(t)
	

	mock.ExpectQuery(`SELECT .+ FROM stored_responses WHERE response_id = .+ AND deleted_at IS NULL`).
		WithArgs("resp_001").
		WillReturnError(sql.ErrConnDone)

	r.GET("/v1/responses/:id", func(c *gin.Context) {
		OpenAIResponsesGet(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/responses/resp_001", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOpenAIResponsesStatus_DBError(t *testing.T) {
	r, mock := setupResponseTestRouter(t)

	mock.ExpectQuery(`SELECT .+ FROM stored_responses WHERE response_id = .+ AND deleted_at IS NULL`).
		WithArgs("resp_001").
		WillReturnError(sql.ErrConnDone)

	r.GET("/v1/responses/:id/status", func(c *gin.Context) {
		OpenAIResponsesStatus(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/responses/resp_001/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
