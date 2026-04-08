package utils

import "database/sql"

// NullFloat64Ptr returns a *float64 for JSON encoding; nil means SQL NULL.
func NullFloat64Ptr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}
