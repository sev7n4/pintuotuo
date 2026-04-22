package handlers

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/metrics"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

// executeProxyChatCompletionStreamFromUpstream 在已拿到上游 HTTP 200 与 SSE body 时，将事件透传给客户端并结算。
// billProv/billModel 为实际计费用的 provider/model（可与 req 中原始请求不同，例如命中备用链时）。
// 仅用于 OpenAI 兼容路径（/chat/completions）。
func executeProxyChatCompletionStreamFromUpstream(
	c *gin.Context,
	upstream *http.Response,
	requestID string,
	userIDInt int,
	req APIProxyRequest,
	billProv string,
	billModel string,
	requestPath string,
	startTime time.Time,
	db *sql.DB,
	billingEngine *billing.BillingEngine,
	winningKey models.MerchantAPIKey,
	merchantID int,
	strictPricingVID *int,
	selectedStrategy string,
	smartCandidatesJSON []byte,
	effectivePolicySource string,
	decisionStart time.Time,
	traceSpan *services.LLMTraceSpan,
	strategySnapshot strategyRuntimeSnapshot,
	retryCountTotal int,
) {
	defer upstream.Body.Close()

	if ct := upstream.Header.Get("Content-Type"); ct != "" {
		c.Writer.Header().Set("Content-Type", ct)
	} else {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
	}
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	c.Writer.WriteHeader(http.StatusOK)
	flusher, okFlush := c.Writer.(http.Flusher)

	br := bufio.NewReader(upstream.Body)
	var streamedBytes int
	var lastUsage *APIUsage

	for {
		line, err := br.ReadBytes('\n')
		if len(line) > 0 {
			streamedBytes += len(line)
			trim := bytes.TrimSpace(line)
			if bytes.HasPrefix(trim, []byte("data: ")) {
				payload := bytes.TrimSpace(trim[6:])
				if !bytes.Equal(payload, []byte("[DONE]")) && len(payload) > 0 {
					var chunk map[string]interface{}
					if json.Unmarshal(payload, &chunk) == nil {
						if u := extractUsageFromStreamChunk(chunk); u != nil && u.TotalTokens > 0 {
							lastUsage = u
						}
					}
				}
			}
			if _, werr := c.Writer.Write(line); werr != nil {
				logger.LogWarn(c.Request.Context(), "api_proxy", "stream write interrupted", map[string]interface{}{
					"request_id": requestID, "error": werr.Error(),
				})
				break
			}
			if okFlush {
				flusher.Flush()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.LogWarn(c.Request.Context(), "api_proxy", "stream read error", map[string]interface{}{
				"request_id": requestID, "error": err.Error(),
			})
			break
		}
	}

	latency := int(time.Since(startTime).Milliseconds())
	services.GetSmartRouter().RecordRequestResult(winningKey.ID, true)
	recordHealthCheckerProxyOutcome(c, winningKey.ID, true, startTime)
	traceSpan.SetStatusCode(http.StatusOK)

	inputTokens := estimateInputTokens(req.Messages)
	outputTokens := streamedBytes / 4
	if outputTokens < 1 {
		outputTokens = 1
	}
	if lastUsage != nil && lastUsage.TotalTokens > 0 {
		inputTokens = lastUsage.PromptTokens
		outputTokens = lastUsage.CompletionTokens
	}

	tokenUsage := billingEngine.CalculateTokenUsage(inputTokens, outputTokens)
	cost, cerr := calculateTokenCost(db, userIDInt, billProv, billModel, inputTokens, outputTokens, strictPricingVID)
	if cerr != nil {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
		logger.LogError(context.Background(), "api_proxy", "stream token cost resolution failed", cerr, map[string]interface{}{
			"user_id": userIDInt, "provider": billProv, "model": billModel, "request_id": requestID,
		})
		return
	}

	if tokenUsage > 0 {
		if settleErr := billingEngine.SettlePreDeduct(userIDInt, requestID, tokenUsage); settleErr != nil {
			logger.LogError(context.Background(), "api_proxy", "stream settle pre-deduct failed", settleErr, map[string]interface{}{
				"user_id": userIDInt, "token_usage": tokenUsage, "request_id": requestID,
			})
		}
	} else {
		billingEngine.CancelPreDeduct(userIDInt, requestID)
	}

	if cost > 0 {
		billReq := req
		billReq.Provider = billProv
		billReq.Model = billModel
		logMerchantSKUID, logProcurementCNY := resolveMerchantSKUProcurementForLog(db, billReq, winningKey.ID, merchantID, inputTokens, outputTokens)
		tx, err := db.Begin()
		if err == nil {
			_, updateErr := tx.Exec(
				"UPDATE merchant_api_keys SET quota_used = quota_used + $1, last_used_at = $2 WHERE id = $3",
				cost, time.Now(), winningKey.ID,
			)
			err = updateErr
			if err == nil {
				_, err = tx.Exec(
					"INSERT INTO api_usage_logs (user_id, key_id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, cost, token_usage, merchant_sku_id, procurement_cost_cny) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)",
					userIDInt, winningKey.ID, requestID, billProv, billModel, "POST", requestPath, http.StatusOK, latency, inputTokens, outputTokens, cost, tokenUsage, nullInt64Arg(logMerchantSKUID), nullFloat64Arg(logProcurementCNY),
				)
			}
			if err != nil {
				tx.Rollback()
				logger.LogError(context.Background(), "api_proxy", "stream transaction rollback", err, map[string]interface{}{
					"user_id": userIDInt, "request_id": requestID,
				})
			} else {
				tx.Commit()
				logger.LogInfo(context.Background(), "api_proxy", "stream request completed", map[string]interface{}{
					"user_id": userIDInt, "provider": billProv, "model": billModel,
					"input_tokens": inputTokens, "output_tokens": outputTokens, "cost": cost,
					"latency_ms": latency, "request_id": requestID, "usage_from_stream": lastUsage != nil,
				})
				metrics.RecordOrderCreation("completed", int64(cost*100), "USD")
			}
		}
		ctx := context.Background()
		cache.Delete(ctx, cache.TokenBalanceKey(userIDInt))
		cache.Delete(ctx, cache.ComputePointBalanceKey(userIDInt))
	}

	decisionPayload := buildRoutingDecisionPayload(smartCandidatesJSON, strategySnapshot, effectivePolicySource)
	_ = insertRoutingDecision(db, requestID, userIDInt, req, selectedStrategy, decisionPayload, winningKey.ID, int(time.Since(decisionStart).Milliseconds()), retryCountTotal)
}

func extractUsageFromStreamChunk(chunk map[string]interface{}) *APIUsage {
	raw, ok := chunk["usage"]
	if !ok || raw == nil {
		return nil
	}
	u, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	pt := intFromJSONNumber(u["prompt_tokens"])
	ct := intFromJSONNumber(u["completion_tokens"])
	tt := intFromJSONNumber(u["total_tokens"])
	if tt == 0 {
		tt = pt + ct
	}
	if tt == 0 {
		return nil
	}
	return &APIUsage{PromptTokens: pt, CompletionTokens: ct, TotalTokens: tt}
}

func intFromJSONNumber(v interface{}) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return 0
	}
}
