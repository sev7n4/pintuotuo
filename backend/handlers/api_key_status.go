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
			if healthResult.Status != "healthy" {
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
		SELECT api_key_id, latency_p50, latency_p95, latency_p99,
		       error_rate, success_rate, connection_pool_size, connection_pool_active,
		       rate_limit_remaining, rate_limit_reset_at, load_balance_weight,
		       last_request_at, updated_at
		FROM api_key_realtime_status
		ORDER BY api_key_id ASC
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

	var statuses []map[string]interface{}
	for rows.Next() {
		var status map[string]interface{} = make(map[string]interface{})
		var (
			apiKeyID             int
			latencyP50           int
			latencyP95           int
			latencyP99           int
			errorRate            float64
			successRate          float64
			connectionPoolSize   int
			connectionPoolActive int
			rateLimitRemaining   int
			rateLimitResetAt     interface{}
			loadBalanceWeight    float64
			lastRequestAt        interface{}
			updatedAt            interface{}
		)

		err := rows.Scan(
			&apiKeyID,
			&latencyP50,
			&latencyP95,
			&latencyP99,
			&errorRate,
			&successRate,
			&connectionPoolSize,
			&connectionPoolActive,
			&rateLimitRemaining,
			&rateLimitResetAt,
			&loadBalanceWeight,
			&lastRequestAt,
			&updatedAt,
		)
		if err != nil {
			continue
		}

		status["api_key_id"] = apiKeyID
		status["latency_p50"] = latencyP50
		status["latency_p95"] = latencyP95
		status["latency_p99"] = latencyP99
		status["error_rate"] = errorRate
		status["success_rate"] = successRate
		status["connection_pool_size"] = connectionPoolSize
		status["connection_pool_active"] = connectionPoolActive
		status["rate_limit_remaining"] = rateLimitRemaining
		status["rate_limit_reset_at"] = rateLimitResetAt
		status["load_balance_weight"] = loadBalanceWeight
		status["last_request_at"] = lastRequestAt
		status["updated_at"] = updatedAt

		statuses = append(statuses, status)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    statuses,
	})
}
