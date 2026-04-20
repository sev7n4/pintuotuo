package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

type ProviderHealthResponse struct {
	APIKeyID      int      `json:"api_key_id"`
	Provider      string   `json:"provider"`
	Status        string   `json:"status"`
	LatencyMs     int      `json:"latency_ms,omitempty"`
	ErrorMessage  string   `json:"error_message,omitempty"`
	Models        []string `json:"models,omitempty"`
	LastCheckedAt string   `json:"last_checked_at,omitempty"`
}

type AllProvidersHealthResponse struct {
	Total     int                      `json:"total"`
	Healthy   int                      `json:"healthy"`
	Degraded  int                      `json:"degraded"`
	Unhealthy int                      `json:"unhealthy"`
	Unknown   int                      `json:"unknown"`
	Providers []ProviderHealthResponse `json:"providers"`
}

func GetProviderHealth(c *gin.Context) {
	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	checker := services.NewHealthChecker()
	health, err := checker.GetProviderHealth(c.Request.Context(), keyID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PROVIDER_NOT_FOUND",
			"Provider not found",
			http.StatusNotFound,
			err,
		))
		return
	}

	lastCheckedAt := ""
	if !health.LastCheckedAt.IsZero() {
		lastCheckedAt = health.LastCheckedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	c.JSON(http.StatusOK, ProviderHealthResponse{
		APIKeyID:      health.APIKeyID,
		Provider:      health.Provider,
		Status:        health.Status,
		LastCheckedAt: lastCheckedAt,
	})
}

func GetAllProvidersHealth(c *gin.Context) {
	checker := services.NewHealthChecker()
	providers, err := checker.GetAllProviderHealth(c.Request.Context())
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	response := AllProvidersHealthResponse{
		Total:     len(providers),
		Providers: make([]ProviderHealthResponse, 0, len(providers)),
	}

	for _, p := range providers {
		lastCheckedAt := ""
		if !p.LastCheckedAt.IsZero() {
			lastCheckedAt = p.LastCheckedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		response.Providers = append(response.Providers, ProviderHealthResponse{
			APIKeyID:      p.APIKeyID,
			Provider:      p.Provider,
			Status:        p.Status,
			LastCheckedAt: lastCheckedAt,
		})

		switch p.Status {
		case services.HealthStatusHealthy:
			response.Healthy++
		case services.HealthStatusDegraded:
			response.Degraded++
		case services.HealthStatusUnhealthy:
			response.Unhealthy++
		default:
			response.Unknown++
		}
	}

	c.JSON(http.StatusOK, response)
}

func TriggerHealthCheck(c *gin.Context) {
	userRole, exists := c.Get("user_role")
	if !exists || userRole != roleAdmin {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"FORBIDDEN",
			"Admin access required",
			http.StatusForbidden,
			nil,
		))
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var keyExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM merchant_api_keys WHERE id = $1)", keyID).Scan(&keyExists)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	if !keyExists {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"PROVIDER_NOT_FOUND",
			"Provider not found",
			http.StatusNotFound,
			nil,
		))
		return
	}

	go func() {
		_ = services.GetHealthScheduler().TriggerImmediateCheck(keyID)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Immediate health check triggered",
		"api_key_id": keyID,
	})
}

func GetHealthCheckHistory(c *gin.Context) {
	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, parseErr := strconv.Atoi(limitStr); parseErr == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	rows, err := db.Query(`
		SELECT id, check_type, status, latency_ms, error_message, models_available, created_at
		FROM api_key_health_history 
		WHERE api_key_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`,
		keyID, limit,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	type HistoryEntry struct {
		ID              int      `json:"id"`
		CheckType       string   `json:"check_type"`
		Status          string   `json:"status"`
		LatencyMs       int      `json:"latency_ms"`
		ErrorMessage    string   `json:"error_message,omitempty"`
		ModelsAvailable []string `json:"models_available,omitempty"`
		CreatedAt       string   `json:"created_at"`
	}

	var history []HistoryEntry
	for rows.Next() {
		var entry HistoryEntry
		var modelsJSON []byte
		var createdAt string

		err := rows.Scan(&entry.ID, &entry.CheckType, &entry.Status, &entry.LatencyMs,
			&entry.ErrorMessage, &modelsJSON, &createdAt)
		if err != nil {
			continue
		}

		if modelsJSON != nil {
			_ = json.Unmarshal(modelsJSON, &entry.ModelsAvailable)
		}
		entry.CreatedAt = createdAt
		history = append(history, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key_id": keyID,
		"history":    history,
	})
}

func GetHealthCheckStats(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	type Stats struct {
		TotalProviders  int `json:"total_providers"`
		HealthyCount    int `json:"healthy_count"`
		DegradedCount   int `json:"degraded_count"`
		UnhealthyCount  int `json:"unhealthy_count"`
		UnknownCount    int `json:"unknown_count"`
		TotalChecks24h  int `json:"total_checks_24h"`
		FailedChecks24h int `json:"failed_checks_24h"`
		AvgLatencyMs    int `json:"avg_latency_ms"`
	}

	var stats Stats

	err := db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE health_status = 'healthy') as healthy,
			COUNT(*) FILTER (WHERE health_status = 'degraded') as degraded,
			COUNT(*) FILTER (WHERE health_status = 'unhealthy') as unhealthy,
			COUNT(*) FILTER (WHERE health_status = 'unknown' OR health_status IS NULL) as unknown
		FROM merchant_api_keys WHERE status = 'active'`,
	).Scan(&stats.TotalProviders, &stats.HealthyCount, &stats.DegradedCount,
		&stats.UnhealthyCount, &stats.UnknownCount)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	err = db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status != 'healthy') as failed,
			COALESCE(AVG(latency_ms), 0)::int as avg_latency
		FROM api_key_health_history 
		WHERE created_at > NOW() - INTERVAL '24 hours'`,
	).Scan(&stats.TotalChecks24h, &stats.FailedChecks24h, &stats.AvgLatencyMs)
	if err != nil {
		stats.TotalChecks24h = 0
		stats.FailedChecks24h = 0
		stats.AvgLatencyMs = 0
	}

	c.JSON(http.StatusOK, stats)
}
