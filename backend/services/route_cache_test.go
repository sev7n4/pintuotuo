package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteCache_GetSet(t *testing.T) {
	cache := NewRouteCache(5 * time.Minute)

	cfg := providerRuntimeConfig{
		Code:           "openai",
		APIBaseURL:     "https://api.openai.com/v1",
		ProviderRegion: "overseas",
	}

	cache.Set("openai", cfg)

	result, ok := cache.Get("openai")
	require.True(t, ok)
	assert.Equal(t, "openai", result.Code)
	assert.Equal(t, "https://api.openai.com/v1", result.APIBaseURL)
}

func TestRouteCache_Get_Miss(t *testing.T) {
	cache := NewRouteCache(5 * time.Minute)

	result, ok := cache.Get("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestRouteCache_Get_Expired(t *testing.T) {
	cache := NewRouteCache(100 * time.Millisecond)

	cfg := providerRuntimeConfig{
		Code:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
	}

	cache.Set("openai", cfg)

	time.Sleep(150 * time.Millisecond)

	result, ok := cache.Get("openai")
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestRouteCache_Invalidate(t *testing.T) {
	cache := NewRouteCache(5 * time.Minute)

	cfg := providerRuntimeConfig{
		Code:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
	}

	cache.Set("openai", cfg)
	cache.Invalidate("openai")

	result, ok := cache.Get("openai")
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestRouteCache_InvalidateAll(t *testing.T) {
	cache := NewRouteCache(5 * time.Minute)

	cfg1 := providerRuntimeConfig{Code: "openai"}
	cfg2 := providerRuntimeConfig{Code: "anthropic"}

	cache.Set("openai", cfg1)
	cache.Set("anthropic", cfg2)

	assert.Equal(t, 2, cache.Size())

	cache.InvalidateAll()

	assert.Equal(t, 0, cache.Size())

	result1, ok1 := cache.Get("openai")
	assert.False(t, ok1)
	assert.Nil(t, result1)

	result2, ok2 := cache.Get("anthropic")
	assert.False(t, ok2)
	assert.Nil(t, result2)
}

func TestRouteCache_Disabled(t *testing.T) {
	cache := NewRouteCache(0)

	cfg := providerRuntimeConfig{
		Code:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
	}

	cache.Set("openai", cfg)

	result, ok := cache.Get("openai")
	assert.False(t, ok)
	assert.Nil(t, result)
	assert.Equal(t, 0, cache.Size())
}

func TestRouteCache_Size(t *testing.T) {
	cache := NewRouteCache(5 * time.Minute)

	assert.Equal(t, 0, cache.Size())

	cache.Set("openai", providerRuntimeConfig{Code: "openai"})
	assert.Equal(t, 1, cache.Size())

	cache.Set("anthropic", providerRuntimeConfig{Code: "anthropic"})
	assert.Equal(t, 2, cache.Size())

	cache.Invalidate("openai")
	assert.Equal(t, 1, cache.Size())
}
