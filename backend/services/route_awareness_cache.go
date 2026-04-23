package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/pintuotuo/backend/models"
)

type RouteAwarenessCache struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewRouteAwarenessCache(redisClient *redis.Client, ttl time.Duration) *RouteAwarenessCache {
	return &RouteAwarenessCache{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

func (c *RouteAwarenessCache) GetRealtimeStatus(ctx context.Context, apiKeyID int) (*models.APIKeyRealtimeStatus, error) {
	key := c.getStatusKey(apiKeyID)

	data, err := c.redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get status from cache: %w", err)
	}

	var status models.APIKeyRealtimeStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	return &status, nil
}

func (c *RouteAwarenessCache) SetRealtimeStatus(ctx context.Context, status *models.APIKeyRealtimeStatus) error {
	key := c.getStatusKey(status.APIKeyID)

	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	if err := c.redisClient.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set status in cache: %w", err)
	}

	return nil
}

func (c *RouteAwarenessCache) DeleteRealtimeStatus(ctx context.Context, apiKeyID int) error {
	key := c.getStatusKey(apiKeyID)

	if err := c.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete status from cache: %w", err)
	}

	return nil
}

func (c *RouteAwarenessCache) GetBatchStatus(ctx context.Context, apiKeyIDs []int) ([]*models.APIKeyRealtimeStatus, error) {
	if len(apiKeyIDs) == 0 {
		return []*models.APIKeyRealtimeStatus{}, nil
	}

	keys := make([]string, len(apiKeyIDs))
	for i, id := range apiKeyIDs {
		keys[i] = c.getStatusKey(id)
	}

	results, err := c.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get batch status from cache: %w", err)
	}

	statuses := make([]*models.APIKeyRealtimeStatus, 0, len(apiKeyIDs))
	for _, result := range results {
		if result == nil {
			continue
		}

		data, ok := result.(string)
		if !ok {
			continue
		}

		var status models.APIKeyRealtimeStatus
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue
		}

		statuses = append(statuses, &status)
	}

	return statuses, nil
}

func (c *RouteAwarenessCache) SetBatchStatus(ctx context.Context, statuses []*models.APIKeyRealtimeStatus) error {
	if len(statuses) == 0 {
		return nil
	}

	pipe := c.redisClient.Pipeline()
	for _, status := range statuses {
		key := c.getStatusKey(status.APIKeyID)

		data, err := json.Marshal(status)
		if err != nil {
			continue
		}

		pipe.Set(ctx, key, data, c.ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set batch status in cache: %w", err)
	}

	return nil
}

func (c *RouteAwarenessCache) IncrementRequestCount(ctx context.Context, apiKeyID int) error {
	key := c.getRequestCountKey(apiKeyID)

	if err := c.redisClient.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to increment request count: %w", err)
	}

	c.redisClient.Expire(ctx, key, time.Minute)

	return nil
}

func (c *RouteAwarenessCache) GetRequestCount(ctx context.Context, apiKeyID int) (int64, error) {
	key := c.getRequestCountKey(apiKeyID)

	count, err := c.redisClient.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get request count: %w", err)
	}

	return count, nil
}

func (c *RouteAwarenessCache) RecordLatency(ctx context.Context, apiKeyID int, latencyMs int) error {
	key := c.getLatencyKey(apiKeyID)

	pipe := c.redisClient.Pipeline()
	pipe.RPush(ctx, key, latencyMs)
	pipe.LTrim(ctx, key, -100, -1)
	pipe.Expire(ctx, key, time.Minute*5)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record latency: %w", err)
	}

	return nil
}

func (c *RouteAwarenessCache) GetLatencyStats(ctx context.Context, apiKeyID int) (p50, p95, p99 int, err error) {
	key := c.getLatencyKey(apiKeyID)

	latencies, err := c.redisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get latencies: %w", err)
	}

	if len(latencies) == 0 {
		return 0, 0, 0, nil
	}

	values := make([]int, 0, len(latencies))
	for _, l := range latencies {
		var latency int
		if _, err := fmt.Sscanf(l, "%d", &latency); err != nil {
			continue
		}
		values = append(values, latency)
	}

	if len(values) == 0 {
		return 0, 0, 0, nil
	}

	sortInts(values)

	p50 = values[len(values)*50/100]
	p95 = values[len(values)*95/100]
	p99 = values[len(values)*99/100]

	return p50, p95, p99, nil
}

func (c *RouteAwarenessCache) getStatusKey(apiKeyID int) string {
	return fmt.Sprintf("route:status:%d", apiKeyID)
}

func (c *RouteAwarenessCache) getRequestCountKey(apiKeyID int) string {
	return fmt.Sprintf("route:requests:%d", apiKeyID)
}

func (c *RouteAwarenessCache) getLatencyKey(apiKeyID int) string {
	return fmt.Sprintf("route:latency:%d", apiKeyID)
}

func sortInts(values []int) {
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}
