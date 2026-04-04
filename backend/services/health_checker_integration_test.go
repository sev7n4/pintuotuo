package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	connStr := "postgres://postgres:postgres@localhost:5432/pintuotuo_test?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Failed to connect to test database: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Test database not available: %v", err)
		return nil
	}

	return db
}

func TestRecordRequestResult_Success(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	checker := &HealthChecker{db: db}
	ctx := context.Background()

	var apiKeyID int
	err := db.QueryRow(`
		INSERT INTO merchant_api_keys (merchant_id, provider, api_key_encrypted, status)
		VALUES (1, 'openai', 'test-key', 'active')
		RETURNING id
	`).Scan(&apiKeyID)
	if err != nil {
		t.Fatalf("Failed to create test api key: %v", err)
	}
	defer db.Exec("DELETE FROM merchant_api_keys WHERE id = $1", apiKeyID)

	err = checker.RecordRequestResult(ctx, apiKeyID, true, 100)
	if err != nil {
		t.Errorf("RecordRequestResult failed: %v", err)
	}

	var consecutiveFailures int
	var healthStatus string
	err = db.QueryRow(`
		SELECT consecutive_failures, health_status 
		FROM merchant_api_keys WHERE id = $1
	`, apiKeyID).Scan(&consecutiveFailures, &healthStatus)

	if err != nil {
		t.Errorf("Failed to query result: %v", err)
	}

	if consecutiveFailures != 0 {
		t.Errorf("Expected consecutive_failures=0, got %d", consecutiveFailures)
	}

	if healthStatus != "healthy" {
		t.Errorf("Expected health_status=healthy, got %s", healthStatus)
	}
}

func TestRecordRequestResult_Failure(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	checker := &HealthChecker{db: db}
	ctx := context.Background()

	var apiKeyID int
	err := db.QueryRow(`
		INSERT INTO merchant_api_keys (merchant_id, provider, api_key_encrypted, status, consecutive_failures)
		VALUES (1, 'openai', 'test-key', 'active', 0)
		RETURNING id
	`).Scan(&apiKeyID)
	if err != nil {
		t.Fatalf("Failed to create test api key: %v", err)
	}
	defer db.Exec("DELETE FROM merchant_api_keys WHERE id = $1", apiKeyID)

	for i := 0; i < 5; i++ {
		err = checker.RecordRequestResult(ctx, apiKeyID, false, 0)
		if err != nil {
			t.Errorf("RecordRequestResult failed at iteration %d: %v", i, err)
		}
	}

	var consecutiveFailures int
	var healthStatus string
	err = db.QueryRow(`
		SELECT consecutive_failures, health_status 
		FROM merchant_api_keys WHERE id = $1
	`, apiKeyID).Scan(&consecutiveFailures, &healthStatus)

	if err != nil {
		t.Errorf("Failed to query result: %v", err)
	}

	if consecutiveFailures != 5 {
		t.Errorf("Expected consecutive_failures=5, got %d", consecutiveFailures)
	}

	if healthStatus != "unhealthy" {
		t.Errorf("Expected health_status=unhealthy, got %s", healthStatus)
	}
}

func TestCalculateFailureRate(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	checker := &HealthChecker{db: db}
	ctx := context.Background()

	var apiKeyID int
	err := db.QueryRow(`
		INSERT INTO merchant_api_keys (merchant_id, provider, api_key_encrypted, status)
		VALUES (1, 'openai', 'test-key', 'active')
		RETURNING id
	`).Scan(&apiKeyID)
	if err != nil {
		t.Fatalf("Failed to create test api key: %v", err)
	}
	defer db.Exec("DELETE FROM merchant_api_keys WHERE id = $1", apiKeyID)

	_, err = db.Exec(`
		INSERT INTO api_key_health_history (api_key_id, check_type, status, latency_ms)
		VALUES 
			($1, 'passive', 'healthy', 100),
			($1, 'passive', 'healthy', 100),
			($1, 'passive', 'unhealthy', 0),
			($1, 'passive', 'healthy', 100)
	`, apiKeyID)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	rate, err := checker.CalculateFailureRate(ctx, apiKeyID, 60)
	if err != nil {
		t.Errorf("CalculateFailureRate failed: %v", err)
	}

	expectedRate := 25.0
	if rate != expectedRate {
		t.Errorf("Expected failure rate=%.2f, got %.2f", expectedRate, rate)
	}
}

func TestMarkAsDegraded(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	checker := &HealthChecker{db: db}
	ctx := context.Background()

	var apiKeyID int
	err := db.QueryRow(`
		INSERT INTO merchant_api_keys (merchant_id, provider, api_key_encrypted, status, health_status)
		VALUES (1, 'openai', 'test-key', 'active', 'healthy')
		RETURNING id
	`).Scan(&apiKeyID)
	if err != nil {
		t.Fatalf("Failed to create test api key: %v", err)
	}
	defer db.Exec("DELETE FROM merchant_api_keys WHERE id = $1", apiKeyID)

	err = checker.MarkAsDegraded(ctx, apiKeyID)
	if err != nil {
		t.Errorf("MarkAsDegraded failed: %v", err)
	}

	var healthStatus string
	err = db.QueryRow(`
		SELECT health_status FROM merchant_api_keys WHERE id = $1
	`, apiKeyID).Scan(&healthStatus)

	if err != nil {
		t.Errorf("Failed to query result: %v", err)
	}

	if healthStatus != "degraded" {
		t.Errorf("Expected health_status=degraded, got %s", healthStatus)
	}
}

