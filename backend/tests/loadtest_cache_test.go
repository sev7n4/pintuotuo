package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoadTestMetrics tracks performance and cache statistics
type LoadTestMetrics struct {
	TotalRequests      int64
	CacheHits          int64
	CacheMisses        int64
	DatabaseQueries    int64
	TotalResponseTime  time.Duration
	MinResponseTime    time.Duration
	MaxResponseTime    time.Duration
	Errors             int64
	StartTime          time.Time
	EndTime            time.Time
}

// CacheHitRatio calculates the percentage of cache hits
func (m *LoadTestMetrics) CacheHitRatio() float64 {
	if m.TotalRequests == 0 {
		return 0
	}
	return float64(m.CacheHits) / float64(m.TotalRequests) * 100
}

// AverageResponseTime calculates mean response time
func (m *LoadTestMetrics) AverageResponseTime() time.Duration {
	if m.TotalRequests == 0 {
		return 0
	}
	return m.TotalResponseTime / time.Duration(m.TotalRequests)
}

// ThroughputPerSecond calculates requests per second
func (m *LoadTestMetrics) ThroughputPerSecond() float64 {
	duration := m.EndTime.Sub(m.StartTime).Seconds()
	if duration == 0 {
		return 0
	}
	return float64(m.TotalRequests) / duration
}

// ProductRequest represents a single product request for load testing
type ProductRequest struct {
	ID        string
	Timestamp time.Time
}

// ProductResponse represents a cached product response
type ProductResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        float64 `json:"price"`
	Stock        int    `json:"stock"`
	Status       string `json:"status"`
	CachedAt     time.Time
	IsCachHit    bool
}

// SimulateProductCache simulates cache behavior for load testing
type SimulateProductCache struct {
	cache      map[string]string
	mu         sync.RWMutex
	hitCount   int64
	missCount  int64
	queryCount int64
}

// NewSimulateProductCache creates a new simulated cache
func NewSimulateProductCache() *SimulateProductCache {
	return &SimulateProductCache{
		cache: make(map[string]string),
	}
}

// Get simulates cache retrieval
func (pc *SimulateProductCache) Get(key string) (string, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	value, exists := pc.cache[key]
	if exists {
		atomic.AddInt64(&pc.hitCount, 1)
		return value, true
	}

	atomic.AddInt64(&pc.missCount, 1)
	return "", false
}

// Set simulates cache storage
func (pc *SimulateProductCache) Set(key, value string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cache[key] = value
}

// SimulateDatabase simulates database query time
func (pc *SimulateProductCache) SimulateDatabase(id string) string {
	atomic.AddInt64(&pc.queryCount, 1)
	// Simulate 20-50ms database query time
	time.Sleep(time.Duration(20+id[len(id)-1]%30) * time.Millisecond)

	product := ProductResponse{
		ID:          int(id[len(id)-1] % 100),
		Name:        fmt.Sprintf("Product-%s", id),
		Description: fmt.Sprintf("Description for product %s", id),
		Price:       99.99,
		Stock:       100,
		Status:      "active",
		IsCachHit:   false,
	}

	data, _ := json.Marshal(product)
	return string(data)
}

// TestCacheLookupColdCache tests performance with empty cache
func TestCacheLookupColdCache(t *testing.T) {
	t.Run("Cold cache - all requests hit database", func(t *testing.T) {
		cache := NewSimulateProductCache()
		metrics := &LoadTestMetrics{
			StartTime: time.Now(),
			MinResponseTime: time.Duration(1 << 63 - 1),
		}

		// 100 requests for 100 different products (no cache hits expected)
		for i := 0; i < 100; i++ {
			id := fmt.Sprintf("product-%d", i)
			start := time.Now()

			// Cache miss
			if _, hit := cache.Get(id); !hit {
				// Simulate database query
				data := cache.SimulateDatabase(id)
				cache.Set(id, data)
				metrics.CacheMisses++
			} else {
				metrics.CacheHits++
			}

			duration := time.Since(start)
			metrics.TotalResponseTime += duration
			metrics.TotalRequests++

			if duration < metrics.MinResponseTime {
				metrics.MinResponseTime = duration
			}
			if duration > metrics.MaxResponseTime {
				metrics.MaxResponseTime = duration
			}
		}

		metrics.EndTime = time.Now()

		// Verify all requests were cache misses
		if metrics.CacheHitRatio() != 0 {
			t.Errorf("Expected 0%% cache hit ratio for cold cache, got %.2f%%", metrics.CacheHitRatio())
		}

		t.Logf("Cold Cache Results:")
		t.Logf("  Total Requests: %d", metrics.TotalRequests)
		t.Logf("  Cache Hits: %d", metrics.CacheHits)
		t.Logf("  Cache Misses: %d", metrics.CacheMisses)
		t.Logf("  Hit Ratio: %.2f%%", metrics.CacheHitRatio())
		t.Logf("  Avg Response Time: %v", metrics.AverageResponseTime())
		t.Logf("  Min Response Time: %v", metrics.MinResponseTime)
		t.Logf("  Max Response Time: %v", metrics.MaxResponseTime)
		t.Logf("  Throughput: %.2f req/sec", metrics.ThroughputPerSecond())
	})
}

