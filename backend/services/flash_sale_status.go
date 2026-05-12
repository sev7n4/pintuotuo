package services

import (
	"database/sql"
	"time"
)

// DBExecutor supports *sql.DB and *sql.Tx for flash sale status transitions.
type DBExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// PromoteFlashSaleStatuses sets upcoming→active when in time window, and active/upcoming→ended when past end.
func PromoteFlashSaleStatuses(db DBExecutor, now time.Time) (changed bool, err error) {
	res1, err := db.Exec(`
		UPDATE flash_sales SET status = 'active', updated_at = CURRENT_TIMESTAMP
		WHERE status = 'upcoming' AND start_time <= $1 AND end_time > $1`,
		now)
	if err != nil {
		return false, err
	}
	n1, _ := res1.RowsAffected()
	res2, err := db.Exec(`
		UPDATE flash_sales SET status = 'ended', updated_at = CURRENT_TIMESTAMP
		WHERE status IN ('upcoming', 'active') AND end_time <= $1`,
		now)
	if err != nil {
		return false, err
	}
	n2, _ := res2.RowsAffected()
	return n1+n2 > 0, nil
}
