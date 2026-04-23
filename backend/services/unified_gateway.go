package services

import (
	"context"
	"fmt"
	"time"
)

type UnifiedGateway interface {
	HandleRequest(ctx context.Context, req *GatewayRequest) (*GatewayResponse, error)
	GetStats() map[string]interface{}
}

type GatewayRequest struct {
	RequestID      string                 `json:"request_id"`
	MerchantID     int                    `json:"merchant_id"`
	Model          string                 `json:"model"`
	Provider       string                 `json:"provider,omitempty"`
	RequestBody    map[string]interface{} `json:"request_body"`
	UserPrefs      map[string]interface{} `json:"user_preferences,omitempty"`
	CostBudget     *float64               `json:"cost_budget,omitempty"`
	ComplianceReqs []string               `json:"compliance_requirements,omitempty"`
	AllowedKeyIDs  []int                  `json:"allowed_key_ids,omitempty"`
	Priority       int                    `json:"priority"`
	Timeout        time.Duration          `json:"timeout"`
}

type GatewayResponse struct {
	RequestID      string                 `json:"request_id"`
	MerchantID     int                    `json:"merchant_id"`
	Model          string                 `json:"model"`
	Provider       string                 `json:"provider"`
	APIKeyID       int                    `json:"api_key_id"`
	ResponseBody   map[string]interface{} `json:"response_body"`
	Status         string                 `json:"status"`
	Error          string                 `json:"error,omitempty"`
	LatencyMs      int                    `json:"latency_ms"`
	TokensUsed     int                    `json:"tokens_used,omitempty"`
	Cost           float64                `json:"cost,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

type GatewayStats struct {
	Requests       int64     `json:"requests"`
	Successes      int64     `json:"successes"`
	Failures       int64     `json:"failures"`
	AvgLatencyMs   float64   `json:"avg_latency_ms"`
	MaxLatencyMs   float64   `json:"max_latency_ms"`
	TotalTokens    int64     `json:"total_tokens"`
	TotalCost      float64   `json:"total_cost"`
	LastReset      time.Time `json:"last_reset"`
}

type UnifiedGatewayImpl struct {
	pipeline      *ThreeLayerRoutingPipeline
	rateLimiter   *RateLimiterFactory
	queueFactory  *QueueFactory
	stats         *GatewayStats
	enableQueue   bool
	queueMaxSize  int
	rateLimitRate int
	rateLimitBurst int
}

func NewUnifiedGateway() *UnifiedGatewayImpl {
	return &UnifiedGatewayImpl{
		pipeline:      NewThreeLayerRoutingPipeline(),
		rateLimiter:   GetRateLimiter(),
		queueFactory:  GetQueueFactory(),
		stats: &GatewayStats{
			LastReset: time.Now(),
		},
		enableQueue:   true,
		queueMaxSize: 1000,
		rateLimitRate: 100,
		rateLimitBurst: 200,
	}
}

func (g *UnifiedGatewayImpl) HandleRequest(ctx context.Context, req *GatewayRequest) (*GatewayResponse, error) {
	startTime := time.Now()

	response := &GatewayResponse{
		RequestID: req.RequestID,
		MerchantID: req.MerchantID,
		Model: req.Model,
		Timestamp: startTime,
	}

	// 1. 限流检查
	rateLimitKey := fmt.Sprintf("merchant:%d", req.MerchantID)
	if !g.rateLimiter.Allow(rateLimitKey, g.rateLimitRate, g.rateLimitBurst) {
		response.Status = "rate_limited"
		response.Error = "rate limit exceeded"
		response.LatencyMs = int(time.Since(startTime).Milliseconds())
		return response, fmt.Errorf("rate limit exceeded")
	}

	// 2. 队列管理
	if g.enableQueue {
		queueKey := fmt.Sprintf("merchant:%d:model:%s", req.MerchantID, req.Model)
		queuedReq := &QueuedRequest{
			RequestID: req.RequestID,
			MerchantID: req.MerchantID,
			Model: req.Model,
			Provider: req.Provider,
			RequestBody: req.RequestBody,
			UserPrefs: req.UserPrefs,
			CostBudget: req.CostBudget,
			ComplianceReqs: req.ComplianceReqs,
			AllowedKeyIDs: req.AllowedKeyIDs,
			Priority: req.Priority,
			Timeout: req.Timeout,
		}

		err := g.queueFactory.Enqueue(queueKey, g.queueMaxSize, queuedReq)
		if err != nil {
			response.Status = "queue_full"
			response.Error = "request queue full"
			response.LatencyMs = int(time.Since(startTime).Milliseconds())
			return response, fmt.Errorf("request queue full")
		}

		// 从队列中取出请求
		dequeuedReq, err := g.queueFactory.Dequeue(queueKey)
		if err != nil || dequeuedReq == nil {
			response.Status = "queue_error"
			response.Error = "failed to dequeue request"
			response.LatencyMs = int(time.Since(startTime).Milliseconds())
			return response, fmt.Errorf("failed to dequeue request")
		}

		// 更新请求参数
		req.RequestBody = dequeuedReq.RequestBody
		req.UserPrefs = dequeuedReq.UserPrefs
		req.CostBudget = dequeuedReq.CostBudget
		req.ComplianceReqs = dequeuedReq.ComplianceReqs
		req.AllowedKeyIDs = dequeuedReq.AllowedKeyIDs
	}

	// 3. 路由决策
	routingReq := &RoutingRequest{
		RequestID: req.RequestID,
		MerchantID: req.MerchantID,
		Model: req.Model,
		Provider: req.Provider,
		RequestBody: req.RequestBody,
		UserPrefs: req.UserPrefs,
		CostBudget: req.CostBudget,
		ComplianceReqs: req.ComplianceReqs,
		AllowedKeyIDs: req.AllowedKeyIDs,
	}

	decision, err := g.pipeline.Execute(ctx, routingReq)
	if err != nil {
		response.Status = "routing_failed"
		response.Error = fmt.Sprintf("routing failed: %v", err)
		response.LatencyMs = int(time.Since(startTime).Milliseconds())
		return response, fmt.Errorf("routing failed: %v", err)
	}

	// 4. 执行请求
	response.APIKeyID = decision.SelectedAPIKeyID
	response.Provider = decision.SelectedProvider

	// 这里应该调用实际的API执行逻辑
	// 暂时模拟一个成功的响应
	response.Status = "success"
	response.ResponseBody = map[string]interface{}{
		"id": req.RequestID,
		"object": "chat.completion",
		"created": time.Now().Unix(),
		"model": req.Model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role": "assistant",
					"content": "This is a simulated response from the unified gateway",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens": 10,
			"completion_tokens": 20,
			"total_tokens": 30,
		},
	}

	response.TokensUsed = 30
	response.Cost = 0.00015
	response.LatencyMs = int(time.Since(startTime).Milliseconds())

	// 5. 更新统计信息
	g.updateStats(response)

	return response, nil
}

func (g *UnifiedGatewayImpl) updateStats(response *GatewayResponse) {
	g.stats.Requests++

	if response.Status == "success" {
		g.stats.Successes++
	} else {
		g.stats.Failures++
	}

	g.stats.AvgLatencyMs = (g.stats.AvgLatencyMs*float64(g.stats.Requests-1) + float64(response.LatencyMs)) / float64(g.stats.Requests)
	if float64(response.LatencyMs) > g.stats.MaxLatencyMs {
		g.stats.MaxLatencyMs = float64(response.LatencyMs)
	}

	if response.TokensUsed > 0 {
		g.stats.TotalTokens += int64(response.TokensUsed)
	}

	if response.Cost > 0 {
		g.stats.TotalCost += response.Cost
	}
}

func (g *UnifiedGatewayImpl) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"requests": g.stats.Requests,
		"successes": g.stats.Successes,
		"failures": g.stats.Failures,
		"success_rate": float64(g.stats.Successes) / float64(g.stats.Requests+1),
		"avg_latency_ms": g.stats.AvgLatencyMs,
		"max_latency_ms": g.stats.MaxLatencyMs,
		"total_tokens": g.stats.TotalTokens,
		"total_cost": g.stats.TotalCost,
		"last_reset": g.stats.LastReset,
	}
}

func (g *UnifiedGatewayImpl) ResetStats() {
	g.stats.Requests = 0
	g.stats.Successes = 0
	g.stats.Failures = 0
	g.stats.AvgLatencyMs = 0
	g.stats.MaxLatencyMs = 0
	g.stats.TotalTokens = 0
	g.stats.TotalCost = 0
	g.stats.LastReset = time.Now()
}

func (g *UnifiedGatewayImpl) SetQueueEnabled(enabled bool) {
	g.enableQueue = enabled
}

func (g *UnifiedGatewayImpl) SetQueueMaxSize(maxSize int) {
	g.queueMaxSize = maxSize
}

func (g *UnifiedGatewayImpl) SetRateLimit(rate, burst int) {
	g.rateLimitRate = rate
	g.rateLimitBurst = burst
}
