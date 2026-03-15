package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
)

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]ServiceStatus `json:"services"`
	Uptime    int64                  `json:"uptime_seconds"`
}

// ServiceStatus represents the status of a single service
type ServiceStatus struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Duration int64  `json:"response_time_ms"`
}

var startTime = time.Now()

// HealthCheck returns overall health status (liveness probe)
// Used by Kubernetes, load balancers, and monitoring systems
func HealthCheck(c *gin.Context) {
	response := HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  make(map[string]ServiceStatus),
		Uptime:    int64(time.Since(startTime).Seconds()),
	}

	// Application is running
	response.Services["application"] = ServiceStatus{
		Status:   "up",
		Duration: 1,
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessProbe checks if all critical dependencies are ready
// Used by Kubernetes and deployment orchestration
func ReadinessProbe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response := HealthCheckResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  make(map[string]ServiceStatus),
		Uptime:    int64(time.Since(startTime).Seconds()),
	}

	allHealthy := true

	// Check Database connectivity
	start := time.Now()
	db := config.GetDB()
	dbDuration := time.Since(start).Milliseconds()

	if db == nil {
		response.Services["database"] = ServiceStatus{
			Status:   "down",
			Message:  "database not initialized",
			Duration: dbDuration,
		}
		allHealthy = false
	} else {
		dbErr := db.PingContext(ctx)
		if dbErr != nil {
			response.Services["database"] = ServiceStatus{
				Status:   "down",
				Message:  dbErr.Error(),
				Duration: dbDuration,
			}
			allHealthy = false
		} else {
			response.Services["database"] = ServiceStatus{
				Status:   "healthy",
				Duration: dbDuration,
			}
		}
	}

	// Check Redis connectivity
	start = time.Now()
	redisClient := cache.GetClient()
	redisDuration := time.Since(start).Milliseconds()

	if redisClient == nil {
		response.Services["redis"] = ServiceStatus{
			Status:   "down",
			Message:  "redis client not initialized",
			Duration: redisDuration,
		}
		allHealthy = false
	} else {
		redisCmd := redisClient.Ping(ctx)
		if redisCmd.Err() != nil {
			response.Services["redis"] = ServiceStatus{
				Status:   "down",
				Message:  redisCmd.Err().Error(),
				Duration: redisDuration,
			}
			allHealthy = false
		} else {
			response.Services["redis"] = ServiceStatus{
				Status:   "healthy",
				Duration: redisDuration,
			}
		}
	}

	// Overall status
	if allHealthy {
		response.Status = "ready"
		c.JSON(http.StatusOK, response)
	} else {
		response.Status = "not_ready"
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// LivenessProbe checks if application is still running
// Kubernetes uses this to determine if pod should be restarted
func LivenessProbe(c *gin.Context) {
	response := HealthCheckResponse{
		Status:    "alive",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  make(map[string]ServiceStatus),
		Uptime:    int64(time.Since(startTime).Seconds()),
	}

	// Application is running if we can respond to this request
	c.JSON(http.StatusOK, response)
}

// Metrics returns basic service metrics
func Metrics(c *gin.Context) {
	response := gin.H{
		"uptime_seconds":   int64(time.Since(startTime).Seconds()),
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"status":           "running",
		"version":          "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}
