package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/middleware"
	"github.com/pintuotuo/backend/services"
)

// AdminGetLedgerReconciliation returns full-database billable-token totals from api_usage_logs
// vs token_transactions usage deductions (both platform Token units).
func AdminGetLedgerReconciliation(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	logSum, txSum, err := services.GlobalUsageLedgerMatch(db)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"RECONCILE_FAILED", "Failed to compute ledger reconciliation", http.StatusInternalServerError, err))
		return
	}
	delta := logSum - txSum
	c.JSON(http.StatusOK, gin.H{
		"usage_log_total": logSum,
		"usage_tx_total":  txSum,
		"delta":           delta,
		"matched":         services.UsageReconcileOK(logSum, txSum),
		"unit":            "tokens",
		"checked_at":      time.Now().UTC().Format(time.RFC3339),
	})
}

// AdminGetLedgerDrift lists users where per-user usage sums diverge (may be slow on large datasets).
func AdminGetLedgerDrift(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 500 {
		pageSize = 500
	}
	offset := (page - 1) * pageSize

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	total, err := services.CountUsageDriftUsers(db)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"RECONCILE_FAILED", "Failed to count drift users", http.StatusInternalServerError, err))
		return
	}
	rows, err := services.ListUsageDriftUsers(db, pageSize, offset)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"RECONCILE_FAILED", "Failed to list drift users", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       rows,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"checked_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// AdminExportLedgerDriftCSV streams up to 50k drift rows as UTF-8 CSV (for Excel / ops).
func AdminExportLedgerDriftCSV(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	rows, err := services.ListUsageDriftUsersForExport(db)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"RECONCILE_EXPORT_FAILED", "Failed to export drift users", http.StatusInternalServerError, err))
		return
	}
	fn := fmt.Sprintf("ledger_drift_%s.csv", time.Now().UTC().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fn))
	if _, err := c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return
	}
	w := csv.NewWriter(c.Writer)
	if err := w.Write([]string{"user_id", "log_sum", "tx_sum", "delta"}); err != nil {
		return
	}
	for _, r := range rows {
		if err := w.Write([]string{
			strconv.Itoa(r.UserID),
			strconv.FormatFloat(r.LogSum, 'f', -1, 64),
			strconv.FormatFloat(r.TxSum, 'f', -1, 64),
			strconv.FormatFloat(r.Delta, 'f', -1, 64),
		}); err != nil {
			return
		}
	}
	w.Flush()
}

// AdminPostLedgerCheck runs the same global check as GET and writes an audit log line (for cron jobs).
func AdminPostLedgerCheck(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	logSum, txSum, err := services.GlobalUsageLedgerMatch(db)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"RECONCILE_FAILED", "Failed to compute ledger reconciliation", http.StatusInternalServerError, err))
		return
	}
	ok := services.UsageReconcileOK(logSum, txSum)
	logger.LogInfo(c.Request.Context(), "reconciliation", "ledger usage check", map[string]interface{}{
		"usage_log_total": logSum,
		"usage_tx_total":  txSum,
		"delta":           logSum - txSum,
		"matched":         ok,
	})
	delta := logSum - txSum
	c.JSON(http.StatusOK, gin.H{
		"usage_log_total": logSum,
		"usage_tx_total":  txSum,
		"delta":           delta,
		"matched":         ok,
		"unit":            "tokens",
		"checked_at":      time.Now().UTC().Format(time.RFC3339),
	})
}

// AdminGetGMVReport returns CNY GMV from orders (paid/completed), separate from internal Token usage.
func AdminGetGMVReport(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startPtr, endPtr *time.Time
	if startStr != "" {
		t, err := time.ParseInLocation("2006-01-02", startStr, time.Local)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError("INVALID_DATE", "Invalid start_date", http.StatusBadRequest, err))
			return
		}
		startOf := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		startPtr = &startOf
	}
	if endStr != "" {
		t, err := time.ParseInLocation("2006-01-02", endStr, time.Local)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError("INVALID_DATE", "Invalid end_date", http.StatusBadRequest, err))
			return
		}
		endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
		endPtr = &endOfDay
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	summary, err := services.GetGMVReportSummary(db, startPtr, endPtr)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GMV_REPORT_FAILED", "Failed to load GMV report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, summary)
}

// AdminGetGMVTrends returns GMV time series (CNY) by day / week / month for paid & completed orders.
// Query: granularity=day|week|month, start_date & end_date (YYYY-MM-DD). If dates omitted, last 90 days.
func AdminGetGMVTrends(c *gin.Context) {
	if !requireAdminRole(c) {
		return
	}
	granularity := c.DefaultQuery("granularity", "day")
	switch granularity {
	case "day", "week", "month":
	default:
		granularity = "day"
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")
	now := time.Now()
	var startAt, endAt time.Time

	if startStr == "" || endStr == "" {
		endAt = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.Local)
		startAt = endAt.AddDate(0, 0, -89)
		startAt = time.Date(startAt.Year(), startAt.Month(), startAt.Day(), 0, 0, 0, 0, time.Local)
	} else {
		tStart, err := time.ParseInLocation("2006-01-02", startStr, time.Local)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError("INVALID_DATE", "Invalid start_date", http.StatusBadRequest, err))
			return
		}
		tEnd, err := time.ParseInLocation("2006-01-02", endStr, time.Local)
		if err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError("INVALID_DATE", "Invalid end_date", http.StatusBadRequest, err))
			return
		}
		if tEnd.Before(tStart) {
			middleware.RespondWithError(c, apperrors.NewAppError("INVALID_RANGE", "end_date must be on or after start_date", http.StatusBadRequest, nil))
			return
		}
		startAt = time.Date(tStart.Year(), tStart.Month(), tStart.Day(), 0, 0, 0, 0, time.Local)
		endAt = time.Date(tEnd.Year(), tEnd.Month(), tEnd.Day(), 23, 59, 59, 999999999, time.Local)
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	trends, err := services.GetGMVTrends(db, granularity, startAt, endAt)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError(
			"GMV_TRENDS_FAILED", "Failed to load GMV trends", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"currency":    "CNY",
		"granularity": granularity,
		"start_date":  startAt.Format("2006-01-02"),
		"end_date":    endAt.Format("2006-01-02"),
		"trends":      trends,
		"checked_at":  time.Now().UTC().Format(time.RFC3339),
	})
}
