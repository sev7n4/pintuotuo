package services

import (
	"context"
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) bool
	SetRate(rate int, burst int)
	GetRate() (int, int)
	GetStats() map[string]interface{}
}

type TokenBucket struct {
	mu           sync.Mutex
	rate         int       // 令牌生成速率（每秒）
	burst        int       // 令牌桶容量
	tokens       float64   // 当前令牌数
	lastRefilled time.Time // 上次填充时间
	ratePerNano  float64   // 每纳秒生成的令牌数
	stats        *RateLimitStats
}

type RateLimitStats struct {
	Requests  int64     `json:"requests"`
	Allowed   int64     `json:"allowed"`
	Denied    int64     `json:"denied"`
	LastReset time.Time `json:"last_reset"`
	mu        sync.Mutex
}

type RateLimiterFactory struct {
	rateLimiters map[string]*TokenBucket
	mu           sync.RWMutex
}

var rateLimiterFactory *RateLimiterFactory
var rateLimiterOnce sync.Once

func GetRateLimiter() *RateLimiterFactory {
	rateLimiterOnce.Do(func() {
		rateLimiterFactory = &RateLimiterFactory{
			rateLimiters: make(map[string]*TokenBucket),
		}
	})
	return rateLimiterFactory
}

func NewTokenBucket(rate, burst int) *TokenBucket {
	return &TokenBucket{
		rate:         rate,
		burst:        burst,
		tokens:       float64(burst),
		lastRefilled: time.Now(),
		ratePerNano:  float64(rate) / 1e9,
		stats: &RateLimitStats{
			LastReset: time.Now(),
		},
	}
}

func (tb *TokenBucket) refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	duration := now.Sub(tb.lastRefilled)
	nanoseconds := float64(duration.Nanoseconds())

	// 计算应该生成的令牌数
	newTokens := nanoseconds * tb.ratePerNano
	tb.tokens += newTokens

	// 确保令牌数不超过桶容量
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}

	tb.lastRefilled = now
}

func (tb *TokenBucket) Allow() bool {
	tb.refill()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.stats.mu.Lock()
	tb.stats.Requests++
	tb.stats.mu.Unlock()

	if tb.tokens >= 1.0 {
		tb.tokens--
		tb.stats.mu.Lock()
		tb.stats.Allowed++
		tb.stats.mu.Unlock()
		return true
	}

	tb.stats.mu.Lock()
	tb.stats.Denied++
	tb.stats.mu.Unlock()
	return false
}

func (tb *TokenBucket) SetRate(rate, burst int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.rate = rate
	tb.burst = burst
	tb.ratePerNano = float64(rate) / 1e9

	// 确保令牌数不超过新的桶容量
	if tb.tokens > float64(burst) {
		tb.tokens = float64(burst)
	}
}

func (tb *TokenBucket) GetRate() (int, int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.rate, tb.burst
}

func (tb *TokenBucket) GetStats() map[string]interface{} {
	tb.stats.mu.Lock()
	defer tb.stats.mu.Unlock()

	return map[string]interface{}{
		"requests":    tb.stats.Requests,
		"allowed":     tb.stats.Allowed,
		"denied":      tb.stats.Denied,
		"denied_rate": float64(tb.stats.Denied) / float64(tb.stats.Requests+1),
		"last_reset":  tb.stats.LastReset,
	}
}

func (tb *TokenBucket) ResetStats() {
	tb.stats.mu.Lock()
	defer tb.stats.mu.Unlock()

	tb.stats.Requests = 0
	tb.stats.Allowed = 0
	tb.stats.Denied = 0
	tb.stats.LastReset = time.Now()
}

func (f *RateLimiterFactory) GetLimiter(key string, rate, burst int) *TokenBucket {
	f.mu.Lock()
	defer f.mu.Unlock()

	limiter, exists := f.rateLimiters[key]
	if !exists {
		limiter = NewTokenBucket(rate, burst)
		f.rateLimiters[key] = limiter
	}

	return limiter
}

func (f *RateLimiterFactory) Allow(key string, rate, burst int) bool {
	limiter := f.GetLimiter(key, rate, burst)
	return limiter.Allow()
}

func (f *RateLimiterFactory) SetRate(key string, rate, burst int) {
	limiter := f.GetLimiter(key, rate, burst)
	limiter.SetRate(rate, burst)
}

func (f *RateLimiterFactory) GetStats(key string) map[string]interface{} {
	f.mu.RLock()
	limiter, exists := f.rateLimiters[key]
	f.mu.RUnlock()

	if !exists {
		return map[string]interface{}{
			"error": "limiter not found",
		}
	}

	return limiter.GetStats()
}

func (f *RateLimiterFactory) GetAllStats() map[string]map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for key, limiter := range f.rateLimiters {
		stats[key] = limiter.GetStats()
	}

	return stats
}

func (f *RateLimiterFactory) ResetStats(key string) {
	f.mu.RLock()
	limiter, exists := f.rateLimiters[key]
	f.mu.RUnlock()

	if exists {
		limiter.ResetStats()
	}
}

func (f *RateLimiterFactory) ResetAllStats() {
	f.mu.RLock()
	limiters := make([]*TokenBucket, 0, len(f.rateLimiters))
	for _, limiter := range f.rateLimiters {
		limiters = append(limiters, limiter)
	}
	f.mu.RUnlock()

	for _, limiter := range limiters {
		limiter.ResetStats()
	}
}

func (f *RateLimiterFactory) RemoveLimiter(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.rateLimiters, key)
}

func (f *RateLimiterFactory) GetLimiterCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.rateLimiters)
}
