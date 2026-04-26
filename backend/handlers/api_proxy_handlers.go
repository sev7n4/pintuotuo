package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

func GetAPIProviders(c *gin.Context) {
	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusOK, legacyProviderList())
		return
	}

	rows, err := db.Query(
		`SELECT code, name, api_format FROM model_providers WHERE status = 'active' ORDER BY sort_order ASC`,
	)
	if err != nil {
		c.JSON(http.StatusOK, legacyProviderList())
		return
	}
	defer rows.Close()

	providers := make([]map[string]interface{}, 0)
	for rows.Next() {
		var code, name, apiFormat string
		if scanErr := rows.Scan(&code, &name, &apiFormat); scanErr != nil {
			continue
		}
		providers = append(providers, map[string]interface{}{
			"name":         code,
			"display_name": name,
			"api_format":   apiFormat,
		})
	}
	if len(providers) == 0 {
		providers = legacyProviderList()
	}

	c.JSON(http.StatusOK, providers)
}

func GetAPIUsageStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var stats struct {
		TotalRequests int     `json:"total_requests"`
		TotalTokens   int     `json:"total_tokens"`
		TotalCost     float64 `json:"total_cost"`
		AvgLatencyMs  int     `json:"avg_latency_ms"`
	}

	db.QueryRow(
		"SELECT COUNT(*), COALESCE(SUM(input_tokens + output_tokens), 0), COALESCE(SUM(cost), 0), COALESCE(AVG(latency_ms), 0) FROM api_usage_logs WHERE user_id = $1",
		userIDInt,
	).Scan(&stats.TotalRequests, &stats.TotalTokens, &stats.TotalCost, &stats.AvgLatencyMs)

	rows, err := db.Query(
		"SELECT provider, COUNT(*) as count, SUM(cost) as cost FROM api_usage_logs WHERE user_id = $1 GROUP BY provider ORDER BY count DESC",
		userIDInt,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	byProvider := make([]map[string]interface{}, 0)
	for rows.Next() {
		var provider string
		var count int
		var cost float64
		if err := rows.Scan(&provider, &count, &cost); err != nil {
			continue
		}
		byProvider = append(byProvider, map[string]interface{}{
			"provider": provider,
			"count":    count,
			"cost":     cost,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":       stats,
		"by_provider": byProvider,
	})
}

func GetAPIRequestTrace(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userIDInt, ok := userID.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}

	requestID := strings.TrimSpace(c.Param("request_id"))
	if requestID == "" {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var usage struct {
		RequestID    string  `json:"request_id"`
		Provider     string  `json:"provider"`
		Model        string  `json:"model"`
		StatusCode   int     `json:"status_code"`
		LatencyMS    int     `json:"latency_ms"`
		InputTokens  int     `json:"input_tokens"`
		OutputTokens int     `json:"output_tokens"`
		Cost         float64 `json:"cost"`
	}
	err := db.QueryRow(
		`SELECT request_id, provider, model, status_code, latency_ms, input_tokens, output_tokens, cost
		 FROM api_usage_logs WHERE request_id = $1 AND user_id = $2 LIMIT 1`,
		requestID, userIDInt,
	).Scan(&usage.RequestID, &usage.Provider, &usage.Model, &usage.StatusCode, &usage.LatencyMS, &usage.InputTokens, &usage.OutputTokens, &usage.Cost)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.RespondWithError(c, apperrors.NewAppError(
				"TRACE_NOT_FOUND",
				"Trace not found",
				http.StatusNotFound,
				nil,
			))
			return
		}
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	var decision struct {
		StrategyUsed          string                  `json:"strategy_used"`
		Candidates            json.RawMessage         `json:"candidates"`
		StrategyRuntime       strategyRuntimeSnapshot `json:"strategy_runtime"`
		SelectedAPIKeyID      *int                    `json:"selected_api_key_id"`
		DecisionLatencyMS     int                     `json:"decision_latency_ms"`
		WasRetry              bool                    `json:"was_retry"`
		RetryCount            int                     `json:"retry_count"`
		CandidatesCount       int                     `json:"candidates_count"`
		TopCandidate          *traceTopCandidate      `json:"top_candidate,omitempty"`
		EffectivePolicySource string                  `json:"effective_policy_source"`
	}
	var selectedID sql.NullInt64
	var decisionRaw json.RawMessage
	rdErr := db.QueryRow(
		`SELECT strategy_used, COALESCE(candidates, '[]'::jsonb), selected_api_key_id, decision_latency_ms, was_retry, retry_count
		 FROM routing_decisions WHERE request_id = $1 AND user_id = $2 ORDER BY created_at DESC LIMIT 1`,
		requestID, userIDInt,
	).Scan(&decision.StrategyUsed, &decisionRaw, &selectedID, &decision.DecisionLatencyMS, &decision.WasRetry, &decision.RetryCount)
	if rdErr == nil && selectedID.Valid {
		val := int(selectedID.Int64)
		decision.SelectedAPIKeyID = &val
	}
	storedPolicySource := ""
	if rdErr == nil {
		parsedPayload, parseErr := parseRoutingDecisionPayload(decisionRaw)
		if parseErr == nil {
			decision.Candidates = parsedPayload.Candidates
			decision.StrategyRuntime = parsedPayload.StrategyRuntime
			storedPolicySource = parsedPayload.EffectivePolicySource
		} else {
			decision.Candidates = decisionRaw
		}
		decision.CandidatesCount, decision.TopCandidate = summarizeRoutingCandidatesForTrace(decision.Candidates)
		decision.EffectivePolicySource = normalizeEffectivePolicySource(storedPolicySource)
	}

	c.JSON(http.StatusOK, gin.H{
		"usage":    usage,
		"decision": decision,
	})
}
