package handlers

import (
	"encoding/json"
	"testing"
)

func TestAnthropicToChatMessages_StringContent(t *testing.T) {
	sys := json.RawMessage(`"你是助手"`)
	msgs := json.RawMessage(`[{"role":"user","content":"hi"}]`)
	out, err := anthropicToChatMessages(sys, msgs)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].Role != "system" || out[1].Role != "user" {
		t.Fatalf("got %+v", out)
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
