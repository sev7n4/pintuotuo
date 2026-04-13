package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

const allProviders = "all"

func GetConsumptionRecords(c *gin.Context) {
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

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	provider := c.Query("provider")
	modelFilter := c.Query("model")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	baseQuery := "FROM api_usage_logs WHERE user_id = $1"
	args := []interface{}{userIDInt}
	argIndex := 2

	if startDate != "" {
		baseQuery += " AND created_at >= $" + strconv.Itoa(argIndex)
		args = append(args, startDate+" 00:00:00")
		argIndex++
	}
	if endDate != "" {
		baseQuery += " AND created_at <= $" + strconv.Itoa(argIndex)
		args = append(args, endDate+" 23:59:59")
		argIndex++
	}
	if provider != "" && provider != allProviders {
		baseQuery += " AND provider = $" + strconv.Itoa(argIndex)
		args = append(args, provider)
		argIndex++
	}
	if modelFilter != "" {
		baseQuery += " AND model = $" + strconv.Itoa(argIndex)
		args = append(args, modelFilter)
		argIndex++
	}

	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	db.QueryRow(countQuery, args...).Scan(&total)

	offset := (page - 1) * pageSize
	dataQuery := "SELECT id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, cost, COALESCE(token_usage, input_tokens + output_tokens) as token_usage, created_at " + baseQuery + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := db.Query(dataQuery, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, statusCode, latencyMs, inputTokens, outputTokens int
		var tokenUsage int64
		var requestID, provider, model, method, path string
		var cost float64
		var createdAt time.Time

		err := rows.Scan(&id, &requestID, &provider, &model, &method, &path, &statusCode, &latencyMs, &inputTokens, &outputTokens, &cost, &tokenUsage, &createdAt)
		if err != nil {
			continue
		}

		records = append(records, map[string]interface{}{
			"id":            id,
			"request_id":    requestID,
			"provider":      provider,
			"model":         model,
			"method":        method,
			"path":          path,
			"status_code":   statusCode,
			"latency_ms":    latencyMs,
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
			"token_usage":   tokenUsage,
			"cost":          cost,
			"created_at":    createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetConsumptionStats(c *gin.Context) {
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

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	provider := c.Query("provider")
	modelFilter := c.Query("model")

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	baseQuery := "FROM api_usage_logs WHERE user_id = $1"
	args := []interface{}{userIDInt}
	argIndex := 2

	if startDate != "" {
		baseQuery += " AND created_at >= $" + strconv.Itoa(argIndex)
		args = append(args, startDate+" 00:00:00")
		argIndex++
	}
	if endDate != "" {
		baseQuery += " AND created_at <= $" + strconv.Itoa(argIndex)
		args = append(args, endDate+" 23:59:59")
		argIndex++
	}
	if provider != "" && provider != allProviders {
		baseQuery += " AND provider = $" + strconv.Itoa(argIndex)
		args = append(args, provider)
		argIndex++
	}
	if modelFilter != "" {
		baseQuery += " AND model = $" + strconv.Itoa(argIndex)
		args = append(args, modelFilter)
		argIndex++
	}

	var stats struct {
		TotalRequests int     `json:"total_requests"`
		TotalTokens   int64   `json:"total_tokens"`
		TotalCost     float64 `json:"total_cost"`
		AvgLatencyMs  int     `json:"avg_latency_ms"`
	}

	// 总 Tokens：优先 token_usage，否则 input+output（与列表行一致）
	statsQuery := "SELECT COUNT(*), COALESCE(SUM(COALESCE(token_usage, (input_tokens + output_tokens)::bigint)), 0), COALESCE(SUM(cost), 0), COALESCE(AVG(latency_ms), 0) " + baseQuery
	db.QueryRow(statsQuery, args...).Scan(&stats.TotalRequests, &stats.TotalTokens, &stats.TotalCost, &stats.AvgLatencyMs)

	providerQuery := "SELECT provider, COUNT(*) as count, SUM(cost) as cost " + baseQuery + " GROUP BY provider ORDER BY cost DESC"
	rows, err := db.Query(providerQuery, args...)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	byProvider := make([]map[string]interface{}, 0)
	for rows.Next() {
		var p string
		var count int
		var cost float64
		if err := rows.Scan(&p, &count, &cost); err != nil {
			continue
		}
		byProvider = append(byProvider, map[string]interface{}{
			"provider": p,
			"count":    count,
			"cost":     cost,
		})
	}

	distinctModels := make([]string, 0)
	dModelQ := "SELECT DISTINCT TRIM(model) AS m " + baseQuery + " AND TRIM(COALESCE(model,'')) <> '' ORDER BY m LIMIT 200"
	dmRows, errDM := db.Query(dModelQ, args...)
	if errDM == nil {
		defer dmRows.Close()
		for dmRows.Next() {
			var m string
			if err := dmRows.Scan(&m); err == nil && m != "" {
				distinctModels = append(distinctModels, m)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":          stats,
		"by_provider":    byProvider,
		"models_in_range": distinctModels,
	})
}
