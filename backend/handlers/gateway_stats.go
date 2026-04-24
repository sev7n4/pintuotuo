package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/services"
)

var unifiedGateway *services.UnifiedGatewayImpl

func GetUnifiedGateway() *services.UnifiedGatewayImpl {
	if unifiedGateway == nil {
		unifiedGateway = services.NewUnifiedGateway()
	}
	return unifiedGateway
}

func GetGatewayStats(c *gin.Context) {
	gateway := GetUnifiedGateway()
	stats := gateway.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"gateway":       stats,
		"rate_limiters": services.GetRateLimiter().GetAllStats(),
		"queues":        services.GetQueueFactory().GetAllStats(),
	})
}

func GetRateLimiterStats(c *gin.Context) {
	key := c.Query("key")
	if key != "" {
		stats := services.GetRateLimiter().GetStats(key)
		c.JSON(http.StatusOK, gin.H{
			"key":   key,
			"stats": stats,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"limiters": services.GetRateLimiter().GetAllStats(),
		"count":    services.GetRateLimiter().GetLimiterCount(),
	})
}

func UpdateRateLimiterConfig(c *gin.Context) {
	var req struct {
		Key   string `json:"key" binding:"required"`
		Rate  int    `json:"rate" binding:"required,min=1"`
		Burst int    `json:"burst" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	services.GetRateLimiter().SetRate(req.Key, req.Rate, req.Burst)

	c.JSON(http.StatusOK, gin.H{
		"message": "rate limiter config updated",
		"key":     req.Key,
		"rate":    req.Rate,
		"burst":   req.Burst,
	})
}

func ResetRateLimiterStats(c *gin.Context) {
	key := c.Query("key")
	if key != "" {
		services.GetRateLimiter().ResetStats(key)
		c.JSON(http.StatusOK, gin.H{"message": "rate limiter stats reset", "key": key})
		return
	}

	services.GetRateLimiter().ResetAllStats()
	c.JSON(http.StatusOK, gin.H{"message": "all rate limiter stats reset"})
}

func GetQueueStats(c *gin.Context) {
	key := c.Query("key")
	if key != "" {
		stats := services.GetQueueFactory().GetStats(key)
		c.JSON(http.StatusOK, gin.H{
			"key":   key,
			"stats": stats,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queues": services.GetQueueFactory().GetAllStats(),
		"count":  services.GetQueueFactory().GetQueueCount(),
	})
}

func UpdateQueueConfig(c *gin.Context) {
	var req struct {
		Key     string `json:"key" binding:"required"`
		MaxSize int    `json:"max_size" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queue := services.GetQueueFactory().GetQueue(req.Key, req.MaxSize)
	queue.SetMaxSize(req.MaxSize)

	c.JSON(http.StatusOK, gin.H{
		"message":  "queue config updated",
		"key":      req.Key,
		"max_size": req.MaxSize,
	})
}

func ResetQueueStats(c *gin.Context) {
	key := c.Query("key")
	if key != "" {
		services.GetQueueFactory().ResetStats(key)
		c.JSON(http.StatusOK, gin.H{"message": "queue stats reset", "key": key})
		return
	}

	services.GetQueueFactory().ResetAllStats()
	c.JSON(http.StatusOK, gin.H{"message": "all queue stats reset"})
}
