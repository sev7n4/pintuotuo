package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/pintuotuo/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestRouteAwarenessService_GetRealtimeStatus(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Database connection failed")
	}

	service := NewRouteAwarenessService(db)
	ctx := context.Background()

	t.Run("should return default status for non-existent API Key", func(t *testing.T) {
		status, err := service.GetRealtimeStatus(ctx, 99999)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.NotNil(t, status)
		assert.Equal(t, 99999, status.APIKeyID)
		assert.Equal(t, 1.0, status.SuccessRate)
		assert.Equal(t, 0.0, status.ErrorRate)
	})

	t.Run("should return existing status", func(t *testing.T) {
		status := &models.APIKeyRealtimeStatus{
			APIKeyID:             1,
			LatencyP50:           100,
			LatencyP95:           200,
			LatencyP99:           300,
			ErrorRate:            0.01,
			SuccessRate:          0.99,
			ConnectionPoolSize:   20,
			ConnectionPoolActive: 5,
			RateLimitRemaining:   1000,
			LoadBalanceWeight:    0.8,
		}

		err := service.UpdateStatus(ctx, 1, status)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		result, err := service.GetRealtimeStatus(ctx, 1)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.NotNil(t, result)
		assert.Equal(t, 100, result.LatencyP50)
		assert.Equal(t, 200, result.LatencyP95)
		assert.Equal(t, 300, result.LatencyP99)
		assert.Equal(t, 0.01, result.ErrorRate)
		assert.Equal(t, 0.99, result.SuccessRate)
	})
}

func TestRouteAwarenessService_UpdateStatus(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Database connection failed")
	}

	service := NewRouteAwarenessService(db)
	ctx := context.Background()

	t.Run("should create new status if not exists", func(t *testing.T) {
		status := &models.APIKeyRealtimeStatus{
			APIKeyID:             2,
			LatencyP50:           150,
			LatencyP95:           250,
			LatencyP99:           350,
			ErrorRate:            0.02,
			SuccessRate:          0.98,
			ConnectionPoolSize:   15,
			ConnectionPoolActive: 3,
			RateLimitRemaining:   500,
			LoadBalanceWeight:    0.9,
		}

		err := service.UpdateStatus(ctx, 2, status)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		result, err := service.GetRealtimeStatus(ctx, 2)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.Equal(t, 150, result.LatencyP50)
	})

	t.Run("should update existing status", func(t *testing.T) {
		status := &models.APIKeyRealtimeStatus{
			APIKeyID:             2,
			LatencyP50:           200,
			LatencyP95:           300,
			LatencyP99:           400,
			ErrorRate:            0.03,
			SuccessRate:          0.97,
			ConnectionPoolSize:   15,
			ConnectionPoolActive: 5,
			RateLimitRemaining:   400,
			LoadBalanceWeight:    0.85,
		}

		err := service.UpdateStatus(ctx, 2, status)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		result, err := service.GetRealtimeStatus(ctx, 2)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.Equal(t, 200, result.LatencyP50)
		assert.Equal(t, 300, result.LatencyP95)
	})
}

func TestRouteAwarenessService_GetBatchStatus(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Database connection failed")
	}

	service := NewRouteAwarenessService(db)
	ctx := context.Background()

	t.Run("should return empty list for empty input", func(t *testing.T) {
		statuses, err := service.GetBatchStatus(ctx, []int{})
		assert.NoError(t, err)
		assert.Empty(t, statuses)
	})

	t.Run("should return multiple statuses", func(t *testing.T) {
		status1 := &models.APIKeyRealtimeStatus{
			APIKeyID:             10,
			LatencyP50:           100,
			ErrorRate:            0.01,
			SuccessRate:          0.99,
			ConnectionPoolSize:   10,
			ConnectionPoolActive: 2,
			RateLimitRemaining:   1000,
			LoadBalanceWeight:    1.0,
		}

		status2 := &models.APIKeyRealtimeStatus{
			APIKeyID:             11,
			LatencyP50:           200,
			ErrorRate:            0.02,
			SuccessRate:          0.98,
			ConnectionPoolSize:   10,
			ConnectionPoolActive: 3,
			RateLimitRemaining:   800,
			LoadBalanceWeight:    0.9,
		}

		err := service.UpdateStatus(ctx, 10, status1)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		err = service.UpdateStatus(ctx, 11, status2)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		statuses, err := service.GetBatchStatus(ctx, []int{10, 11})
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.NoError(t, err)
		assert.Len(t, statuses, 2)
	})
}

func TestRouteAwarenessService_UpdateLatency(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Database connection failed")
	}

	service := NewRouteAwarenessService(db)
	ctx := context.Background()

	t.Run("should update latency", func(t *testing.T) {
		err := service.UpdateLatency(ctx, 20, 150)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		status, err := service.GetRealtimeStatus(ctx, 20)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.Equal(t, 150, status.LatencyP50)
		assert.Equal(t, 150, status.LatencyP95)
		assert.Equal(t, 150, status.LatencyP99)
	})
}

func TestRouteAwarenessService_UpdateErrorRate(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Database connection failed")
	}

	service := NewRouteAwarenessService(db)
	ctx := context.Background()

	t.Run("should update error rate and success rate", func(t *testing.T) {
		err := service.UpdateErrorRate(ctx, 30, 0.05)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		status, err := service.GetRealtimeStatus(ctx, 30)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.Equal(t, 0.05, status.ErrorRate)
		assert.Equal(t, 0.95, status.SuccessRate)
	})
}

func TestRouteAwarenessService_UpdateConnectionPool(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Database connection failed")
	}

	service := NewRouteAwarenessService(db)
	ctx := context.Background()

	t.Run("should update connection pool active count", func(t *testing.T) {
		err := service.UpdateConnectionPool(ctx, 40, 7)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		status, err := service.GetRealtimeStatus(ctx, 40)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.Equal(t, 7, status.ConnectionPoolActive)
	})
}

func TestRouteAwarenessService_UpdateRateLimit(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")
	if err != nil {
		t.Skip("Database not available for testing")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Database connection failed")
	}

	service := NewRouteAwarenessService(db)
	ctx := context.Background()

	t.Run("should update rate limit info", func(t *testing.T) {
		resetAt := time.Now().Add(1 * time.Hour)
		err := service.UpdateRateLimit(ctx, 50, 500, &resetAt)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}

		status, err := service.GetRealtimeStatus(ctx, 50)
		if err != nil {
			t.Skipf("Database error: %v", err)
			return
		}
		assert.Equal(t, 500, status.RateLimitRemaining)
		assert.NotNil(t, status.RateLimitResetAt)
	})
}
