package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildUpstreamRequest_MultimodalContent(t *testing.T) {
	t.Run("preserves string content format", func(t *testing.T) {
		req := APIProxyRequest{
			Provider: "openai",
			Model:    "gpt-4",
			Messages: []ChatMessage{
				{
					Role: "user",
					Content: MessageContent{
						Text: "Hello, world!",
					},
				},
			},
		}

		rb := map[string]interface{}{
			"model":    req.Model,
			"messages": req.Messages,
			"stream":   false,
		}

		data, err := json.Marshal(rb)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		messages := parsed["messages"].([]interface{})
		require.Len(t, messages, 1)

		msg := messages[0].(map[string]interface{})
		assert.Equal(t, "user", msg["role"])
		assert.Equal(t, "Hello, world!", msg["content"])
	})

	t.Run("preserves array content format", func(t *testing.T) {
		req := APIProxyRequest{
			Provider: "openai",
			Model:    "gpt-4-vision-preview",
			Messages: []ChatMessage{
				{
					Role: "user",
					Content: MessageContent{
						Parts: []ContentPart{
							{
								Type: "text",
								Text: "What's in this image?",
							},
							{
								Type: "image_url",
								ImageURL: &ImageURL{
									URL:    "https://example.com/image.png",
									Detail: "high",
								},
							},
						},
					},
				},
			},
		}

		rb := map[string]interface{}{
			"model":    req.Model,
			"messages": req.Messages,
			"stream":   false,
		}

		data, err := json.Marshal(rb)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		messages := parsed["messages"].([]interface{})
		require.Len(t, messages, 1)

		msg := messages[0].(map[string]interface{})
		assert.Equal(t, "user", msg["role"])

		content := msg["content"].([]interface{})
		require.Len(t, content, 2)

		textPart := content[0].(map[string]interface{})
		assert.Equal(t, "text", textPart["type"])
		assert.Equal(t, "What's in this image?", textPart["text"])

		imagePart := content[1].(map[string]interface{})
		assert.Equal(t, "image_url", imagePart["type"])
		imageURL := imagePart["image_url"].(map[string]interface{})
		assert.Equal(t, "https://example.com/image.png", imageURL["url"])
		assert.Equal(t, "high", imageURL["detail"])
	})

	t.Run("preserves mixed message types", func(t *testing.T) {
		req := APIProxyRequest{
			Provider: "openai",
			Model:    "gpt-4-vision-preview",
			Messages: []ChatMessage{
				{
					Role: "system",
					Content: MessageContent{
						Text: "You are a helpful assistant.",
					},
				},
				{
					Role: "user",
					Content: MessageContent{
						Parts: []ContentPart{
							{
								Type: "text",
								Text: "What's in this image?",
							},
							{
								Type: "image_url",
								ImageURL: &ImageURL{
									URL: "https://example.com/image.png",
								},
							},
						},
					},
				},
				{
					Role: "assistant",
					Content: MessageContent{
						Text: "This is a beautiful sunset.",
					},
				},
			},
		}

		rb := map[string]interface{}{
			"model":    req.Model,
			"messages": req.Messages,
			"stream":   false,
		}

		data, err := json.Marshal(rb)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		messages := parsed["messages"].([]interface{})
		require.Len(t, messages, 3)

		systemMsg := messages[0].(map[string]interface{})
		assert.Equal(t, "system", systemMsg["role"])
		assert.Equal(t, "You are a helpful assistant.", systemMsg["content"])

		userMsg := messages[1].(map[string]interface{})
		assert.Equal(t, "user", userMsg["role"])
		userContent := userMsg["content"].([]interface{})
		require.Len(t, userContent, 2)

		assistantMsg := messages[2].(map[string]interface{})
		assert.Equal(t, "assistant", assistantMsg["role"])
		assert.Equal(t, "This is a beautiful sunset.", assistantMsg["content"])
	})
}

func TestUpstreamRequest_OpenAICompatibility(t *testing.T) {
	t.Run("OpenAI format with multimodal content", func(t *testing.T) {
		messages := []ChatMessage{
			{
				Role: "user",
				Content: MessageContent{
					Parts: []ContentPart{
						{
							Type: "text",
							Text: "Describe this image",
						},
						{
							Type: "image_url",
							ImageURL: &ImageURL{
								URL:    "https://example.com/image.png",
								Detail: "high",
							},
						},
					},
				},
			},
		}

		rb := map[string]interface{}{
			"model":    "gpt-4-vision-preview",
			"messages": messages,
			"stream":   false,
		}

		data, err := json.Marshal(rb)
		require.NoError(t, err)

		expected := `{"model":"gpt-4-vision-preview","messages":[{"role":"user","content":[{"type":"text","text":"Describe this image"},{"type":"image_url","image_url":{"url":"https://example.com/image.png","detail":"high"}}]}],"stream":false}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("OpenAI format with text only", func(t *testing.T) {
		messages := []ChatMessage{
			{
				Role: "user",
				Content: MessageContent{
					Text: "Hello!",
				},
			},
		}

		rb := map[string]interface{}{
			"model":    "gpt-4",
			"messages": messages,
			"stream":   false,
		}

		data, err := json.Marshal(rb)
		require.NoError(t, err)

		expected := `{"model":"gpt-4","messages":[{"role":"user","content":"Hello!"}],"stream":false}`
		assert.JSONEq(t, expected, string(data))
	})
}

func TestUpstreamRequest_AnthropicCompatibility(t *testing.T) {
	t.Run("Anthropic format with multimodal content", func(t *testing.T) {
		messages := []ChatMessage{
			{
				Role: "user",
				Content: MessageContent{
					Parts: []ContentPart{
						{
							Type: "text",
							Text: "What's in this image?",
						},
						{
							Type: "image_url",
							ImageURL: &ImageURL{
								URL: "https://example.com/image.png",
							},
						},
					},
				},
			},
		}

		rb := map[string]interface{}{
			"model":      "claude-3-opus-20240229",
			"messages":   messages,
			"max_tokens": 1024,
		}

		data, err := json.Marshal(rb)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "claude-3-opus-20240229", parsed["model"])
		assert.Equal(t, float64(1024), parsed["max_tokens"])

		msgs := parsed["messages"].([]interface{})
		require.Len(t, msgs, 1)

		msg := msgs[0].(map[string]interface{})
		assert.Equal(t, "user", msg["role"])

		content := msg["content"].([]interface{})
		require.Len(t, content, 2)
	})
}

func TestUpstreamRequest_ZhipuCompatibility(t *testing.T) {
	t.Run("Zhipu format with multimodal content", func(t *testing.T) {
		messages := []ChatMessage{
			{
				Role: "user",
				Content: MessageContent{
					Parts: []ContentPart{
						{
							Type: "text",
							Text: "分析这张图片",
						},
						{
							Type: "image_url",
							ImageURL: &ImageURL{
								URL: "https://example.com/image.png",
							},
						},
					},
				},
			},
		}

		rb := map[string]interface{}{
			"model":    "glm-4v",
			"messages": messages,
			"stream":   false,
		}

		data, err := json.Marshal(rb)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "glm-4v", parsed["model"])

		msgs := parsed["messages"].([]interface{})
		require.Len(t, msgs, 1)

		msg := msgs[0].(map[string]interface{})
		content := msg["content"].([]interface{})
		require.Len(t, content, 2)
	})
}
