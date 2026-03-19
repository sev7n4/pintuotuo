package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type QueryOptimizer struct {
	db *sql.DB
}

func NewQueryOptimizer(db *sql.DB) *QueryOptimizer {
	return &QueryOptimizer{db: db}
}

func (q *QueryOptimizer) Paginate(ctx context.Context, query string, args []interface{}, page, pageSize int) (*sql.Rows, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	argCount := len(args)
	fullQuery := fmt.Sprintf("%s LIMIT $%d OFFSET $%d", query, argCount+1, argCount+2)
	args = append(args, pageSize, offset)

	return q.db.QueryContext(ctx, fullQuery, args...)
}

func (q *QueryOptimizer) Count(ctx context.Context, table string, whereClause string, args ...interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM " + table
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	var count int64
	err := q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (q *QueryOptimizer) Exists(ctx context.Context, table, column string, value interface{}) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)", table, column)

	var exists bool
	err := q.db.QueryRowContext(ctx, query, value).Scan(&exists)
	return exists, err
}

func (q *QueryOptimizer) SoftDelete(ctx context.Context, table string, id int64) error {
	query := fmt.Sprintf("UPDATE %s SET status = 'deleted', updated_at = $1 WHERE id = $2", table)
	_, err := q.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
