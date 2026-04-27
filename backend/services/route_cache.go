package services

import (
	"sync"
	"time"
)

type cachedProviderConfig struct {
	Config    providerRuntimeConfig
	ExpiredAt time.Time
}

type providerRuntimeConfig struct {
	Code           string
	Name           string
	APIBaseURL     string
	APIFormat      string
	ProviderRegion string
	RouteStrategy  map[string]interface{}
	Endpoints      map[string]interface{}
}

type RouteCache struct {
	mu      sync.RWMutex
	cache   map[string]cachedProviderConfig
	ttl     time.Duration
	enabled bool
}

func NewRouteCache(ttl time.Duration) *RouteCache {
	return &RouteCache{
		cache:   make(map[string]cachedProviderConfig),
		ttl:     ttl,
		enabled: ttl > 0,
	}
}

func (c *RouteCache) Get(providerCode string) (*providerRuntimeConfig, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if cached, ok := c.cache[providerCode]; ok {
		if time.Now().Before(cached.ExpiredAt) {
			return &cached.Config, true
		}
	}
	return nil, false
}

func (c *RouteCache) Set(providerCode string, cfg providerRuntimeConfig) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[providerCode] = cachedProviderConfig{
		Config:    cfg,
		ExpiredAt: time.Now().Add(c.ttl),
	}
}

func (c *RouteCache) Invalidate(providerCode string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, providerCode)
}

func (c *RouteCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]cachedProviderConfig)
}

func (c *RouteCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}
