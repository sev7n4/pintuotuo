package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatMessage_StructDefinition(t *testing.T) {
	t.Run("simple text message", func(t *testing.T) {
		msg := ChatMessage{
			Role: "user",
			Content: MessageContent{
				Text: "Hello, world!",
			},
		}

		assert.Equal(t, "user", msg.Role)
		assert.Equal(t, "Hello, world!", msg.Content.Text)
		assert.Nil(t, msg.Content.Parts)
	})

	t.Run("multimodal message with text and image", func(t *testing.T) {
		msg := ChatMessage{
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
		}

		assert.Equal(t, "user", msg.Role)
		assert.Empty(t, msg.Content.Text)
		require.Len(t, msg.Content.Parts, 2)
		assert.Equal(t, "text", msg.Content.Parts[0].Type)
		assert.Equal(t, "What's in this image?", msg.Content.Parts[0].Text)
		assert.Equal(t, "image_url", msg.Content.Parts[1].Type)
		assert.Equal(t, "https://example.com/image.png", msg.Content.Parts[1].ImageURL.URL)
		assert.Equal(t, "high", msg.Content.Parts[1].ImageURL.Detail)
	})

	t.Run("message with name field", func(t *testing.T) {
		msg := ChatMessage{
			Role: "user",
			Name: "Alice",
			Content: MessageContent{
				Text: "Hello!",
			},
		}

		assert.Equal(t, "user", msg.Role)
		assert.Equal(t, "Alice", msg.Name)
		assert.Equal(t, "Hello!", msg.Content.Text)
	})
}

func TestMessageContent_UnmarshalJSON(t *testing.T) {
	t.Run("string content", func(t *testing.T) {
		jsonData := `"Hello, world!"`
		var content MessageContent
		err := json.Unmarshal([]byte(jsonData), &content)
		require.NoError(t, err)

		assert.Equal(t, "Hello, world!", content.Text)
		assert.Nil(t, content.Parts)
	})

	t.Run("array content with text only", func(t *testing.T) {
		jsonData := `[{"type": "text", "text": "Hello"}]`
		var content MessageContent
		err := json.Unmarshal([]byte(jsonData), &content)
		require.NoError(t, err)

		assert.Empty(t, content.Text)
		require.Len(t, content.Parts, 1)
		assert.Equal(t, "text", content.Parts[0].Type)
		assert.Equal(t, "Hello", content.Parts[0].Text)
	})

	t.Run("array content with text and image", func(t *testing.T) {
		jsonData := `[
			{"type": "text", "text": "What's in this image?"},
			{"type": "image_url", "image_url": {"url": "https://example.com/image.png", "detail": "high"}}
		]`
		var content MessageContent
		err := json.Unmarshal([]byte(jsonData), &content)
		require.NoError(t, err)

		assert.Empty(t, content.Text)
		require.Len(t, content.Parts, 2)
		assert.Equal(t, "text", content.Parts[0].Type)
		assert.Equal(t, "What's in this image?", content.Parts[0].Text)
		assert.Equal(t, "image_url", content.Parts[1].Type)
		require.NotNil(t, content.Parts[1].ImageURL)
		assert.Equal(t, "https://example.com/image.png", content.Parts[1].ImageURL.URL)
		assert.Equal(t, "high", content.Parts[1].ImageURL.Detail)
	})

	t.Run("array content with image without detail", func(t *testing.T) {
		jsonData := `[{"type": "image_url", "image_url": {"url": "https://example.com/image.png"}}]`
		var content MessageContent
		err := json.Unmarshal([]byte(jsonData), &content)
		require.NoError(t, err)

		assert.Empty(t, content.Text)
		require.Len(t, content.Parts, 1)
		assert.Equal(t, "image_url", content.Parts[0].Type)
		require.NotNil(t, content.Parts[0].ImageURL)
		assert.Equal(t, "https://example.com/image.png", content.Parts[0].ImageURL.URL)
		assert.Empty(t, content.Parts[0].ImageURL.Detail)
	})

	t.Run("empty string content", func(t *testing.T) {
		jsonData := `""`
		var content MessageContent
		err := json.Unmarshal([]byte(jsonData), &content)
		require.NoError(t, err)

		assert.Empty(t, content.Text)
		assert.Nil(t, content.Parts)
	})

	t.Run("empty array content", func(t *testing.T) {
		jsonData := `[]`
		var content MessageContent
		err := json.Unmarshal([]byte(jsonData), &content)
		require.NoError(t, err)

		assert.Empty(t, content.Text)
		assert.Nil(t, content.Parts)
	})

	t.Run("invalid JSON type", func(t *testing.T) {
		jsonData := `123`
		var content MessageContent
		err := json.Unmarshal([]byte(jsonData), &content)
		assert.Error(t, err)
	})
}

