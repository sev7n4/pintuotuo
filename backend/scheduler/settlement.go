package scheduler

import (
	"database/sql"
	"log"
	"time"

	"github.com/pintuotuo/backend/config"
)

type SettlementScheduler struct {
	interval  time.Duration
	stopChan  chan struct{}
	isRunning bool
}

func NewSettlementScheduler(interval time.Duration) *SettlementScheduler {
	return &SettlementScheduler{
		interval:  interval,
		stopChan:  make(chan struct{}),
		isRunning: false,
	}
}

func (s *SettlementScheduler) Start() {
	if s.isRunning {
		return
	}

	s.isRunning = true
	ticker := time.NewTicker(s.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.processSettlements()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	log.Println("Settlement scheduler started")
}

func (s *SettlementScheduler) Stop() {
	if !s.isRunning {
		return
	}

	s.isRunning = false
	close(s.stopChan)
	log.Println("Settlement scheduler stopped")
}

func (s *SettlementScheduler) processSettlements() {
	db := config.GetDB()
	if db == nil {
		log.Println("Settlement scheduler: database not available")
		return
	}

	rows, err := db.Query(
		`SELECT id, merchant_id FROM merchant_settlements WHERE status = 'pending' AND created_at < $1`,
		time.Now().Add(-24*time.Hour),
	)
	if err != nil {
		log.Printf("Settlement scheduler: failed to query pending settlements: %v", err)
		return
	}
	defer rows.Close()

	type pendingSettlement struct {
		id         int
		merchantID int
	}

	var settlements []pendingSettlement
	for rows.Next() {
		var ps pendingSettlement
		if err := rows.Scan(&ps.id, &ps.merchantID); err != nil {
			log.Printf("Settlement scheduler: failed to scan settlement: %v", err)
			continue
		}
		settlements = append(settlements, ps)
	}

	for _, settlement := range settlements {
		s.processSettlement(db, settlement.id, settlement.merchantID)
	}

	s.generateMonthlySettlements(db)
}

func (s *SettlementScheduler) processSettlement(db *sql.DB, settlementID, merchantID int) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Settlement scheduler: failed to begin transaction for settlement %d: %v", settlementID, err)
		return
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRow("SELECT status FROM merchant_settlements WHERE id = $1 FOR UPDATE", settlementID).Scan(&status)
	if err != nil {
		log.Printf("Settlement scheduler: failed to lock settlement %d: %v", settlementID, err)
		return
	}

	if status != "pending" {
		return
	}

	_, err = tx.Exec(
		"UPDATE merchant_settlements SET status = 'processing', updated_at = CURRENT_TIMESTAMP WHERE id = $1",
		settlementID,
	)
	if err != nil {
		log.Printf("Settlement scheduler: failed to update settlement %d to processing: %v", settlementID, err)
		return
	}

	if commitErr := tx.Commit(); commitErr != nil {
		log.Printf("Settlement scheduler: failed to commit transaction for settlement %d: %v", settlementID, commitErr)
		return
	}

	_, err = db.Exec(
		"UPDATE merchant_settlements SET status = 'completed', settled_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1",
		settlementID,
	)
	if err != nil {
		log.Printf("Settlement scheduler: failed to complete settlement %d: %v", settlementID, err)
		return
	}

	log.Printf("Settlement scheduler: completed settlement %d for merchant %d", settlementID, merchantID)
}

func (s *SettlementScheduler) generateMonthlySettlements(db *sql.DB) {
	now := time.Now()
	if now.Day() != 1 {
		return
	}

	lastMonth := now.AddDate(0, -1, 0)
	periodStart := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	rows, err := db.Query(
		`SELECT DISTINCT s.merchant_id 
		 FROM skus s 
		 JOIN orders o ON o.sku_id = s.id 
		 WHERE o.status = 'completed' 
		 AND o.updated_at >= $1 AND o.updated_at < $2
		 AND s.merchant_id IS NOT NULL`,
		periodStart, periodEnd,
	)
	if err != nil {
		log.Printf("Settlement scheduler: failed to query merchants for monthly settlement: %v", err)
		return
	}
	defer rows.Close()

	var merchantIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		merchantIDs = append(merchantIDs, id)
	}

	platformFeeRate := 0.05

	for _, merchantID := range merchantIDs {
		var existingID int
		err := db.QueryRow(
			"SELECT id FROM merchant_settlements WHERE merchant_id = $1 AND period_start = $2",
			merchantID, periodStart,
		).Scan(&existingID)
		if err == nil {
			continue
		}

		var totalSales float64
		db.QueryRow(
			`SELECT COALESCE(SUM(o.total_price), 0) FROM orders o 
			 JOIN skus s ON o.sku_id = s.id 
			 WHERE s.merchant_id = $1 AND o.status = 'completed' 
			 AND o.updated_at >= $2 AND o.updated_at < $3`,
			merchantID, periodStart, periodEnd,
		).Scan(&totalSales)

		if totalSales < 100 {
			continue
		}

		platformFee := totalSales * platformFeeRate
		settlementAmount := totalSales - platformFee

		_, err = db.Exec(
			`INSERT INTO merchant_settlements (merchant_id, period_start, period_end, total_sales, platform_fee, settlement_amount, status) 
			 VALUES ($1, $2, $3, $4, $5, $6, 'pending')`,
			merchantID, periodStart, periodEnd, totalSales, platformFee, settlementAmount,
		)
		if err != nil {
			log.Printf("Settlement scheduler: failed to create monthly settlement for merchant %d: %v", merchantID, err)
			continue
		}

		log.Printf("Settlement scheduler: created monthly settlement for merchant %d, amount: %.2f", merchantID, settlementAmount)
	}
}
