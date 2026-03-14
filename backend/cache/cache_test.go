package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheKeyBuilders(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "ProductKey",
			fn:       func() string { return ProductKey(123) },
			expected: "product:123",
		},
		{
			name:     "ProductListKey",
			fn:       func() string { return ProductListKey(1, 20, "active") },
			expected: "products:list:active:page:1:limit:20",
		},
		{
			name:     "UserKey",
			fn:       func() string { return UserKey(456) },
			expected: "user:456",
		},
		{
			name:     "GroupKey",
			fn:       func() string { return GroupKey(789) },
			expected: "group:789",
		},
		{
			name:     "OrderKey",
			fn:       func() string { return OrderKey(111) },
			expected: "order:111",
		},
		{
			name:     "TokenBalanceKey",
			fn:       func() string { return TokenBalanceKey(222) },
			expected: "token:balance:222",
		},
		{
			name:     "SessionKey",
			fn:       func() string { return SessionKey(333) },
			expected: "session:333",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCacheTTLConstants(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		expected time.Duration
	}{
		{"ProductCacheTTL", ProductCacheTTL, 1 * time.Hour},
		{"ProductListTTL", ProductListTTL, 5 * time.Minute},
		{"UserCacheTTL", UserCacheTTL, 30 * time.Minute},
		{"OrderCacheTTL", OrderCacheTTL, 10 * time.Minute},
		{"TokenBalanceTTL", TokenBalanceTTL, 5 * time.Minute},
		{"SearchResultsTTL", SearchResultsTTL, 10 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ttl)
		})
	}
}

func TestCacheTTLOrdering(t *testing.T) {
	// Verify TTL ordering makes sense
	assert.Greater(t, ProductCacheTTL, ProductListTTL, "Product detail cache should be longer than list cache")
	assert.Greater(t, UserCacheTTL, ProductListTTL, "User cache should be longer than product list cache")
	assert.Greater(t, ProductCacheTTL, GroupCacheTTL, "Product cache should be longer than group cache (which is 0)")
}

func TestProductSearchKeyBuilding(t *testing.T) {
	key := ProductSearchKey("laptop", 1, 20)
	assert.Equal(t, "products:search:laptop:page:1:limit:20", key)

	key2 := ProductSearchKey("iphone pro max", 2, 50)
	assert.Equal(t, "products:search:iphone pro max:page:2:limit:50", key2)
}

func TestGroupListKeyBuilding(t *testing.T) {
	activeKey := GroupListKey(1, 20, "active")
	assert.Equal(t, "groups:list:active:page:1:limit:20", activeKey)

	completedKey := GroupListKey(2, 50, "completed")
	assert.Equal(t, "groups:list:completed:page:2:limit:50", completedKey)
}

func TestOrderListKeyBuilding(t *testing.T) {
	key := OrderListKey(123, 1, 20)
	assert.Equal(t, "orders:user:123:page:1:limit:20", key)

	key2 := OrderListKey(999, 5, 100)
	assert.Equal(t, "orders:user:999:page:5:limit:100", key2)
}

// Test cache key collisions are unlikely with different parameters
func TestCacheKeyUniqueness(t *testing.T) {
	keys := map[string]bool{}

	// Generate various keys
	testKeys := []string{
		ProductKey(1),
		ProductKey(2),
		ProductListKey(1, 20, "active"),
		ProductListKey(1, 20, "inactive"),
		ProductListKey(2, 20, "active"),
		UserKey(1),
		UserKey(2),
		GroupKey(1),
		GroupKey(2),
		OrderKey(1),
		OrderKey(2),
		TokenBalanceKey(1),
		TokenBalanceKey(2),
		SessionKey(1),
		SessionKey(2),
	}

	for _, key := range testKeys {
		assert.False(t, keys[key], "Key %s already exists", key)
		keys[key] = true
	}

	assert.Equal(t, len(testKeys), len(keys), "All keys should be unique")
}

// Integration test - structure is sound (actual redis test would be in integration tests)
func TestCacheStructure(t *testing.T) {
	// Verify cache constants are configured correctly
	assert.True(t, ProductCacheTTL > 0, "ProductCacheTTL should be positive")
	assert.True(t, ProductListTTL > 0, "ProductListTTL should be positive")
	assert.True(t, UserCacheTTL > 0, "UserCacheTTL should be positive")
	assert.True(t, TokenBalanceTTL > 0, "TokenBalanceTTL should be positive")
	assert.Equal(t, time.Duration(0), GroupCacheTTL, "GroupCacheTTL should be 0 (no caching for real-time data)")
}
