package services

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ProviderClient struct {
	Client   *http.Client
	LastUsed time.Time
	InUse    bool
}

type ConnectionPool struct {
	mu          sync.RWMutex
	clients     map[int]*ProviderClient
	maxIdle     int
	maxTotal    int
	idleTimeout time.Duration
}

var (
	connectionPools = make(map[int]*ConnectionPool)
	poolMutex       sync.RWMutex
)

func NewConnectionPool(maxIdle, maxTotal int, idleTimeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		clients:     make(map[int]*ProviderClient),
		maxIdle:     maxIdle,
		maxTotal:    maxTotal,
		idleTimeout: idleTimeout,
	}
}

func GetConnectionPool(apiKeyID int) *ConnectionPool {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	if pool, exists := connectionPools[apiKeyID]; exists {
		return pool
	}

	pool := NewConnectionPool(10, 100, 5*time.Minute)
	connectionPools[apiKeyID] = pool
	return pool
}

func (p *ConnectionPool) GetClient(apiKeyID int) (*http.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists := p.clients[apiKeyID]; exists {
		client.LastUsed = time.Now()
		return client.Client, nil
	}

	if len(p.clients) >= p.maxTotal {
		return nil, fmt.Errorf("connection pool exhausted")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
			DisableCompression:    true,
			MaxConnsPerHost:       10,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}

	p.clients[apiKeyID] = &ProviderClient{
		Client:   client,
		LastUsed: time.Now(),
		InUse:    false,
	}

	return client, nil
}

func (p *ConnectionPool) ReleaseClient(apiKeyID int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, exists := p.clients[apiKeyID]; exists {
		client.InUse = false
		client.LastUsed = time.Now()
	}
}

func (p *ConnectionPool) CleanupIdle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for id, client := range p.clients {
		if !client.InUse && now.Sub(client.LastUsed) > p.idleTimeout {
			delete(p.clients, id)
		}
	}
}

func (p *ConnectionPool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := len(p.clients)
	inUse := 0
	for _, client := range p.clients {
		if client.InUse {
			inUse++
		}
	}

	return map[string]interface{}{
		"total_connections": total,
		"in_use":            inUse,
		"idle":              total - inUse,
		"max_idle":          p.maxIdle,
		"max_total":         p.maxTotal,
	}
}