// TestCacheLookupWarmCache tests performance with populated cache
func TestCacheLookupWarmCache(t *testing.T) {
	t.Run("Warm cache - repeated requests hit cache", func(t *testing.T) {
		cache := NewSimulateProductCache()
		metrics := &LoadTestMetrics{
			StartTime: time.Now(),
			MinResponseTime: time.Duration(1 << 63 - 1),
		}

		// Warm up: populate cache with 20 products
		for i := 0; i < 20; i++ {
			id := fmt.Sprintf("product-%d", i)
			data := cache.SimulateDatabase(id)
			cache.Set(id, data)
		}

		// Load test: 1000 requests for these 20 products (repeated)
		for i := 0; i < 1000; i++ {
			id := fmt.Sprintf("product-%d", i%20)
			start := time.Now()

			if value, hit := cache.Get(id); hit {
				// Cache hit - minimal processing
				_ = value
				metrics.CacheHits++
			} else {
				// Cache miss - simulate database query
				data := cache.SimulateDatabase(id)
				cache.Set(id, data)
				metrics.CacheMisses++
			}

			duration := time.Since(start)
			metrics.TotalResponseTime += duration
			metrics.TotalRequests++

			if duration < metrics.MinResponseTime {
				metrics.MinResponseTime = duration
			}
			if duration > metrics.MaxResponseTime {
				metrics.MaxResponseTime = duration
			}
		}

		metrics.EndTime = time.Now()

		// Verify cache hit ratio is high
		hitRatio := metrics.CacheHitRatio()
		if hitRatio < 95 {
			t.Logf("Warning: Cache hit ratio %.2f%% is lower than expected 95%%", hitRatio)
		}

		t.Logf("Warm Cache Results:")
		t.Logf("  Total Requests: %d", metrics.TotalRequests)
		t.Logf("  Cache Hits: %d", metrics.CacheHits)
		t.Logf("  Cache Misses: %d", metrics.CacheMisses)
		t.Logf("  Hit Ratio: %.2f%%", hitRatio)
		t.Logf("  Avg Response Time: %v", metrics.AverageResponseTime())
		t.Logf("  Min Response Time: %v", metrics.MinResponseTime)
		t.Logf("  Max Response Time: %v", metrics.MaxResponseTime)
		t.Logf("  Throughput: %.2f req/sec", metrics.ThroughputPerSecond())
	})
}

// TestCacheHitRatioTarget70Percent tests achieving 70% hit ratio target
func TestCacheHitRatioTarget70Percent(t *testing.T) {
	t.Run("Mixed workload - target 70% cache hit ratio", func(t *testing.T) {
		cache := NewSimulateProductCache()
		metrics := &LoadTestMetrics{
			StartTime: time.Now(),
			MinResponseTime: time.Duration(1 << 63 - 1),
		}

		// Simulate 50 unique products
		numProducts := 50

		// 2000 requests with realistic distribution
		numRequests := 2000
		for i := 0; i < numRequests; i++ {
			// 70% of requests go to 20 popular products
			// 30% of requests go to other products
			var id string
			if i%10 < 7 {
				// Popular products
				id = fmt.Sprintf("product-%d", i%20)
			} else {
				// Less popular products
				id = fmt.Sprintf("product-%d", 20+(i%numProducts))
			}

			start := time.Now()

			if value, hit := cache.Get(id); hit {
				_ = value
				metrics.CacheHits++
			} else {
				data := cache.SimulateDatabase(id)
				cache.Set(id, data)
				metrics.CacheMisses++
			}

			duration := time.Since(start)
			metrics.TotalResponseTime += duration
			metrics.TotalRequests++

			if duration < metrics.MinResponseTime {
				metrics.MinResponseTime = duration
			}
			if duration > metrics.MaxResponseTime {
				metrics.MaxResponseTime = duration
			}
		}

		metrics.EndTime = time.Now()

		hitRatio := metrics.CacheHitRatio()

		// Verify hit ratio is within acceptable range (60-80%)
		if hitRatio < 60 {
			t.Errorf("Cache hit ratio %.2f%% is below acceptable minimum of 60%%", hitRatio)
		}
		if hitRatio > 80 {
			t.Logf("Cache hit ratio %.2f%% exceeds expected target of 70%%", hitRatio)
		}

		t.Logf("Target Hit Ratio (70%%) Results:")
		t.Logf("  Total Requests: %d", metrics.TotalRequests)
		t.Logf("  Cache Hits: %d", metrics.CacheHits)
		t.Logf("  Cache Misses: %d", metrics.CacheMisses)
		t.Logf("  Hit Ratio: %.2f%% (Target: 70%%)", hitRatio)
		t.Logf("  Avg Response Time: %v", metrics.AverageResponseTime())
		t.Logf("  Min Response Time: %v", metrics.MinResponseTime)
		t.Logf("  Max Response Time: %v", metrics.MaxResponseTime)
		t.Logf("  Throughput: %.2f req/sec", metrics.ThroughputPerSecond())
		t.Logf("  Database Queries: %d", metrics.DatabaseQueries)
	})
}