func TestGetProviderHealth(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	checker := &HealthChecker{db: db}
	ctx := context.Background()

	now := time.Now()
	var apiKeyID int
	err := db.QueryRow(`
		INSERT INTO merchant_api_keys (
			merchant_id, provider, api_key_encrypted, status, 
			health_status, health_check_level, last_health_check_at, consecutive_failures
		)
		VALUES (1, 'openai', 'test-key', 'active', 'healthy', 'medium', $1, 0)
		RETURNING id
	`, now).Scan(&apiKeyID)
	if err != nil {
		t.Fatalf("Failed to create test api key: %v", err)
	}
	defer db.Exec("DELETE FROM merchant_api_keys WHERE id = $1", apiKeyID)

	health, err := checker.GetProviderHealth(ctx, apiKeyID)
	if err != nil {
		t.Errorf("GetProviderHealth failed: %v", err)
	}

	if health == nil {
		t.Fatal("Expected health data, got nil")
	}

	if health.APIKeyID != apiKeyID {
		t.Errorf("Expected api_key_id=%d, got %d", apiKeyID, health.APIKeyID)
	}

	if health.Provider != "openai" {
		t.Errorf("Expected provider=openai, got %s", health.Provider)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected status=healthy, got %s", health.Status)
	}
}

func TestGetAllProviderHealth(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	checker := &HealthChecker{db: db}
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := db.Exec(`
			INSERT INTO merchant_api_keys (merchant_id, provider, api_key_encrypted, status, health_status)
			VALUES (1, $1, 'test-key', 'active', 'healthy')
		`, []string{"openai", "anthropic", "google"}[i])
		if err != nil {
			t.Fatalf("Failed to create test api key: %v", err)
		}
	}
	defer db.Exec("DELETE FROM merchant_api_keys WHERE api_key_encrypted = 'test-key'")

	providers, err := checker.GetAllProviderHealth(ctx)
	if err != nil {
		t.Errorf("GetAllProviderHealth failed: %v", err)
	}

	if len(providers) < 3 {
		t.Errorf("Expected at least 3 providers, got %d", len(providers))
	}

	for _, p := range providers {
		if p.Status == "" {
			t.Error("Provider status should not be empty")
		}
		if p.Provider == "" {
			t.Error("Provider name should not be empty")
		}
	}
}

func TestSaveHealthCheckResult(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	checker := &HealthChecker{db: db}
	ctx := context.Background()

	var apiKeyID int
	err := db.QueryRow(`
		INSERT INTO merchant_api_keys (merchant_id, provider, api_key_encrypted, status, health_status)
		VALUES (1, 'openai', 'test-key', 'active', 'unknown')
		RETURNING id
	`).Scan(&apiKeyID)
	if err != nil {
		t.Fatalf("Failed to create test api key: %v", err)
	}
	defer db.Exec("DELETE FROM merchant_api_keys WHERE id = $1", apiKeyID)

	result := &HealthCheckResult{
		Success:     true,
		Status:      HealthStatusHealthy,
		LatencyMs:   150,
		ModelsFound: []string{"gpt-4", "gpt-3.5-turbo"},
		CheckType:   "full",
	}

	err = checker.SaveHealthCheckResult(ctx, apiKeyID, result)
	if err != nil {
		t.Errorf("SaveHealthCheckResult failed: %v", err)
	}

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM api_key_health_history WHERE api_key_id = $1
	`, apiKeyID).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query history: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 history record, got %d", count)
	}

	var healthStatus string
	err = db.QueryRow(`
		SELECT health_status FROM merchant_api_keys WHERE id = $1
	`, apiKeyID).Scan(&healthStatus)
	if err != nil {
		t.Errorf("Failed to query health status: %v", err)
	}

	if healthStatus != "healthy" {
		t.Errorf("Expected health_status=healthy, got %s", healthStatus)
	}
}

func TestHealthScheduler_StartStop(t *testing.T) {
	scheduler := GetHealthScheduler()

	scheduler.Start()
	time.Sleep(100 * time.Millisecond)

	stats := scheduler.GetStats()
	if !stats["running"].(bool) {
		t.Error("Expected scheduler to be running")
	}

	scheduler.Stop()

	stats = scheduler.GetStats()
	if stats["running"].(bool) {
		t.Error("Expected scheduler to be stopped")
	}
}

func TestHealthScheduler_SetCheckLevel(t *testing.T) {
	scheduler := GetHealthScheduler()

	scheduler.SetCheckLevel(HealthCheckLevelHigh)

	stats := scheduler.GetStats()
	if stats["check_level"].(string) != "high" {
		t.Errorf("Expected check_level=high, got %s", stats["check_level"])
	}

	if stats["interval"].(int) != 60 {
		t.Errorf("Expected interval=60, got %d", stats["interval"])
	}
}
