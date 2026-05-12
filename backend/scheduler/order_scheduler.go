package scheduler

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/services"
)

type OrderScheduler struct {
	interval time.Duration
	timeout  time.Duration
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewOrderScheduler(interval, timeout time.Duration) *OrderScheduler {
	return &OrderScheduler{
		interval: interval,
		timeout:  timeout,
		stopChan: make(chan struct{}),
	}
}

func (s *OrderScheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Printf("Order scheduler started: checking every %v for orders older than %v", s.interval, s.timeout)
}

func (s *OrderScheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Order scheduler stopped")
}

func (s *OrderScheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.cancelExpiredOrders()
		}
	}
}

func (s *OrderScheduler) cancelExpiredOrders() {
	db := config.GetDB()
	if db == nil {
		log.Printf("Scheduler: database not available")
		return
	}

	ctx := context.Background()
	cutoffTime := time.Now().Add(-s.timeout)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Scheduler: failed to start transaction: %v", err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT o.id, o.sku_id, o.quantity 
		FROM orders o 
		WHERE o.status = 'pending' 
		AND o.created_at < $1`,
		cutoffTime,
	)
	if err != nil {
		log.Printf("Scheduler: failed to query expired orders: %v", err)
		return
	}

	type orderInfo struct {
		id       int
		skuID    sql.NullInt64
		quantity int
	}

	var orders []orderInfo
	for rows.Next() {
		var o orderInfo
		if err := rows.Scan(&o.id, &o.skuID, &o.quantity); err != nil {
			log.Printf("Scheduler: failed to scan order: %v", err)
			continue
		}
		orders = append(orders, o)
	}
	rows.Close()

	if len(orders) == 0 {
		return
	}

	log.Printf("Scheduler: found %d expired orders to cancel", len(orders))

	for _, o := range orders {
		_, err := tx.ExecContext(ctx,
			"UPDATE orders SET status = $1 WHERE id = $2",
			"canceled", o.id,
		)
		if err != nil {
			log.Printf("Scheduler: failed to cancel order %d: %v", o.id, err)
			continue
		}

		if o.skuID.Valid && o.skuID.Int64 > 0 {
			_, err = tx.ExecContext(ctx,
				"UPDATE skus SET stock = CASE WHEN stock = -1 THEN -1 ELSE stock + $1 END WHERE id = $2",
				o.quantity, o.skuID.Int64,
			)
			if err != nil {
				log.Printf("Scheduler: failed to restore stock for order %d: %v", o.id, err)
			}
		}

		if _, err := tx.ExecContext(ctx, services.SQLRestoreFlashSaleStockFromOrderItems, o.id); err != nil {
			log.Printf("Scheduler: failed to restore flash sale stock for order %d: %v", o.id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Scheduler: failed to commit transaction: %v", err)
		return
	}

	log.Printf("Scheduler: successfully canceled %d expired orders", len(orders))
}

func (s *OrderScheduler) CancelOrderManually(orderID int) error {
	db := config.GetDB()
	if db == nil {
		return apperrors.ErrDatabaseError
	}

	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.NewAppError(
			"TRANSACTION_START_FAILED",
			"Failed to start transaction",
			500,
			err,
		)
	}
	defer tx.Rollback()

	var skuID sql.NullInt64
	var quantity int
	err = tx.QueryRowContext(ctx,
		"SELECT sku_id, quantity FROM orders WHERE id = $1 AND status = 'pending'",
		orderID,
	).Scan(&skuID, &quantity)
	if err != nil {
		return apperrors.ErrOrderNotFound
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE orders SET status = 'canceled' WHERE id = $1",
		orderID,
	)
	if err != nil {
		return apperrors.NewAppError(
			"ORDER_CANCEL_FAILED",
			"Failed to cancel order",
			500,
			err,
		)
	}

	if skuID.Valid && skuID.Int64 > 0 {
		_, err = tx.ExecContext(ctx,
			"UPDATE skus SET stock = CASE WHEN stock = -1 THEN -1 ELSE stock + $1 END WHERE id = $2",
			quantity, skuID.Int64,
		)
		if err != nil {
			return apperrors.NewAppError(
				"STOCK_RESTORE_FAILED",
				"Failed to restore stock",
				500,
				err,
			)
		}
	}

	return tx.Commit()
}