// TestConcurrentCacheAccess tests cache performance under concurrent load
func TestConcurrentCacheAccess(t *testing.T) {
	t.Run("Concurrent requests with 100 goroutines", func(t *testing.T) {
		cache := NewSimulateProductCache()
		metrics := &LoadTestMetrics{
			StartTime: time.Now(),
			MinResponseTime: time.Duration(1 << 63 - 1),
		}

		// Pre-warm cache
		for i := 0; i < 30; i++ {
			id := fmt.Sprintf("product-%d", i)
			data := cache.SimulateDatabase(id)
			cache.Set(id, data)
		}

		numGoroutines := 100
		requestsPerGoroutine := 100
		var wg sync.WaitGroup
		var mutex sync.Mutex

		minTime := time.Duration(1 << 63 - 1)
		maxTime := time.Duration(0)

		start := time.Now()

		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for r := 0; r < requestsPerGoroutine; r++ {
					id := fmt.Sprintf("product-%d", (goroutineID+r)%30)

					reqStart := time.Now()

					if value, hit := cache.Get(id); hit {
						_ = value
						atomic.AddInt64(&metrics.CacheHits, 1)
					} else {
						data := cache.SimulateDatabase(id)
						cache.Set(id, data)
						atomic.AddInt64(&metrics.CacheMisses, 1)
					}

					duration := time.Since(reqStart)
					atomic.AddInt64(&metrics.TotalRequests, 1)

					mutex.Lock()
					metrics.TotalResponseTime += duration
					if duration < minTime {
						minTime = duration
					}
					if duration > maxTime {
						maxTime = duration
					}
					mutex.Unlock()
				}
			}(g)
		}

		wg.Wait()
		metrics.EndTime = time.Now()

		if minTime != time.Duration(1<<63-1) {
			metrics.MinResponseTime = minTime
		}
		metrics.MaxResponseTime = maxTime

		t.Logf("Concurrent Load Test Results:")
		t.Logf("  Goroutines: %d", numGoroutines)
		t.Logf("  Requests per Goroutine: %d", requestsPerGoroutine)
		t.Logf("  Total Requests: %d", metrics.TotalRequests)
		t.Logf("  Cache Hits: %d", metrics.CacheHits)
		t.Logf("  Cache Misses: %d", metrics.CacheMisses)
		t.Logf("  Hit Ratio: %.2f%%", metrics.CacheHitRatio())
		t.Logf("  Avg Response Time: %v", metrics.AverageResponseTime())
		t.Logf("  Min Response Time: %v", metrics.MinResponseTime)
		t.Logf("  Max Response Time: %v", metrics.MaxResponseTime)
		t.Logf("  Total Duration: %v", time.Since(start))
		t.Logf("  Throughput: %.2f req/sec", metrics.ThroughputPerSecond())

		// Verify cache hit ratio is reasonable under concurrent load
		if metrics.CacheHitRatio() < 50 {
			t.Logf("Warning: Cache hit ratio %.2f%% is lower than expected under concurrent load", metrics.CacheHitRatio())
		}
	})
}

