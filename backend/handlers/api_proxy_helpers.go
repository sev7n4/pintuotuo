package handlers

import (
	"database/sql"
)

func nullInt64Arg(n sql.NullInt64) interface{} {
	if n.Valid {
		return n.Int64
	}
	return nil
}

func nullFloat64Arg(n sql.NullFloat64) interface{} {
	if n.Valid {
		return n.Float64
	}
	return nil
}
