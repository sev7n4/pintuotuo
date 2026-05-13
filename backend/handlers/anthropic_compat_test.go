package handlers

import (
	"encoding/json"
	"testing"
)

func TestAnthropicJSONNumberToInt(t *testing.T) {
	if n, ok := anthropicJSONNumberToInt(json.Number("42")); !ok || n != 42 {
		t.Fatalf("json.Number: ok=%v n=%d", ok, n)
	}
	if n, ok := anthropicJSONNumberToInt(float64(7)); !ok || n != 7 {
		t.Fatalf("float64: ok=%v n=%d", ok, n)
	}
	if _, ok := anthropicJSONNumberToInt("x"); ok {
		t.Fatal("expected false for string")
	}
}

func TestOpenAICompletionToAnthropicMessage(t *testing.T) {
	raw := []byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`)
	m, err := openAICompletionToAnthropicMessage(raw, "alibaba/qwen-turbo")
	if err != nil {
		t.Fatal(err)
	}
	if m["model"] != "alibaba/qwen-turbo" {
		t.Fatalf("model: %v", m["model"])
	}
	content, ok := m["content"].([]map[string]interface{})
	if !ok || len(content) != 1 || content[0]["text"] != "ok" {
		t.Fatalf("content: %#v", m["content"])
	}
}
