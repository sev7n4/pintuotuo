package services

import (
	"testing"
	"time"
)

func TestNewConnectionPool(t *testing.T) {
	pool := NewConnectionPool(10, 100, 5*time.Minute)

	if pool == nil {
		t.Fatal("Expected connection pool, got nil")
	}

	if pool.maxIdle != 10 {
		t.Errorf("Expected maxIdle=10, got %d", pool.maxIdle)
	}

	if pool.maxTotal != 100 {
		t.Errorf("Expected maxTotal=100, got %d", pool.maxTotal)
	}

	if pool.idleTimeout != 5*time.Minute {
		t.Errorf("Expected idleTimeout=5m, got %v", pool.idleTimeout)
	}
}

func TestGetConnectionPool(t *testing.T) {
	pool1 := GetConnectionPool(1)
	pool2 := GetConnectionPool(1)

	if pool1 != pool2 {
		t.Error("Expected same instance for same API key ID")
	}

	pool3 := GetConnectionPool(2)
	if pool1 == pool3 {
		t.Error("Expected different instances for different API key IDs")
	}
}

func TestConnectionPoolGetClient(t *testing.T) {
	pool := NewConnectionPool(10, 100, 5*time.Minute)

	client, err := pool.GetClient(1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Error("Expected HTTP client, got nil")
	}

	client2, err := pool.GetClient(1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client != client2 {
		t.Error("Expected same client for same API key ID")
	}
}

func TestConnectionPoolReleaseClient(t *testing.T) {
	pool := NewConnectionPool(10, 100, 5*time.Minute)

	client, _ := pool.GetClient(1)
	if client == nil {
		t.Fatal("Expected client, got nil")
	}

	pool.ReleaseClient(1)

	stats := pool.Stats()
	if stats["in_use"].(int) != 0 {
		t.Error("Expected no clients in use after release")
	}
}

func TestConnectionPoolStats(t *testing.T) {
	pool := NewConnectionPool(10, 100, 5*time.Minute)

	pool.GetClient(1)
	pool.GetClient(2)
	pool.GetClient(3)

	stats := pool.Stats()

	if stats["total_connections"].(int) != 3 {
		t.Errorf("Expected 3 total connections, got %d", stats["total_connections"])
	}

	if stats["max_idle"].(int) != 10 {
		t.Errorf("Expected max_idle=10, got %d", stats["max_idle"])
	}

	if stats["max_total"].(int) != 100 {
		t.Errorf("Expected max_total=100, got %d", stats["max_total"])
	}
}

func TestConnectionPoolCleanupIdle(t *testing.T) {
	pool := NewConnectionPool(10, 100, 100*time.Millisecond)

	pool.GetClient(1)
	pool.GetClient(2)

	stats := pool.Stats()
	if stats["total_connections"].(int) != 2 {
		t.Errorf("Expected 2 connections, got %d", stats["total_connections"])
	}

	time.Sleep(150 * time.Millisecond)

	pool.CleanupIdle()

	stats = pool.Stats()
	if stats["total_connections"].(int) != 0 {
		t.Errorf("Expected 0 connections after cleanup, got %d", stats["total_connections"])
	}
}

func TestConnectionPoolExhaustion(t *testing.T) {
	pool := NewConnectionPool(2, 2, 5*time.Minute)

	_, err1 := pool.GetClient(1)
	_, err2 := pool.GetClient(2)

	if err1 != nil || err2 != nil {
		t.Fatal("Expected no error for first two clients")
	}

	_, err3 := pool.GetClient(3)
	if err3 == nil {
		t.Error("Expected error when pool is exhausted")
	}
}