func TestMessageContent_MarshalJSON(t *testing.T) {
	t.Run("string content preserves format", func(t *testing.T) {
		content := MessageContent{
			Text: "Hello, world!",
		}

		data, err := json.Marshal(content)
		require.NoError(t, err)
		assert.Equal(t, `"Hello, world!"`, string(data))
	})

	t.Run("array content preserves format", func(t *testing.T) {
		content := MessageContent{
			Parts: []ContentPart{
				{
					Type: "text",
					Text: "Hello",
				},
				{
					Type: "image_url",
					ImageURL: &ImageURL{
						URL:    "https://example.com/image.png",
						Detail: "high",
					},
				},
			},
		}

		data, err := json.Marshal(content)
		require.NoError(t, err)

		var unmarshaled []ContentPart
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		require.Len(t, unmarshaled, 2)
		assert.Equal(t, "text", unmarshaled[0].Type)
		assert.Equal(t, "Hello", unmarshaled[0].Text)
		assert.Equal(t, "image_url", unmarshaled[1].Type)
	})

	t.Run("empty content marshals to empty string", func(t *testing.T) {
		content := MessageContent{}

		data, err := json.Marshal(content)
		require.NoError(t, err)
		assert.Equal(t, `""`, string(data))
	})
}

func TestChatMessage_JSONRoundTrip(t *testing.T) {
	t.Run("simple text message round trip", func(t *testing.T) {
		original := ChatMessage{
			Role: "user",
			Content: MessageContent{
				Text: "Hello, world!",
			},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled ChatMessage
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, original.Role, unmarshaled.Role)
		assert.Equal(t, original.Content.Text, unmarshaled.Content.Text)
		assert.Nil(t, unmarshaled.Content.Parts)
	})

	t.Run("multimodal message round trip", func(t *testing.T) {
		original := ChatMessage{
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
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled ChatMessage
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, original.Role, unmarshaled.Role)
		require.Len(t, unmarshaled.Content.Parts, 2)
		assert.Equal(t, "text", unmarshaled.Content.Parts[0].Type)
		assert.Equal(t, "What's in this image?", unmarshaled.Content.Parts[0].Text)
		assert.Equal(t, "image_url", unmarshaled.Content.Parts[1].Type)
		assert.Equal(t, "https://example.com/image.png", unmarshaled.Content.Parts[1].ImageURL.URL)
		assert.Equal(t, "high", unmarshaled.Content.Parts[1].ImageURL.Detail)
	})

	t.Run("message with name field round trip", func(t *testing.T) {
		original := ChatMessage{
			Role: "user",
			Name: "Alice",
			Content: MessageContent{
				Text: "Hello!",
			},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var unmarshaled ChatMessage
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, "user", unmarshaled.Role)
		assert.Equal(t, "Alice", unmarshaled.Name)
		assert.Equal(t, "Hello!", unmarshaled.Content.Text)
	})
}

func TestAPIProxyRequest_WithMultimodalMessages(t *testing.T) {
	t.Run("request with multimodal messages", func(t *testing.T) {
		jsonData := `{
			"provider": "openai",
			"model": "gpt-4-vision-preview",
			"messages": [
				{
					"role": "user",
					"content": [
						{"type": "text", "text": "What's in this image?"},
						{"type": "image_url", "image_url": {"url": "https://example.com/image.png", "detail": "high"}}
					]
				},
				{
					"role": "assistant",
					"content": "This is a beautiful sunset."
				}
			]
		}`

		var req APIProxyRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "openai", req.Provider)
		assert.Equal(t, "gpt-4-vision-preview", req.Model)
		require.Len(t, req.Messages, 2)

		assert.Equal(t, "user", req.Messages[0].Role)
		require.Len(t, req.Messages[0].Content.Parts, 2)
		assert.Equal(t, "text", req.Messages[0].Content.Parts[0].Type)
		assert.Equal(t, "What's in this image?", req.Messages[0].Content.Parts[0].Text)
		assert.Equal(t, "image_url", req.Messages[0].Content.Parts[1].Type)

		assert.Equal(t, "assistant", req.Messages[1].Role)
		assert.Equal(t, "This is a beautiful sunset.", req.Messages[1].Content.Text)
	})

	t.Run("request with mixed message types", func(t *testing.T) {
		jsonData := `{
			"provider": "openai",
			"model": "gpt-4",
			"messages": [
				{
					"role": "system",
					"content": "You are a helpful assistant."
				},
				{
					"role": "user",
					"content": "Hello!"
				}
			]
		}`

		var req APIProxyRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Len(t, req.Messages, 2)
		assert.Equal(t, "system", req.Messages[0].Role)
		assert.Equal(t, "You are a helpful assistant.", req.Messages[0].Content.Text)
		assert.Equal(t, "user", req.Messages[1].Role)
		assert.Equal(t, "Hello!", req.Messages[1].Content.Text)
	})
}
