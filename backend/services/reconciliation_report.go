package services

import (
	"database/sql"
	"fmt"
	"time"
)

// GlobalUsageLedgerMatch compares total billable tokens from api_usage_logs with total usage deductions
// in token_transactions (type=usage), both in platform Token units — same口径 as user balance / 消费明细.
func GlobalUsageLedgerMatch(db *sql.DB) (usageLogTotal, usageTxTotal float64, err error) {
	err = db.QueryRow(`
		SELECT
			COALESCE((SELECT SUM(`+SQLBillableTokensPerLogRow+`) FROM api_usage_logs), 0),
			COALESCE((SELECT SUM(-amount) FROM token_transactions WHERE type = 'usage'), 0)
	`).Scan(&usageLogTotal, &usageTxTotal)
	return
}

// UsageDriftRow is one user where per-user sums diverge beyond UsageReconcileEpsilon.
type UsageDriftRow struct {
	UserID int     `json:"user_id"`
	LogSum float64 `json:"log_sum"`
	TxSum  float64 `json:"tx_sum"`
	Delta  float64 `json:"delta"`
}

const exportDriftMaxRows = 50000

func queryUsageDriftPage(db *sql.DB, limit, offset int) ([]UsageDriftRow, error) {
	rows, err := db.Query(`
		WITH log AS (
			SELECT user_id, SUM(`+SQLBillableTokensPerLogRow+`) AS c FROM api_usage_logs GROUP BY user_id
		),
		tx AS (
			SELECT user_id, SUM(-amount) AS c FROM token_transactions WHERE type = 'usage' GROUP BY user_id
		),
		uids AS (
			SELECT user_id FROM log
			UNION
			SELECT user_id FROM tx
		)
		SELECT uids.user_id,
			COALESCE(l.c, 0),
			COALESCE(t.c, 0),
			COALESCE(l.c, 0) - COALESCE(t.c, 0)
		FROM uids
		LEFT JOIN log l ON l.user_id = uids.user_id
		LEFT JOIN tx t ON t.user_id = uids.user_id
		WHERE ABS(COALESCE(l.c, 0) - COALESCE(t.c, 0)) > $1
		ORDER BY ABS(COALESCE(l.c, 0) - COALESCE(t.c, 0)) DESC
		LIMIT $2 OFFSET $3
	`, usageReconcileEpsilon, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []UsageDriftRow
	for rows.Next() {
		var r UsageDriftRow
		if err := rows.Scan(&r.UserID, &r.LogSum, &r.TxSum, &r.Delta); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if out == nil {
		out = []UsageDriftRow{}
	}
	return out, rows.Err()
}

// ListUsageDriftUsers returns users where per-user SUM(billable tokens in api_usage_logs) != SUM(-amount for usage).
func ListUsageDriftUsers(db *sql.DB, limit, offset int) ([]UsageDriftRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	return queryUsageDriftPage(db, limit, offset)
}

// ListUsageDriftUsersForExport returns up to exportDriftMaxRows drift rows for CSV download (admin only).
func ListUsageDriftUsersForExport(db *sql.DB) ([]UsageDriftRow, error) {
	return queryUsageDriftPage(db, exportDriftMaxRows, 0)
}

// CountUsageDriftUsers returns how many users have a non-zero drift beyond epsilon.
func CountUsageDriftUsers(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`
		WITH log AS (
			SELECT user_id, SUM(`+SQLBillableTokensPerLogRow+`) AS c FROM api_usage_logs GROUP BY user_id
		),
		tx AS (
			SELECT user_id, SUM(-amount) AS c FROM token_transactions WHERE type = 'usage' GROUP BY user_id
		),
		uids AS (
			SELECT user_id FROM log
			UNION
			SELECT user_id FROM tx
		)
		SELECT COUNT(*) FROM (
			SELECT uids.user_id
			FROM uids
			LEFT JOIN log l ON l.user_id = uids.user_id
			LEFT JOIN tx t ON t.user_id = uids.user_id
			WHERE ABS(COALESCE(l.c, 0) - COALESCE(t.c, 0)) > $1
		) drift
	`, usageReconcileEpsilon).Scan(&n)
	return n, err
}

// GMVReportSummary is retail order GMV in CNY (orders.total_price), separate from internal Token usage.
type GMVReportSummary struct {
	OrderCount int     `json:"order_count"`
	GMVCNY     float64 `json:"gmv_cny"`
	Currency   string  `json:"currency"`
	StartDate  *string `json:"start_date,omitempty"`
	EndDate    *string `json:"end_date,omitempty"`
}

// GetGMVReportSummary sums total_price for paid/completed orders; optional date range on orders.created_at.
func GetGMVReportSummary(db *sql.DB, start, end *time.Time) (GMVReportSummary, error) {
	var out GMVReportSummary
	base := `SELECT COUNT(*), COALESCE(SUM(total_price), 0) FROM orders WHERE status IN ('paid', 'completed')`
	var err error
	switch {
	case start != nil && end != nil:
		err = db.QueryRow(base+` AND created_at >= $1 AND created_at <= $2`, *start, *end).Scan(&out.OrderCount, &out.GMVCNY)
		s := start.Format("2006-01-02")
		e := end.Format("2006-01-02")
		out.StartDate = &s
		out.EndDate = &e
	case start != nil:
		err = db.QueryRow(base+` AND created_at >= $1`, *start).Scan(&out.OrderCount, &out.GMVCNY)
		s := start.Format("2006-01-02")
		out.StartDate = &s
	case end != nil:
		err = db.QueryRow(base+` AND created_at <= $1`, *end).Scan(&out.OrderCount, &out.GMVCNY)
		e := end.Format("2006-01-02")
		out.EndDate = &e
	default:
		err = db.QueryRow(base).Scan(&out.OrderCount, &out.GMVCNY)
	}
	out.Currency = "CNY"
	return out, err
}

// GMVTrendPoint is GMV aggregated for one period (day / calendar week / month).
type GMVTrendPoint struct {
	Period     string  `json:"period"`
	OrderCount int     `json:"order_count"`
	GMVCNY     float64 `json:"gmv_cny"`
}

// GetGMVTrends returns paid/completed order GMV series by granularity within [start, end] on orders.created_at.
func GetGMVTrends(db *sql.DB, granularity string, start, end time.Time) ([]GMVTrendPoint, error) {
	var periodExpr, groupByExpr string
	switch granularity {
	case "week":
		periodExpr = "to_char(date_trunc('week', o.created_at), 'YYYY-MM-DD')"
		groupByExpr = "date_trunc('week', o.created_at)"
	case "month":
		periodExpr = "to_char(date_trunc('month', o.created_at), 'YYYY-MM')"
		groupByExpr = "date_trunc('month', o.created_at)"
	default:
		periodExpr = "to_char((o.created_at)::date, 'YYYY-MM-DD')"
		groupByExpr = "(o.created_at)::date"
	}

	query := fmt.Sprintf(`
		SELECT %s AS period,
			COUNT(*)::int,
			COALESCE(SUM(o.total_price), 0)::float8
		FROM orders o
		WHERE o.status IN ('paid', 'completed')
		  AND o.created_at >= $1 AND o.created_at <= $2
		GROUP BY %s
		ORDER BY %s ASC
	`, periodExpr, groupByExpr, groupByExpr)

	rows, err := db.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []GMVTrendPoint
	for rows.Next() {
		var p GMVTrendPoint
		if err := rows.Scan(&p.Period, &p.OrderCount, &p.GMVCNY); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []GMVTrendPoint{}
	}
	return out, rows.Err()
}
