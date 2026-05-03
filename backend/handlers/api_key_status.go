package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

func GetAPIKeyStatus(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	apiKeyIDStr := c.Param("id")
	apiKeyID, err := strconv.Atoi(apiKeyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid API Key ID",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Database connection error",
		})
		return
	}

	awarenessService := services.NewRouteAwarenessService(db)
	status, err := awarenessService.GetRealtimeStatus(c.Request.Context(), apiKeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get API Key status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    status,
	})
}

func GetBatchAPIKeyStatus(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	var request struct {
		APIKeyIDs []int `json:"api_key_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Database connection error",
		})
		return
	}

	awarenessService := services.NewRouteAwarenessService(db)
	statuses, err := awarenessService.GetBatchStatus(c.Request.Context(), request.APIKeyIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get API Key statuses",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    statuses,
	})
}

func TriggerStatusCollect(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Database connection error",
		})
		return
	}

	awarenessService := services.NewRouteAwarenessService(db)
	healthChecker := services.NewHealthChecker()

	go func() {
		ctx := c.Request.Context()
		query := `SELECT id, merchant_id, name, provider, status, api_key_encrypted, api_secret_encrypted, endpoint_url FROM merchant_api_keys WHERE status = 'active'`
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("Failed to get active API keys: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var apiKey models.MerchantAPIKey
			var apiSecretEncrypted, endpointURL sql.NullString
			if err := rows.Scan(&apiKey.ID, &apiKey.MerchantID, &apiKey.Name, &apiKey.Provider, &apiKey.Status, &apiKey.APIKeyEncrypted, &apiSecretEncrypted, &endpointURL); err != nil {
				continue
			}
			if apiSecretEncrypted.Valid {
				apiKey.APISecretEncrypted = apiSecretEncrypted.String
			}
			if endpointURL.Valid {
				apiKey.EndpointURL = endpointURL.String
			}

			healthResult, err := healthChecker.LightweightPing(ctx, &apiKey)
			if err != nil {
				log.Printf("Health check failed for API key %d: %v", apiKey.ID, err)
				continue
			}

			latencyMs := 0
			if healthResult.LatencyMs > 0 {
				latencyMs = int(healthResult.LatencyMs)
			}

			statusUpdate := &models.APIKeyRealtimeStatus{
				APIKeyID:    apiKey.ID,
				LatencyP50:  latencyMs,
				LatencyP95:  latencyMs,
				LatencyP99:  latencyMs,
				ErrorRate:   0,
				SuccessRate: 1.0,
			}
			if healthResult.Status != services.HealthStatusHealthy {
				statusUpdate.ErrorRate = 1.0
				statusUpdate.SuccessRate = 0
			}

			if err := awarenessService.UpdateStatus(ctx, apiKey.ID, statusUpdate); err != nil {
				log.Printf("Failed to update status for API key %d: %v", apiKey.ID, err)
			}
		}
		log.Println("Status collection triggered successfully")
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"message": "状态采集已触发，请稍后刷新查看结果",
		},
	})
}

func GetAllAPIKeyStatus(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Database connection error",
		})
		return
	}

	query := `
		SELECT 
			mak.id, mak.merchant_id, mak.name, mak.provider, mak.status,
			mak.region, mak.security_level, mak.endpoint_url, mak.health_status,
			COALESCE(ars.latency_p50, 0) as latency_p50,
			COALESCE(ars.latency_p95, 0) as latency_p95,
			COALESCE(ars.latency_p99, 0) as latency_p99,
			COALESCE(ars.error_rate, 0) as error_rate,
			COALESCE(ars.success_rate, 1.0) as success_rate,
			COALESCE(ars.connection_pool_size, 0) as connection_pool_size,
			COALESCE(ars.connection_pool_active, 0) as connection_pool_active,
			COALESCE(ars.rate_limit_remaining, 0) as rate_limit_remaining,
			COALESCE(ars.load_balance_weight, 1.0) as load_balance_weight,
			ars.last_request_at,
			ars.updated_at,
			mak.created_at
		FROM merchant_api_keys mak
		LEFT JOIN api_key_realtime_status ars ON mak.id = ars.api_key_id
		ORDER BY mak.id ASC
	`

	rows, err := db.QueryContext(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to query API Key statuses",
		})
		return
	}
	defer rows.Close()

	var statuses []MerchantAPIKeyStatusResponse
	for rows.Next() {
		var status MerchantAPIKeyStatusResponse
		var region, securityLevel, endpointURL, healthStatus sql.NullString
		var lastRequestAt, statusUpdatedAt, createdAt sql.NullTime

		err := rows.Scan(
			&status.ID,
			&status.MerchantID,
			&status.Name,
			&status.Provider,
			&status.Status,
			&region,
			&securityLevel,
			&endpointURL,
			&healthStatus,
			&status.LatencyP50,
			&status.LatencyP95,
			&status.LatencyP99,
			&status.ErrorRate,
			&status.SuccessRate,
			&status.ConnectionPoolSize,
			&status.ConnectionPoolActive,
			&status.RateLimitRemaining,
			&status.LoadBalanceWeight,
			&lastRequestAt,
			&statusUpdatedAt,
			&createdAt,
		)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		if region.Valid {
			status.Region = region.String
		}
		if securityLevel.Valid {
			status.SecurityLevel = securityLevel.String
		}
		if endpointURL.Valid {
			status.EndpointURL = endpointURL.String
		}
		if healthStatus.Valid {
			status.HealthStatus = healthStatus.String
		}
		if lastRequestAt.Valid {
			status.LastRequestAt = lastRequestAt.Time.Format("2006-01-02 15:04:05")
		}
		if statusUpdatedAt.Valid {
			status.StatusUpdatedAt = statusUpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if createdAt.Valid {
			status.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}

		statuses = append(statuses, status)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    statuses,
	})
}

type MerchantAPIKeyStatusResponse struct {
	ID                   int     `json:"id"`
	MerchantID           int     `json:"merchant_id"`
	Name                 string  `json:"name"`
	Provider             string  `json:"provider"`
	Status               string  `json:"status"`
	Region               string  `json:"region"`
	SecurityLevel        string  `json:"security_level"`
	EndpointURL          string  `json:"endpoint_url"`
	HealthStatus         string  `json:"health_status"`
	LatencyP50           int     `json:"latency_p50"`
	LatencyP95           int     `json:"latency_p95"`
	LatencyP99           int     `json:"latency_p99"`
	ErrorRate            float64 `json:"error_rate"`
	SuccessRate          float64 `json:"success_rate"`
	ConnectionPoolSize   int     `json:"connection_pool_size"`
	ConnectionPoolActive int     `json:"connection_pool_active"`
	RateLimitRemaining   int     `json:"rate_limit_remaining"`
	LoadBalanceWeight    float64 `json:"load_balance_weight"`
	LastRequestAt        string  `json:"last_request_at"`
	StatusUpdatedAt      string  `json:"status_updated_at"`
	CreatedAt            string  `json:"created_at"`
}

func GetAPIKeyStatusByMerchantID(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "merchant_id is required",
		})
		return
	}

	merchantID, err := strconv.Atoi(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid merchant_id",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Database connection error",
		})
		return
	}

	query := `
		SELECT 
			mak.id, mak.merchant_id, mak.name, mak.provider, mak.status,
			mak.region, mak.security_level, mak.endpoint_url, mak.health_status,
			COALESCE(ars.latency_p50, 0) as latency_p50,
			COALESCE(ars.latency_p95, 0) as latency_p95,
			COALESCE(ars.latency_p99, 0) as latency_p99,
			COALESCE(ars.error_rate, 0) as error_rate,
			COALESCE(ars.success_rate, 1.0) as success_rate,
			COALESCE(ars.connection_pool_size, 0) as connection_pool_size,
			COALESCE(ars.connection_pool_active, 0) as connection_pool_active,
			COALESCE(ars.rate_limit_remaining, 0) as rate_limit_remaining,
			COALESCE(ars.load_balance_weight, 1.0) as load_balance_weight,
			ars.last_request_at,
			ars.updated_at as status_updated_at,
			mak.created_at
		FROM merchant_api_keys mak
		LEFT JOIN api_key_realtime_status ars ON mak.id = ars.api_key_id
		WHERE mak.merchant_id = $1
		ORDER BY mak.id ASC
	`

	rows, err := db.QueryContext(c.Request.Context(), query, merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to query API Key statuses",
		})
		return
	}
	defer rows.Close()

	var statuses []MerchantAPIKeyStatusResponse
	for rows.Next() {
		var status MerchantAPIKeyStatusResponse
		var region, securityLevel, endpointURL, healthStatus sql.NullString
		var lastRequestAt, statusUpdatedAt, createdAt sql.NullTime

		err := rows.Scan(
			&status.ID,
			&status.MerchantID,
			&status.Name,
			&status.Provider,
			&status.Status,
			&region,
			&securityLevel,
			&endpointURL,
			&healthStatus,
			&status.LatencyP50,
			&status.LatencyP95,
			&status.LatencyP99,
			&status.ErrorRate,
			&status.SuccessRate,
			&status.ConnectionPoolSize,
			&status.ConnectionPoolActive,
			&status.RateLimitRemaining,
			&status.LoadBalanceWeight,
			&lastRequestAt,
			&statusUpdatedAt,
			&createdAt,
		)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		if region.Valid {
			status.Region = region.String
		}
		if securityLevel.Valid {
			status.SecurityLevel = securityLevel.String
		}
		if endpointURL.Valid {
			status.EndpointURL = endpointURL.String
		}
		if healthStatus.Valid {
			status.HealthStatus = healthStatus.String
		}
		if lastRequestAt.Valid {
			status.LastRequestAt = lastRequestAt.Time.Format("2006-01-02 15:04:05")
		}
		if statusUpdatedAt.Valid {
			status.StatusUpdatedAt = statusUpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if createdAt.Valid {
			status.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}

		statuses = append(statuses, status)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    statuses,
	})
}
