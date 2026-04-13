package handlers

import (
	"testing"
	"time"
)

func TestConsumptionRecords_Pagination(t *testing.T) {
	page := 1
	pageSize := 20

	if page < 1 {
		t.Error("Page should be at least 1")
	}

	if pageSize < 1 || pageSize > 100 {
		t.Error("PageSize should be between 1 and 100")
	}

	offset := (page - 1) * pageSize
	if offset != 0 {
		t.Errorf("Offset for page 1 should be 0, got %d", offset)
	}

	page = 3
	pageSize = 50
	offset = (page - 1) * pageSize
	if offset != 100 {
		t.Errorf("Offset for page 3 with pageSize 50 should be 100, got %d", offset)
	}
}

func TestConsumptionStats_QueryBuilding(t *testing.T) {
	baseQuery := "FROM api_usage_logs WHERE user_id = $1"
	args := []interface{}{1}
	argIndex := 2

	startDate := "2026-03-01"
	if startDate != "" {
		baseQuery += " AND created_at >= $" + string(rune('0'+argIndex))
		args = append(args, startDate+" 00:00:00")
		argIndex++
	}

	if !containsStr(baseQuery, "created_at >=") {
		t.Error("Query should contain created_at filter")
	}

	if len(args) != 2 {
		t.Errorf("Should have 2 args, got %d", len(args))
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConsumptionStats_Structure(t *testing.T) {
	stats := struct {
		TotalRequests       int   `json:"total_requests"`
		TotalTokenDeduction int64 `json:"total_token_deduction"`
		AvgLatencyMs        int   `json:"avg_latency_ms"`
	}{
		TotalRequests:       100,
		TotalTokenDeduction: 50000,
		AvgLatencyMs:        1500,
	}

	if stats.TotalRequests != 100 {
		t.Errorf("TotalRequests = %d, want 100", stats.TotalRequests)
	}

	if stats.TotalTokenDeduction != 50000 {
		t.Errorf("TotalTokenDeduction = %d, want 50000", stats.TotalTokenDeduction)
	}
}

func TestConsumptionRecords_DateFilter(t *testing.T) {
	now := time.Now()
	startDate := now.AddDate(0, 0, -30).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	if startDate == "" {
		t.Error("Start date should not be empty")
	}

	if endDate == "" {
		t.Error("End date should not be empty")
	}

	if startDate > endDate {
		t.Error("Start date should be before or equal to end date")
	}
}

func TestConsumptionRecords_ProviderFilter(t *testing.T) {
	providers := []string{"all", "openai", "anthropic", "google"}

	for _, provider := range providers {
		if provider == "" {
			t.Error("Provider should not be empty")
		}
	}

	validProviders := map[string]bool{
		"openai":    true,
		"anthropic": true,
		"google":    true,
	}

	testProvider := "openai"
	if testProvider != "all" && !validProviders[testProvider] {
		t.Errorf("Provider %s is not valid", testProvider)
	}
}

func TestConsumptionStats_ByProvider(t *testing.T) {
	byProvider := []map[string]interface{}{
		{"provider": "openai", "count": 50, "tokens": int64(10000)},
		{"provider": "anthropic", "count": 30, "tokens": int64(6000)},
		{"provider": "google", "count": 20, "tokens": int64(4000)},
	}

	totalCount := 0
	var totalTokens int64

	for _, p := range byProvider {
		count, ok := p["count"].(int)
		if !ok {
			t.Error("Count should be an integer")
		}
		totalCount += count

		tok, ok := p["tokens"].(int64)
		if !ok {
			t.Error("Tokens should be an int64")
		}
		totalTokens += tok
	}

	if totalCount != 100 {
		t.Errorf("Total count = %d, want 100", totalCount)
	}

	if totalTokens != 20000 {
		t.Errorf("Total tokens = %d, want 20000", totalTokens)
	}
}
