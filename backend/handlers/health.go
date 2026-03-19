package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/db"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]Service `json:"services"`
	System    SystemInfo        `json:"system"`
}

type Service struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemAllocMB   uint64 `json:"mem_alloc_mb"`
	MemTotalMB   uint64 `json:"mem_total_mb"`
	MemSysMB     uint64 `json:"mem_sys_mb"`
}

type DBPoolStats struct {
	MaxOpenConnections int    `json:"max_open_connections"`
	OpenConnections    int    `json:"open_connections"`
	InUse              int    `json:"in_use"`
	Idle               int    `json:"idle"`
	WaitCount          int64  `json:"wait_count"`
	WaitDuration       string `json:"wait_duration"`
	MaxIdleClosed      int64  `json:"max_idle_closed"`
	MaxLifetimeClosed  int64  `json:"max_lifetime_closed"`
}

func HealthCheck(c *gin.Context) {
	services := make(map[string]Service)

	start := time.Now()
	err := db.HealthCheck(c.Request.Context())
	latency := time.Since(start)

	if err != nil {
		services["database"] = Service{
			Status: "unhealthy",
			Error:  err.Error(),
		}
	} else {
		services["database"] = Service{
			Status:  "healthy",
			Latency: latency.String(),
		}
	}

	start = time.Now()
	redisErr := cache.HealthCheck(c.Request.Context())
	latency = time.Since(start)

	if redisErr != nil {
		services["redis"] = Service{
			Status: "unhealthy",
			Error:  redisErr.Error(),
		}
	} else {
		services["redis"] = Service{
			Status:  "healthy",
			Latency: latency.String(),
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	system := SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemAllocMB:   m.Alloc / 1024 / 1024,
		MemTotalMB:   m.TotalAlloc / 1024 / 1024,
		MemSysMB:     m.Sys / 1024 / 1024,
	}

	overallStatus := "healthy"
	for _, svc := range services {
		if svc.Status != "healthy" {
			overallStatus = "degraded"
			break
		}
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
		Services:  services,
		System:    system,
	})
}

func DBStats(c *gin.Context) {
	stats := db.GetPoolStats()
	if stats == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "database not initialized",
		})
		return
	}

	c.JSON(http.StatusOK, DBPoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration.String(),
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	})
}

func ReadyCheck(c *gin.Context) {
	if err := db.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func LiveCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
