package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/services"
)

type StrategyConfigHandler struct {
	strategyEngine *services.RoutingStrategyEngine
}

func NewStrategyConfigHandler() *StrategyConfigHandler {
	return &StrategyConfigHandler{
		strategyEngine: services.NewRoutingStrategyEngine(),
	}
}

func (h *StrategyConfigHandler) GetStrategies(c *gin.Context) {
	strategies := []map[string]interface{}{
		{
			"id":          "performance_first",
			"name":        "性能优先",
			"description": "优先选择延迟最低的 API Key",
			"weights": map[string]float64{
				"latency":      0.5,
				"cost":         0.1,
				"reliability":  0.2,
				"security":     0.1,
				"load_balance": 0.1,
			},
		},
		{
			"id":          "price_first",
			"name":        "价格优先",
			"description": "优先选择成本最低的 API Key",
			"weights": map[string]float64{
				"latency":      0.1,
				"cost":         0.6,
				"reliability":  0.15,
				"security":     0.05,
				"load_balance": 0.1,
			},
		},
		{
			"id":          "reliability_first",
			"name":        "可靠性优先",
			"description": "优先选择成功率最高的 API Key",
			"weights": map[string]float64{
				"latency":      0.2,
				"cost":         0.1,
				"reliability":  0.5,
				"security":     0.1,
				"load_balance": 0.1,
			},
		},
		{
			"id":          "security_first",
			"name":        "安全优先",
			"description": "优先选择安全等级最高的 API Key",
			"weights": map[string]float64{
				"latency":      0.1,
				"cost":         0.1,
				"reliability":  0.2,
				"security":     0.5,
				"load_balance": 0.1,
			},
		},
		{
			"id":          "balanced",
			"name":        "均衡策略",
			"description": "平衡各项指标的均衡策略",
			"weights": map[string]float64{
				"latency":      0.25,
				"cost":         0.25,
				"reliability":  0.25,
				"security":     0.15,
				"load_balance": 0.1,
			},
		},
		{
			"id":          "auto",
			"name":        "自动模式",
			"description": "根据请求特征自动选择最优策略",
			"weights": map[string]float64{
				"latency":      0.3,
				"cost":         0.2,
				"reliability":  0.3,
				"security":     0.1,
				"load_balance": 0.1,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    strategies,
	})
}

func (h *StrategyConfigHandler) GetStrategyWeights(c *gin.Context) {
	strategyID := c.Param("id")

	goal := services.StrategyGoal(strategyID)
	weights, err := h.strategyEngine.GetStrategyWeights(goal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"strategy": strategyID,
			"weights":  weights,
		},
	})
}

func (h *StrategyConfigHandler) AnalyzeRequest(c *gin.Context) {
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	analyzer := services.NewRequestAnalyzer()

	bodyBytes, _ := json.Marshal(requestBody)
	req := &http.Request{
		Method: "POST",
		Body:   io.NopCloser(bytes.NewBuffer(bodyBytes)),
	}

	analysis, err := analyzer.Analyze(c.Request.Context(), req, bodyBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analysis,
	})
}

func (h *StrategyConfigHandler) DefineStrategyGoal(c *gin.Context) {
	var reqCtx services.RequestContext
	if err := c.ShouldBindJSON(&reqCtx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request context",
		})
		return
	}

	output, err := h.strategyEngine.DefineGoal(c.Request.Context(), &reqCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    output,
	})
}
