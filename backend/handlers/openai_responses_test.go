package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
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
	count := calculateToolBilling(output)
	assert.Equal(t, 2, count)
}

func TestCalculateToolBilling_NoTools(t *testing.T) {
	output := []ResponseOutputItem{
		{Type: "message", Status: "completed"},
	}
	count := calculateToolBilling(output)
	assert.Equal(t, 0, count)
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
