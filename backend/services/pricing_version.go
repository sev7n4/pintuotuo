package services

import (
	"database/sql"
)

const pricingVersionCodeBaseline = "baseline"

// BaselinePricingVersionID returns the id of the baseline retail pricing version (migration 045), or invalid if missing.
func BaselinePricingVersionID(q interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}) sql.NullInt64 {
	var id sql.NullInt64
	err := q.QueryRow(
		`SELECT id FROM pricing_versions WHERE code = $1 LIMIT 1`,
		pricingVersionCodeBaseline,
	).Scan(&id)
	if err != nil {
		return sql.NullInt64{}
	}
	return id
}
