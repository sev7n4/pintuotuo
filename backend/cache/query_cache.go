package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pintuotuo/backend/db"
)

type QueryCache struct {
	cache *RedisCache
	ttl   time.Duration
}

func NewQueryCache(cache *RedisCache, ttl time.Duration) *QueryCache {
	return &QueryCache{
		cache: cache,
		ttl:   ttl,
	}
}

func (qc *QueryCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := qc.cache.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

func (qc *QueryCache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return qc.cache.Set(ctx, key, string(data), qc.ttl)
}

func (qc *QueryCache) Delete(ctx context.Context, key string) error {
	return qc.cache.Delete(ctx, key)
}

func (qc *QueryCache) DeletePattern(ctx context.Context, pattern string) error {
	return qc.cache.DeletePattern(ctx, pattern)
}

func (qc *QueryCache) GetOrSet(ctx context.Context, key string, dest interface{}, fetchFn func() (interface{}, error)) error {
	err := qc.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	data, err := fetchFn()
	if err != nil {
		return err
	}

	if err := qc.Set(ctx, key, data); err != nil {
		return err
	}

	dataBytes, _ := json.Marshal(data)
	return json.Unmarshal(dataBytes, dest)
}

func CacheKey(parts ...string) string {
	return "query:" + joinParts(parts)
}

func UserCacheKey(userID int64, parts ...string) string {
	return CacheKey(append([]string{"user", fmt.Sprintf("%d", userID)}, parts...)...)
}

func MerchantCacheKey(merchantID int64, parts ...string) string {
	return CacheKey(append([]string{"merchant", fmt.Sprintf("%d", merchantID)}, parts...)...)
}

func ProductCacheKey(productID int64) string {
	return CacheKey("product", fmt.Sprintf("%d", productID))
}

func OrderCacheKey(orderID int64) string {
	return CacheKey("order", fmt.Sprintf("%d", orderID))
}

func joinParts(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ":"
		}
		result += part
	}
	return result
}

func InvalidateUserCache(ctx context.Context, cache *RedisCache, userID int64) error {
	pattern := fmt.Sprintf("query:user:%d:*", userID)
	return cache.DeletePattern(ctx, pattern)
}

func InvalidateMerchantCache(ctx context.Context, cache *RedisCache, merchantID int64) error {
	pattern := fmt.Sprintf("query:merchant:%d:*", merchantID)
	return cache.DeletePattern(ctx, pattern)
}

func InvalidateProductCache(ctx context.Context, cache *RedisCache, productID int64) error {
	key := ProductCacheKey(productID)
	return cache.Delete(ctx, key)
}

func GetDBPoolStats(ctx context.Context) (*db.PoolStats, error) {
	stats := db.GetPoolStats()
	if stats == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return stats, nil
}
