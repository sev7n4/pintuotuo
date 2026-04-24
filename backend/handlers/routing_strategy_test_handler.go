package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/services"
)

func TestRoutingStrategy(c *gin.Context) {
	var req struct {
		StrategyCode string `json:"strategy_code" binding:"required"`
		Model        string `json:"model" binding:"required"`
		Provider     string `json:"provider"`
		MerchantID   int    `json:"merchant_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MerchantID == 0 {
		req.MerchantID = 1
	}

	strategy := services.RoutingStrategy(req.StrategyCode)
	candidates, err := services.GetSmartRouter().GetCandidatesWithKeyAllowlist(
		c.Request.Context(),
		req.Model,
		req.Provider,
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(candidates) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"strategy_code": req.StrategyCode,
			"model":         req.Model,
			"provider":      req.Provider,
			"candidates":    []interface{}{},
			"selected":      nil,
			"message":       "no available candidates",
		})
		return
	}

	healthyCandidates := services.GetSmartRouter().FilterUnhealthy(candidates)
	verifiedCandidates := services.GetSmartRouter().FilterUnverified(healthyCandidates)

	services.GetSmartRouter().CalculateScores(verifiedCandidates, strategy)

	selected := verifiedCandidates[0]

	c.JSON(http.StatusOK, gin.H{
		"strategy_code":    req.StrategyCode,
		"model":            req.Model,
		"provider":         req.Provider,
		"total_candidates": len(candidates),
		"healthy_count":    len(healthyCandidates),
		"verified_count":   len(verifiedCandidates),
		"candidates":       verifiedCandidates,
		"selected": gin.H{
			"api_key_id":     selected.APIKeyID,
			"provider":       selected.Provider,
			"model":          selected.Model,
			"score":          selected.Score,
			"price_score":    selected.PriceScore,
			"latency_score":  selected.LatencyScore,
			"success_score":  selected.SuccessScore,
			"health_status":  selected.HealthStatus,
			"avg_latency_ms": selected.AvgLatencyMs,
			"success_rate":   selected.SuccessRate,
		},
	})
}
