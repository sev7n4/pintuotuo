package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEstimateImageTokens(t *testing.T) {
	tests := []struct {
		name     string
		detail   string
		expected int
	}{
		{
			name:     "low detail image",
			detail:   "low",
			expected: 85,
		},
		{
			name:     "high detail image (default 1024x1024)",
			detail:   "high",
			expected: 765,
		},
		{
			name:     "auto detail defaults to high",
			detail:   "auto",
			expected: 765,
		},
		{
			name:     "empty detail defaults to high",
			detail:   "",
			expected: 765,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateImageTokens(tt.detail)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateInputTokens_WithImages(t *testing.T) {
	tests := []struct {
		name     string
		messages []ChatMessage
		expected int
	}{
		{
			name: "text only message",
			messages: []ChatMessage{
				{
					Role: "user",
					Content: MessageContent{
						Text: "Hello, world!",
					},
				},
			},
			expected: (len("user") + len("Hello, world!")) / 4,
		},
		{
			name: "multimodal message with one image",
			messages: []ChatMessage{
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
			expected: (len("user") + len("What's in this image?") + 765) / 4,
		},
		{
			name: "multimodal message with low detail image",
			messages: []ChatMessage{
				{
					Role: "user",
					Content: MessageContent{
						Parts: []ContentPart{
							{
								Type: "text",
								Text: "Describe this",
							},
							{
								Type: "image_url",
								ImageURL: &ImageURL{
									URL:    "https://example.com/image.png",
									Detail: "low",
								},
							},
						},
					},
				},
			},
			expected: (len("user") + len("Describe this") + 85) / 4,
		},
		{
			name: "multimodal message with multiple images",
			messages: []ChatMessage{
				{
					Role: "user",
					Content: MessageContent{
						Parts: []ContentPart{
							{
								Type: "text",
								Text: "Compare these images",
							},
							{
								Type: "image_url",
								ImageURL: &ImageURL{
									URL:    "https://example.com/image1.png",
									Detail: "high",
								},
							},
							{
								Type: "image_url",
								ImageURL: &ImageURL{
									URL:    "https://example.com/image2.png",
									Detail: "low",
								},
							},
						},
					},
				},
			},
			expected: (len("user") + len("Compare these images") + 765 + 85) / 4,
		},
		{
			name: "mixed messages",
			messages: []ChatMessage{
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
								Text: "What's this?",
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
			expected: (len("system") + len("You are a helpful assistant.") + len("user") + len("What's this?") + 765) / 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateInputTokens(tt.messages)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateInputTokens_EdgeCases(t *testing.T) {
	t.Run("empty message content", func(t *testing.T) {
		messages := []ChatMessage{
			{
				Role:    "user",
				Content: MessageContent{},
			},
		}
		result := estimateInputTokens(messages)
		assert.Equal(t, len("user")/4, result)
	})

	t.Run("message with empty parts", func(t *testing.T) {
		messages := []ChatMessage{
			{
				Role: "user",
				Content: MessageContent{
					Parts: []ContentPart{},
				},
			},
		}
		result := estimateInputTokens(messages)
		assert.Equal(t, len("user")/4, result)
	})

	t.Run("image part without ImageURL", func(t *testing.T) {
		messages := []ChatMessage{
			{
				Role: "user",
				Content: MessageContent{
					Parts: []ContentPart{
						{
							Type: "image_url",
						},
					},
				},
			},
		}
		result := estimateInputTokens(messages)
		assert.Equal(t, (len("user")+765)/4, result)
	})
}