// TestCacheInvalidationImpact tests performance impact of cache invalidation
func TestCacheInvalidationImpact(t *testing.T) {
	t.Run("Cache invalidation and refresh cycle", func(t *testing.T) {
		cache := NewSimulateProductCache()
		metrics := &LoadTestMetrics{
			StartTime: time.Now(),
			MinResponseTime: time.Duration(1 << 63 - 1),
		}

		numCycles := 5
		requestsPerCycle := 200

		for cycle := 0; cycle < numCycles; cycle++ {
			// Phase 1: Warm up cache (20 products)
			if cycle == 0 {
				for i := 0; i < 20; i++ {
					id := fmt.Sprintf("product-%d", i)
					data := cache.SimulateDatabase(id)
					cache.Set(id, data)
				}
			}

			// Phase 2: Regular requests
			for i := 0; i < requestsPerCycle; i++ {
				id := fmt.Sprintf("product-%d", i%20)

				start := time.Now()

				if value, hit := cache.Get(id); hit {
					_ = value
					metrics.CacheHits++
				} else {
					data := cache.SimulateDatabase(id)
					cache.Set(id, data)
					metrics.CacheMisses++
				}

				duration := time.Since(start)
				metrics.TotalResponseTime += duration
				metrics.TotalRequests++

				if duration < metrics.MinResponseTime {
					metrics.MinResponseTime = duration
				}
				if duration > metrics.MaxResponseTime {
					metrics.MaxResponseTime = duration
				}
			}

			// Phase 3: Cache invalidation (simulate product update)
			if cycle < numCycles-1 {
				// Clear 5 products from cache
				cache.mu.Lock()
				for i := 0; i < 5; i++ {
					id := fmt.Sprintf("product-%d", cycle*5+i)
					delete(cache.cache, id)
				}
				cache.mu.Unlock()
			}
		}

		metrics.EndTime = time.Now()

		t.Logf("Cache Invalidation Impact Results:")
		t.Logf("  Total Requests: %d", metrics.TotalRequests)
		t.Logf("  Cache Hits: %d", metrics.CacheHits)
		t.Logf("  Cache Misses: %d", metrics.CacheMisses)
		t.Logf("  Hit Ratio: %.2f%%", metrics.CacheHitRatio())
		t.Logf("  Avg Response Time: %v", metrics.AverageResponseTime())
		t.Logf("  Min Response Time: %v", metrics.MinResponseTime)
		t.Logf("  Max Response Time: %v", metrics.MaxResponseTime)
		t.Logf("  Throughput: %.2f req/sec", metrics.ThroughputPerSecond())
		t.Logf("  Cycles: %d", numCycles)
	})
}

// TestRealRedisCache tests actual Redis cache performance (optional)
func TestRealRedisCache(t *testing.T) {
	t.Run("Real Redis cache performance", func(t *testing.T) {
		// Skip test if Redis is not available
		rdb := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			t.Skip("Redis not available, skipping real cache test")
		}

		// Clear test keys
		rdb.FlushDB(ctx)

		metrics := &LoadTestMetrics{
			StartTime: time.Now(),
			MinResponseTime: time.Duration(1 << 63 - 1),
		}

		// 1000 requests with mixed hit/miss
		for i := 0; i < 1000; i++ {
			id := fmt.Sprintf("test-product-%d", i%100)

			start := time.Now()

			// Try to get from cache
			val, err := rdb.Get(ctx, id).Result()

			if err == redis.Nil {
				// Cache miss - simulate database query
				time.Sleep(30 * time.Millisecond)

				product := ProductResponse{
					ID:    i % 100,
					Name:  fmt.Sprintf("Product-%d", i%100),
					Price: 99.99,
					Stock: 100,
				}

				data, _ := json.Marshal(product)
				rdb.Set(ctx, id, string(data), 5*time.Minute)

				metrics.CacheMisses++
			} else if err != nil {
				metrics.Errors++
			} else {
				// Cache hit
				_ = val
				metrics.CacheHits++
			}

			duration := time.Since(start)
			metrics.TotalResponseTime += duration
			metrics.TotalRequests++

			if duration < metrics.MinResponseTime {
				metrics.MinResponseTime = duration
			}
			if duration > metrics.MaxResponseTime {
				metrics.MaxResponseTime = duration
			}
		}

		metrics.EndTime = time.Now()

		t.Logf("Real Redis Cache Results:")
		t.Logf("  Total Requests: %d", metrics.TotalRequests)
		t.Logf("  Cache Hits: %d", metrics.CacheHits)
		t.Logf("  Cache Misses: %d", metrics.CacheMisses)
		t.Logf("  Hit Ratio: %.2f%%", metrics.CacheHitRatio())
		t.Logf("  Avg Response Time: %v", metrics.AverageResponseTime())
		t.Logf("  Min Response Time: %v", metrics.MinResponseTime)
		t.Logf("  Max Response Time: %v", metrics.MaxResponseTime)
		t.Logf("  Throughput: %.2f req/sec", metrics.ThroughputPerSecond())
		t.Logf("  Errors: %d", metrics.Errors)

		// Cleanup
		rdb.FlushDB(ctx)
		rdb.Close()
	})
}

// BenchmarkCacheLookup benchmarks cache lookup performance
func BenchmarkCacheLookup(b *testing.B) {
	cache := NewSimulateProductCache()

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("product-%d", i)
		cache.Set(id, fmt.Sprintf(`{"id":%d,"name":"Product %d"}`, i, i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("product-%d", i%100)
		_, _ = cache.Get(id)
	}
}

// BenchmarkDatabaseQuery benchmarks database query simulation
func BenchmarkDatabaseQuery(b *testing.B) {
	cache := NewSimulateProductCache()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("product-%d", i%100)
		_ = cache.SimulateDatabase(id)
	}
}
