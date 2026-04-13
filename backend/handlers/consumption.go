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
	// C 端不返回 cost（内部/商户记账）；用户可见扣减口径为 输入+输出 tokens
	dataQuery := "SELECT id, request_id, provider, model, method, path, status_code, latency_ms, input_tokens, output_tokens, created_at " + baseQuery + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
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
		var requestID, provider, model, method, path string
		var createdAt time.Time

		err := rows.Scan(&id, &requestID, &provider, &model, &method, &path, &statusCode, &latencyMs, &inputTokens, &outputTokens, &createdAt)
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
		TotalRequests       int   `json:"total_requests"`
		TotalTokenDeduction int64 `json:"total_token_deduction"` // SUM(input+output)，与用户可见「扣减」一致
		AvgLatencyMs        int   `json:"avg_latency_ms"`
	}

	statsQuery := "SELECT COUNT(*), COALESCE(SUM((input_tokens::bigint + output_tokens::bigint)), 0), COALESCE(AVG(latency_ms), 0) " + baseQuery
	db.QueryRow(statsQuery, args...).Scan(&stats.TotalRequests, &stats.TotalTokenDeduction, &stats.AvgLatencyMs)

	providerQuery := "SELECT provider, COUNT(*) as count, COALESCE(SUM((input_tokens::bigint + output_tokens::bigint)), 0) as tokens " + baseQuery + " GROUP BY provider ORDER BY tokens DESC"
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
		var tokens int64
		if err := rows.Scan(&p, &count, &tokens); err != nil {
			continue
		}
		byProvider = append(byProvider, map[string]interface{}{
			"provider": p,
			"count":    count,
			"tokens":   tokens,
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

	// C 端「模型对比」气泡图：按 provider + model 聚合，与用户可见扣减口径一致（输入+输出）
	modelComparison := make([]map[string]interface{}, 0)
	mcQuery := "SELECT provider, TRIM(COALESCE(model,'')) AS model, COUNT(*)::bigint, " +
		"COALESCE(SUM((input_tokens::bigint + output_tokens::bigint)), 0), " +
		"COALESCE(AVG((input_tokens::bigint + output_tokens::bigint))::float8, 0), " +
		"COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY latency_ms), 0)::float8, " +
		"COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms), 0)::float8, " +
		"COALESCE(SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END)::float8 / NULLIF(COUNT(*), 0), 0) " +
		baseQuery + " GROUP BY provider, TRIM(COALESCE(model,'')) " +
		"HAVING TRIM(COALESCE(model,'')) <> '' " +
		"ORDER BY SUM((input_tokens::bigint + output_tokens::bigint)) DESC NULLS LAST LIMIT 60"
	mcRows, errMC := db.Query(mcQuery, args...)
	if errMC == nil && mcRows != nil {
		defer mcRows.Close()
		for mcRows.Next() {
			var prov, m string
			var reqCount int64
			var totalDed int64
			var avgDed, p50, p95, succ float64
			if err := mcRows.Scan(&prov, &m, &reqCount, &totalDed, &avgDed, &p50, &p95, &succ); err != nil {
				continue
			}
			modelComparison = append(modelComparison, map[string]interface{}{
				"provider":              prov,
				"model":                 m,
				"request_count":         reqCount,
				"total_token_deduction": totalDed,
				"avg_token_deduction":   avgDed,
				"latency_p50_ms":        p50,
				"latency_p95_ms":        p95,
				"success_rate":          succ,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":            stats,
		"by_provider":      byProvider,
		"models_in_range":  distinctModels,
		"model_comparison": modelComparison,
	})
}
