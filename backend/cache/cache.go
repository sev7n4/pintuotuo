package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

const (
	ProductCacheTTL  = 1 * time.Hour
	ProductListTTL   = 5 * time.Minute
	UserCacheTTL     = 30 * time.Minute
	GroupCacheTTL    = time.Duration(0) // No caching for real-time groups
	OrderCacheTTL    = 10 * time.Minute
	TokenBalanceTTL  = 5 * time.Minute
	SearchResultsTTL = 10 * time.Minute
	ReferralCodeTTL  = 30 * time.Minute
	ReferralStatsTTL = 5 * time.Minute
	MerchantCacheTTL = 30 * time.Minute
)

// Init initializes the Redis client
func Init() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client = redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// Get retrieves a value from cache
func Get(ctx context.Context, key string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	return client.Get(ctx, key).Result()
}

// Set sets a value in cache with TTL
func Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return client.Set(ctx, key, value, ttl).Err()
}

// SetNX sets a value only if the key doesn't exist
func SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	result, err := client.SetNX(ctx, key, value, ttl).Result()
	return result, err
}

// Delete removes one or more keys from cache
func Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in cache
func Exists(ctx context.Context, key string) (bool, error) {
	if client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	val, err := client.Exists(ctx, key).Result()
	return val > 0, err
}

// Increment increments an integer value
func Increment(ctx context.Context, key string) (int64, error) {
	if client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}
	return client.Incr(ctx, key).Result()
}

// Decrement decrements an integer value
func Decrement(ctx context.Context, key string) (int64, error) {
	if client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}
	return client.Decr(ctx, key).Result()
}

// IncrementBy increments by a specific amount
func IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	if client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}
	return client.IncrBy(ctx, key, delta).Result()
}

// GetClient returns the Redis client for advanced operations
func GetClient() *redis.Client {
	return client
}

// Cache key builders
func ProductKey(id int) string {
	return fmt.Sprintf("product:%d", id)
}

func ProductListKey(page, perPage int, status string) string {
	return fmt.Sprintf("products:list:%s:page:%d:limit:%d", status, page, perPage)
}

func ProductSearchKey(query string, page, perPage int) string {
	return fmt.Sprintf("products:search:%s:page:%d:limit:%d", query, page, perPage)
}

func UserKey(id int) string {
	return fmt.Sprintf("user:%d", id)
}

func GroupKey(id int) string {
	return fmt.Sprintf("group:%d", id)
}

func GroupListKey(page, perPage int, status string) string {
	return fmt.Sprintf("groups:list:%s:page:%d:limit:%d", status, page, perPage)
}

func OrderKey(id int) string {
	return fmt.Sprintf("order:%d", id)
}

func OrderListKey(userID, page, perPage int) string {
	return fmt.Sprintf("orders:user:%d:page:%d:limit:%d", userID, page, perPage)
}

func TokenBalanceKey(userID int) string {
	return fmt.Sprintf("token:balance:%d", userID)
}

func SessionKey(userID int) string {
	return fmt.Sprintf("session:%d", userID)
}

func ReferralCodeKey(userID int) string {
	return fmt.Sprintf("referral:code:%d", userID)
}

func ReferralStatsKey(userID int) string {
	return fmt.Sprintf("referral:stats:%d", userID)
}

func MerchantKey(userID int) string {
	return fmt.Sprintf("merchant:%d", userID)
}

func MerchantProductsKey(merchantID int) string {
	return fmt.Sprintf("merchant:%d:products", merchantID)
}

func MerchantAPIKeysKey(merchantID int) string {
	return fmt.Sprintf("merchant:%d:apikeys", merchantID)
}

func SPUKey(id int) string {
	return fmt.Sprintf("spu:%d", id)
}

func SPUListKey(page, perPage int, provider, tier, status string) string {
	return fmt.Sprintf("spus:list:%s:%s:%s:page:%d:limit:%d", provider, tier, status, page, perPage)
}

func SKUKey(id int) string {
	return fmt.Sprintf("sku:%d", id)
}

func SKUListKey(page, perPage int, spuID, skuType, scope, skuStatus, spuStatus, provider, q, misaligned string) string {
	return fmt.Sprintf("skus:list:%s:%s:%s:%s:%s:%s:%s:%s:page:%d:limit:%d",
		spuID, skuType, scope, skuStatus, spuStatus, provider, q, misaligned, page, perPage)
}

func ComputePointBalanceKey(userID int) string {
	return fmt.Sprintf("compute_points:balance:%d", userID)
}

func MerchantSKUsKey(merchantID int, status string) string {
	return fmt.Sprintf("merchant:%d:skus:%s", merchantID, status)
}

func AvailableSKUsKey(merchantID int, provider, skuType string) string {
	return fmt.Sprintf("merchant:%d:available_skus:%s:%s", merchantID, provider, skuType)
}

// InvalidatePatterns invalidates cache keys matching a pattern
func InvalidatePatterns(ctx context.Context, pattern string) error {
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	iter := client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return client.Del(ctx, keys...).Err()
	}

	return nil
}

func DeletePattern(ctx context.Context, pattern string) error {
	return InvalidatePatterns(ctx, pattern)
}

func HealthCheck(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return client.Ping(ctx).Err()
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache() *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return Get(ctx, key)
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return Set(ctx, key, value, ttl)
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return Delete(ctx, key)
}

func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	return DeletePattern(ctx, pattern)
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	return Exists(ctx, key)
}
