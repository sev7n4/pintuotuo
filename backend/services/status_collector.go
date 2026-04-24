package services

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/pintuotuo/backend/models"
)

type StatusCollector struct {
	db            *sql.DB
	awareness     IRouteAwareness
	interval      time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	healthChecker *HealthChecker
}

func NewStatusCollector(db *sql.DB, awareness IRouteAwareness, healthChecker *HealthChecker, interval time.Duration) *StatusCollector {
	return &StatusCollector{
		db:            db,
		awareness:     awareness,
		interval:      interval,
		stopChan:      make(chan struct{}),
		healthChecker: healthChecker,
	}
}

func (c *StatusCollector) Start() {
	c.wg.Add(1)
	go c.run()
	log.Println("StatusCollector started with interval:", c.interval)
}

func (c *StatusCollector) Stop() {
	close(c.stopChan)
	c.wg.Wait()
	log.Println("StatusCollector stopped")
}

func (c *StatusCollector) CollectOnce() {
	go c.collect()
}

func (c *StatusCollector) run() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.collect()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

func (c *StatusCollector) collect() {
	ctx := context.Background()

	apiKeys, err := c.getActiveAPIKeys(ctx)
	if err != nil {
		log.Printf("Failed to get active API keys: %v", err)
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, apiKey := range apiKeys {
		wg.Add(1)
		go func(key *models.MerchantAPIKey) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			status, err := c.collectStatusForAPIKey(ctx, key)
			if err != nil {
				log.Printf("Failed to collect status for API key %d: %v", key.ID, err)
				return
			}

			if err := c.awareness.UpdateStatus(ctx, key.ID, status); err != nil {
				log.Printf("Failed to update status for API key %d: %v", key.ID, err)
			}
		}(apiKey)
	}

	wg.Wait()
}

func (c *StatusCollector) getActiveAPIKeys(ctx context.Context) ([]*models.MerchantAPIKey, error) {
	query := `
		SELECT id, merchant_id, name, provider, status, region, security_level
		FROM merchant_api_keys
		WHERE status = 'active'
		ORDER BY id
	`

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []*models.MerchantAPIKey
	for rows.Next() {
		var key models.MerchantAPIKey
		err := rows.Scan(
			&key.ID,
			&key.MerchantID,
			&key.Name,
			&key.Provider,
			&key.Status,
			&key.Region,
			&key.SecurityLevel,
		)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, &key)
	}

	return apiKeys, nil
}

func (c *StatusCollector) collectStatusForAPIKey(ctx context.Context, apiKey *models.MerchantAPIKey) (*models.APIKeyRealtimeStatus, error) {
	currentStatus, err := c.awareness.GetRealtimeStatus(ctx, apiKey.ID)
	if err != nil {
		return nil, err
	}

	latency, err := c.measureLatency(ctx, apiKey)
	if err != nil {
		log.Printf("Failed to measure latency for API key %d: %v", apiKey.ID, err)
		latency = currentStatus.LatencyP50
	}

	errorRate := c.calculateErrorRate(ctx, apiKey)
	successRate := 1.0 - errorRate

	connectionPoolActive := c.getConnectionPoolActive(apiKey.ID)

	rateLimitRemaining, rateLimitResetAt := c.getRateLimitInfo(apiKey.ID)

	loadBalanceWeight := c.calculateLoadBalanceWeight(currentStatus)

	now := time.Now()
	status := &models.APIKeyRealtimeStatus{
		APIKeyID:             apiKey.ID,
		LatencyP50:           latency,
		LatencyP95:           int(float64(latency) * 1.5),
		LatencyP99:           int(float64(latency) * 2.0),
		ErrorRate:            errorRate,
		SuccessRate:          successRate,
		ConnectionPoolSize:   currentStatus.ConnectionPoolSize,
		ConnectionPoolActive: connectionPoolActive,
		RateLimitRemaining:   rateLimitRemaining,
		RateLimitResetAt:     rateLimitResetAt,
		LoadBalanceWeight:    loadBalanceWeight,
		LastRequestAt:        &now,
		UpdatedAt:            now,
	}

	return status, nil
}

func (c *StatusCollector) measureLatency(ctx context.Context, apiKey *models.MerchantAPIKey) (int, error) {
	if c.healthChecker == nil {
		return 100, nil
	}

	start := time.Now()

	_, err := c.healthChecker.LightweightPing(ctx, apiKey)
	if err != nil {
		return 0, err
	}

	latency := int(time.Since(start).Milliseconds())
	return latency, nil
}

func (c *StatusCollector) calculateErrorRate(ctx context.Context, apiKey *models.MerchantAPIKey) float64 {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN decision_result = 'failed' THEN 1 ELSE 0 END), 0)::FLOAT / 
			NULLIF(COUNT(*), 0) as error_rate
		FROM routing_decision_logs
		WHERE api_key_id = $1
		  AND created_at > NOW() - INTERVAL '5 minutes'
	`

	var errorRate float64
	err := c.db.QueryRowContext(ctx, query, apiKey.ID).Scan(&errorRate)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0.0
		}
		return 0.0
	}

	return errorRate
}

func (c *StatusCollector) getConnectionPoolActive(apiKeyID int) int {
	return 0
}

func (c *StatusCollector) getRateLimitInfo(apiKeyID int) (int, *time.Time) {
	return 0, nil
}

func (c *StatusCollector) calculateLoadBalanceWeight(currentStatus *models.APIKeyRealtimeStatus) float64 {
	if currentStatus.ErrorRate > 0.5 {
		return 0.0
	}

	if currentStatus.ErrorRate > 0.2 {
		return 0.3
	}

	if currentStatus.ErrorRate > 0.1 {
		return 0.6
	}

	latencyScore := 1.0
	if currentStatus.LatencyP50 > 1000 {
		latencyScore = 0.5
	} else if currentStatus.LatencyP50 > 500 {
		latencyScore = 0.7
	} else if currentStatus.LatencyP50 > 200 {
		latencyScore = 0.9
	}

	errorScore := 1.0 - currentStatus.ErrorRate

	weight := (latencyScore * 0.6) + (errorScore * 0.4)

	return weight
}
