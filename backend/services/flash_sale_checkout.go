package services

import (
	"database/sql"
	"net/http"
	"time"

	apperrors "github.com/pintuotuo/backend/errors"
)

// ReserveFlashSaleLine validates an active flash window, per-user limit, and increments stock_sold (call inside order tx).
func ReserveFlashSaleLine(tx *sql.Tx, userID int, flashSaleID, skuID, qty int, now time.Time) (unitPrice float64, appErr *apperrors.AppError) {
	if qty < 1 {
		return 0, apperrors.ErrInvalidRequest
	}

	var flashPrice float64
	var stockLimit, stockSold, perUserLimit int
	var status string
	var startTime, endTime time.Time

	err := tx.QueryRow(`
		SELECT fsp.flash_price, fsp.stock_limit, fsp.stock_sold, fsp.per_user_limit,
		       fs.status, fs.start_time, fs.end_time
		  FROM flash_sale_products fsp
		  JOIN flash_sales fs ON fs.id = fsp.flash_sale_id
		 WHERE fsp.flash_sale_id = $1 AND fsp.sku_id = $2
		 FOR UPDATE OF fsp`,
		flashSaleID, skuID,
	).Scan(&flashPrice, &stockLimit, &stockSold, &perUserLimit, &status, &startTime, &endTime)
	if err == sql.ErrNoRows {
		return 0, apperrors.NewAppError(
			"FLASH_SALE_SKU_NOT_IN_SALE",
			"秒杀场次不包含该商品",
			http.StatusBadRequest,
			nil,
		)
	}
	if err != nil {
		return 0, apperrors.ErrDatabaseError
	}

	if status != routingStrategyStatusActive {
		return 0, apperrors.NewAppError(
			"FLASH_SALE_NOT_ACTIVE",
			"秒杀未在进行中",
			http.StatusBadRequest,
			nil,
		)
	}
	if now.Before(startTime) || !endTime.After(now) {
		return 0, apperrors.NewAppError(
			"FLASH_SALE_OUT_OF_WINDOW",
			"不在秒杀有效期内",
			http.StatusBadRequest,
			nil,
		)
	}
	if stockSold+qty > stockLimit {
		return 0, apperrors.NewAppError(
			"FLASH_SALE_SOLD_OUT",
			"秒杀库存不足",
			http.StatusBadRequest,
			nil,
		)
	}

	limit := perUserLimit
	if limit < 1 {
		limit = 1
	}

	var bought int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(oi.quantity), 0)::int
		  FROM order_items oi
		  JOIN orders o ON o.id = oi.order_id
		 WHERE o.user_id = $1
		   AND oi.flash_sale_id = $2
		   AND oi.sku_id = $3
		   AND o.status IN ('pending', 'paid', 'completed')`,
		userID, flashSaleID, skuID,
	).Scan(&bought)
	if err != nil {
		return 0, apperrors.ErrDatabaseError
	}
	if bought+qty > limit {
		return 0, apperrors.NewAppError(
			"FLASH_SALE_PER_USER_LIMIT",
			"超过该秒杀场次的每人限购数量",
			http.StatusBadRequest,
			nil,
		)
	}

	res, err := tx.Exec(`
		UPDATE flash_sale_products
		   SET stock_sold = stock_sold + $1,
		       updated_at = CURRENT_TIMESTAMP
		 WHERE flash_sale_id = $2
		   AND sku_id = $3
		   AND stock_limit >= stock_sold + $1`,
		qty, flashSaleID, skuID)
	if err != nil {
		return 0, apperrors.ErrDatabaseError
	}
	ra, _ := res.RowsAffected()
	if ra != 1 {
		return 0, apperrors.NewAppError(
			"FLASH_SALE_SOLD_OUT",
			"秒杀库存不足",
			http.StatusBadRequest,
			nil,
		)
	}

	return flashPrice, nil
}
