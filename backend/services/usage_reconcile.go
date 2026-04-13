package services

import (
	"database/sql"
	"math"
)

const usageReconcileEpsilon = 1e-6

// SQLBillableTokensPerLogRow is a PostgreSQL expression for billable platform tokens per api_usage_logs row.
const SQLBillableTokensPerLogRow = `COALESCE(token_usage::bigint, (input_tokens + output_tokens)::bigint)`

// ReconcileUserUsage returns the sum of billable tokens from api_usage_logs and the sum of -amount for usage
// token_transactions for the same user (both platform Token units).
func ReconcileUserUsage(db *sql.DB, userID int) (usageLogSum, usageTxSum float64, err error) {
	err = db.QueryRow(`
		SELECT
			COALESCE((SELECT SUM(`+SQLBillableTokensPerLogRow+`) FROM api_usage_logs WHERE user_id = $1), 0),
			COALESCE((SELECT SUM(-amount) FROM token_transactions WHERE user_id = $1 AND type = 'usage'), 0)
	`, userID).Scan(&usageLogSum, &usageTxSum)
	return
}

// UsageReconcileOK reports whether the two sums match within a small epsilon.
func UsageReconcileOK(usageLogSum, usageTxSum float64) bool {
	return math.Abs(usageLogSum-usageTxSum) < usageReconcileEpsilon
}
