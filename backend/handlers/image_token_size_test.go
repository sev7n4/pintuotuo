package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEstimateImageTokensWithSize(t *testing.T) {
	tests := []struct {
		name     string
		detail   string
		width    int
		height   int
		expected int
	}{
		{
			name:     "low detail ignores size",
			detail:   "low",
			width:    2048,
			height:   2048,
			expected: 85,
		},
		{
			name:     "high detail 1024x1024",
			detail:   "high",
			width:    1024,
			height:   1024,
			expected: 85 + 2*2*170 + 255,
		},
		{
			name:     "high detail 512x512",
			detail:   "high",
			width:    512,
			height:   512,
			expected: 85 + 1*1*170 + 255,
		},
		{
			name:     "high detail 2048x2048",
			detail:   "high",
			width:    2048,
			height:   2048,
			expected: 85 + 4*4*170 + 255,
		},
		{
			name:     "high detail 768x768",
			detail:   "high",
			width:    768,
			height:   768,
			expected: 85 + 2*2*170 + 255,
		},
		{
			name:     "high detail with zero size falls back to default",
			detail:   "high",
			width:    0,
			height:   0,
			expected: 765,
		},
		{
			name:     "auto detail with size uses dynamic calculation",
			detail:   "auto",
			width:    1024,
			height:   1024,
			expected: 85 + 2*2*170 + 255,
		},
		{
			name:     "auto detail without size falls back to default",
			detail:   "auto",
			width:    0,
			height:   0,
			expected: 765,
		},
		{
			name:     "empty detail with size uses dynamic calculation",
			detail:   "",
			width:    512,
			height:   512,
			expected: 85 + 1*1*170 + 255,
		},
		{
			name:     "non-standard size 600x400",
			detail:   "high",
			width:    600,
			height:   400,
			expected: 85 + 2*1*170 + 255,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateImageTokensWithSize(tt.detail, tt.width, tt.height)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateImageTokens_BackwardCompatible(t *testing.T) {
	assert.Equal(t, 85, estimateImageTokens("low"))
	assert.Equal(t, 765, estimateImageTokens("high"))
	assert.Equal(t, 765, estimateImageTokens("auto"))
	assert.Equal(t, 765, estimateImageTokens(""))
}
