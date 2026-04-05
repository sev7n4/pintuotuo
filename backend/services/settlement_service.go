package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/models"
)

type SettlementService struct {
	db *sql.DB
}

type SettlementData struct {
	MerchantID       int
	TotalSales       float64
	TotalOrders      int
	PlatformFee      float64
	SettlementAmount float64
}

type DisputeData struct {
	ID              int        `json:"id"`
	SettlementID    int        `json:"settlement_id"`
	MerchantID      int        `json:"merchant_id"`
	DisputeType     string     `json:"dispute_type"`
	DisputeReason   string     `json:"dispute_reason"`
	EvidenceURLs    []string   `json:"evidence_urls,omitempty"`
	OriginalAmount  float64    `json:"original_amount"`
	DisputedAmount  float64    `json:"disputed_amount"`
	AdjustedAmount  float64    `json:"adjusted_amount,omitempty"`
	Status          string     `json:"status"`
	HandledBy       *int       `json:"handled_by,omitempty"`
	HandledAt       *time.Time `json:"handled_at,omitempty"`
	ResolutionNotes string     `json:"resolution_notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ReconciliationData struct {
	ID                 int                    `json:"id"`
	SettlementID       int                    `json:"settlement_id"`
	OrderCountExpected int                    `json:"order_count_expected"`
	OrderCountActual   int                    `json:"order_count_actual"`
	OrderCountDiff     int                    `json:"order_count_diff"`
	UsageExpected      float64                `json:"usage_expected"`
	UsageActual        float64                `json:"usage_actual"`
	UsageDiff          float64                `json:"usage_diff"`
	AmountExpected     float64                `json:"amount_expected"`
	AmountActual       float64                `json:"amount_actual"`
	AmountDiff         float64                `json:"amount_diff"`
	HasAnomalies       bool                   `json:"has_anomalies"`
	AnomalyDetails     map[string]interface{} `json:"anomaly_details,omitempty"`
	ReconciledAt       time.Time              `json:"reconciled_at"`
	ReconciledBy       *int                   `json:"reconciled_by,omitempty"`
}

var (
	settlementService     *SettlementService
	settlementServiceOnce sync.Once
)

func GetSettlementService() *SettlementService {
	settlementServiceOnce.Do(func() {
		settlementService = &SettlementService{
			db: config.GetDB(),
		}
	})
	return settlementService
}

func (s *SettlementService) GenerateMonthlySettlements(periodStart, periodEnd time.Time) ([]models.MerchantSettlement, error) {
	ctx := context.Background()

	logger.LogInfo(ctx, "settlement_service", "Starting monthly settlement generation", map[string]interface{}{
		"period_start": periodStart.Format("2006-01-02"),
		"period_end":   periodEnd.Format("2006-01-02"),
	})

	query := `
		SELECT 
			m.id,
			COALESCE(SUM(o.total_price), 0) as total_sales,
			COUNT(o.id) as total_orders
		FROM merchants m
		LEFT JOIN orders o ON m.user_id = o.user_id 
			AND o.status = 'completed'
			AND o.created_at >= $1 
			AND o.created_at <= $2
		WHERE m.status = 'approved'
		GROUP BY m.id
		HAVING COUNT(o.id) > 0
	`

	rows, err := s.db.Query(query, periodStart, periodEnd)
	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to query merchant sales data", err, nil)
		return nil, fmt.Errorf("failed to query merchant sales: %w", err)
	}
	defer rows.Close()

	var settlements []models.MerchantSettlement
	platformFeeRate := 0.05 // 5% 平台费率

	for rows.Next() {
		var data SettlementData
		err := rows.Scan(&data.MerchantID, &data.TotalSales, &data.TotalOrders)
		if err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to scan merchant data", err, nil)
			continue
		}

		// 检查是否已存在该周期的结算
		var existingID int
		err = s.db.QueryRow(
			"SELECT id FROM merchant_settlements WHERE merchant_id = $1 AND period_start = $2 AND period_end = $3",
			data.MerchantID, periodStart, periodEnd,
		).Scan(&existingID)

		if err == nil {
			logger.LogInfo(ctx, "settlement_service", "Settlement already exists", map[string]interface{}{
				"merchant_id":   data.MerchantID,
				"settlement_id": existingID,
			})
			continue
		}

		data.PlatformFee = data.TotalSales * platformFeeRate
		data.SettlementAmount = data.TotalSales - data.PlatformFee

		var settlementID int
		err = s.db.QueryRow(
			`INSERT INTO merchant_settlements 
			 (merchant_id, period_start, period_end, total_sales, platform_fee, settlement_amount, status)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 RETURNING id`,
			data.MerchantID, periodStart, periodEnd, data.TotalSales, data.PlatformFee, data.SettlementAmount, "pending",
		).Scan(&settlementID)

		if err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to create settlement", err, map[string]interface{}{
				"merchant_id": data.MerchantID,
			})
			continue
		}

		settlement := models.MerchantSettlement{
			ID:               settlementID,
			MerchantID:       data.MerchantID,
			PeriodStart:      periodStart,
			PeriodEnd:        periodEnd,
			TotalSales:       data.TotalSales,
			PlatformFee:      data.PlatformFee,
			SettlementAmount: data.SettlementAmount,
			Status:           "pending",
		}
		settlements = append(settlements, settlement)

		logger.LogInfo(ctx, "settlement_service", "Settlement created", map[string]interface{}{
			"settlement_id":     settlementID,
			"merchant_id":       data.MerchantID,
			"total_sales":       data.TotalSales,
			"settlement_amount": data.SettlementAmount,
		})
	}

	logger.LogInfo(ctx, "settlement_service", "Monthly settlement generation completed", map[string]interface{}{
		"total_settlements": len(settlements),
	})

	return settlements, nil
}

func (s *SettlementService) MerchantConfirm(settlementID, merchantID int) error {
	ctx := context.Background()

	var dbMerchantID int
	var status string
	err := s.db.QueryRow(
		"SELECT merchant_id, status FROM merchant_settlements WHERE id = $1",
		settlementID,
	).Scan(&dbMerchantID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("settlement not found")
		}
		return fmt.Errorf("failed to query settlement: %w", err)
	}

	if dbMerchantID != merchantID {
		return fmt.Errorf("not authorized to confirm this settlement")
	}

	_, err = s.db.Exec(
		`UPDATE merchant_settlements 
		 SET merchant_confirmed = $1, 
		     merchant_confirmed_at = CURRENT_TIMESTAMP,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id = $2 AND merchant_id = $3`,
		true, settlementID, merchantID,
	)

	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to confirm settlement", err, map[string]interface{}{
			"settlement_id": settlementID,
			"merchant_id":   merchantID,
		})
		return fmt.Errorf("failed to confirm settlement: %w", err)
	}

	logger.LogInfo(ctx, "settlement_service", "Settlement confirmed by merchant", map[string]interface{}{
		"settlement_id": settlementID,
		"merchant_id":   merchantID,
	})

	return nil
}

func (s *SettlementService) FinanceApprove(settlementID, financeUserID int) error {
	ctx := context.Background()

	var merchantConfirmed, financeApproved bool
	err := s.db.QueryRow(
		"SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id = $1",
		settlementID,
	).Scan(&merchantConfirmed, &financeApproved)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("settlement not found")
		}
		return fmt.Errorf("failed to query settlement: %w", err)
	}

	if !merchantConfirmed {
		return fmt.Errorf("merchant confirmation required before finance approval")
	}

	if financeApproved {
		return fmt.Errorf("settlement already approved by finance")
	}

	_, err = s.db.Exec(
		`UPDATE merchant_settlements 
		 SET finance_approved = $1, 
		     finance_approved_at = CURRENT_TIMESTAMP,
		     finance_approved_by = $2,
		     status = 'processing',
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id = $3`,
		true, financeUserID, settlementID,
	)

	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to approve settlement", err, map[string]interface{}{
			"settlement_id":   settlementID,
			"finance_user_id": financeUserID,
		})
		return fmt.Errorf("failed to approve settlement: %w", err)
	}

	logger.LogInfo(ctx, "settlement_service", "Settlement approved by finance", map[string]interface{}{
		"settlement_id":   settlementID,
		"finance_user_id": financeUserID,
	})

	return nil
}

func (s *SettlementService) MarkAsPaid(settlementID, userID int) error {
	ctx := context.Background()

	var financeApproved bool
	err := s.db.QueryRow(
		"SELECT finance_approved FROM merchant_settlements WHERE id = $1",
		settlementID,
	).Scan(&financeApproved)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("settlement not found")
		}
		return fmt.Errorf("failed to query settlement: %w", err)
	}

	if !financeApproved {
		return fmt.Errorf("finance approval required before marking as paid")
	}

	_, err = s.db.Exec(
		`UPDATE merchant_settlements 
		 SET marked_paid_at = CURRENT_TIMESTAMP,
		     marked_paid_by = $1,
		     status = 'completed',
		     settled_at = CURRENT_TIMESTAMP,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id = $2`,
		userID, settlementID,
	)

	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to mark settlement as paid", err, map[string]interface{}{
			"settlement_id": settlementID,
			"user_id":       userID,
		})
		return fmt.Errorf("failed to mark settlement as paid: %w", err)
	}

	logger.LogInfo(ctx, "settlement_service", "Settlement marked as paid", map[string]interface{}{
		"settlement_id": settlementID,
		"user_id":       userID,
	})

	return nil
}

func (s *SettlementService) SubmitDispute(settlementID, merchantID int, disputeType, reason string, originalAmount, disputedAmount float64) (*DisputeData, error) {
	ctx := context.Background()

	var dbMerchantID int
	err := s.db.QueryRow(
		"SELECT merchant_id FROM merchant_settlements WHERE id = $1",
		settlementID,
	).Scan(&dbMerchantID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("settlement not found")
		}
		return nil, fmt.Errorf("failed to query settlement: %w", err)
	}

	if dbMerchantID != merchantID {
		return nil, fmt.Errorf("not authorized to submit dispute for this settlement")
	}

	var disputeID int
	err = s.db.QueryRow(
		`INSERT INTO settlement_disputes 
		 (settlement_id, merchant_id, dispute_type, dispute_reason, original_amount, disputed_amount, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		settlementID, merchantID, disputeType, reason, originalAmount, disputedAmount, "pending",
	).Scan(&disputeID)

	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to submit dispute", err, map[string]interface{}{
			"settlement_id": settlementID,
			"merchant_id":   merchantID,
		})
		return nil, fmt.Errorf("failed to submit dispute: %w", err)
	}

	dispute := &DisputeData{
		ID:             disputeID,
		SettlementID:   settlementID,
		MerchantID:     merchantID,
		DisputeType:    disputeType,
		DisputeReason:  reason,
		OriginalAmount: originalAmount,
		DisputedAmount: disputedAmount,
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	logger.LogInfo(ctx, "settlement_service", "Dispute submitted", map[string]interface{}{
		"dispute_id":    disputeID,
		"settlement_id": settlementID,
		"merchant_id":   merchantID,
		"dispute_type":  disputeType,
	})

	return dispute, nil
}

func (s *SettlementService) ProcessDispute(disputeID, handlerID int, resolution string, adjustedAmount float64) error {
	ctx := context.Background()

	var dispute DisputeData
	var evidenceURLs []byte
	err := s.db.QueryRow(
		"SELECT id, settlement_id, merchant_id, dispute_type, dispute_reason, evidence_urls, original_amount, disputed_amount, status FROM settlement_disputes WHERE id = $1",
		disputeID,
	).Scan(&dispute.ID, &dispute.SettlementID, &dispute.MerchantID, &dispute.DisputeType, &dispute.DisputeReason, &evidenceURLs, &dispute.OriginalAmount, &dispute.DisputedAmount, &dispute.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("dispute not found")
		}
		return fmt.Errorf("failed to query dispute: %w", err)
	}

	_, err = s.db.Exec(
		`UPDATE settlement_disputes 
		 SET status = 'resolved',
		     handled_by = $1,
		     handled_at = CURRENT_TIMESTAMP,
		     resolution_notes = $2,
		     adjusted_amount = $3,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id = $4`,
		handlerID, resolution, adjustedAmount, disputeID,
	)

	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to process dispute", err, map[string]interface{}{
			"dispute_id": disputeID,
			"handler_id": handlerID,
		})
		return fmt.Errorf("failed to process dispute: %w", err)
	}

	if adjustedAmount > 0 {
		err = s.adjustSettlementAmount(dispute.SettlementID, adjustedAmount)
		if err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to adjust settlement amount", err, map[string]interface{}{
				"settlement_id":   dispute.SettlementID,
				"adjusted_amount": adjustedAmount,
			})
		}
	}

	logger.LogInfo(ctx, "settlement_service", "Dispute processed", map[string]interface{}{
		"dispute_id":      disputeID,
		"handler_id":      handlerID,
		"adjusted_amount": adjustedAmount,
	})

	return nil
}

func (s *SettlementService) adjustSettlementAmount(settlementID int, adjustedAmount float64) error {
	_, err := s.db.Exec(
		`UPDATE merchant_settlements 
		 SET settlement_amount = $1,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id = $2`,
		adjustedAmount, settlementID,
	)
	return err
}

func (s *SettlementService) ReconcileOrders(settlementID int) (*ReconciliationData, error) {
	ctx := context.Background()

	var merchantID int
	var periodStart, periodEnd time.Time
	err := s.db.QueryRow(
		"SELECT merchant_id, period_start, period_end FROM merchant_settlements WHERE id = $1",
		settlementID,
	).Scan(&merchantID, &periodStart, &periodEnd)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("settlement not found")
		}
		return nil, fmt.Errorf("failed to query settlement: %w", err)
	}

	var expectedCount int
	var expectedAmount float64
	err = s.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(total_price), 0)
		 FROM orders 
		 WHERE user_id = (SELECT user_id FROM merchants WHERE id = $1)
		   AND status = 'completed'
		   AND created_at >= $2 
		   AND created_at <= $3`,
		merchantID, periodStart, periodEnd,
	).Scan(&expectedCount, &expectedAmount)

	if err != nil {
		return nil, fmt.Errorf("failed to query expected orders: %w", err)
	}

	var actualCount int
	var actualAmount float64
	err = s.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(total_price), 0)
		 FROM order_items oi
		 JOIN orders o ON oi.order_id = o.id
		 WHERE o.user_id = (SELECT user_id FROM merchants WHERE id = $1)
		   AND o.status = 'completed'
		   AND o.created_at >= $2 
		   AND o.created_at <= $3`,
		merchantID, periodStart, periodEnd,
	).Scan(&actualCount, &actualAmount)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query actual orders: %w", err)
	}

	orderCountDiff := expectedCount - actualCount
	amountDiff := expectedAmount - actualAmount
	hasAnomalies := orderCountDiff != 0 || amountDiff != 0

	var anomalyDetails map[string]interface{}
	if hasAnomalies {
		anomalyDetails = map[string]interface{}{
			"order_count_diff": orderCountDiff,
			"amount_diff":      amountDiff,
			"expected_count":   expectedCount,
			"actual_count":     actualCount,
			"expected_amount":  expectedAmount,
			"actual_amount":    actualAmount,
		}
	}

	var reconciliationID int
	var anomalyJSON []byte
	if anomalyDetails != nil {
		anomalyJSON, err = json.Marshal(anomalyDetails)
		if err != nil {
			anomalyJSON = nil
		}
	}

	err = s.db.QueryRow(
		`INSERT INTO settlement_reconciliations 
		 (settlement_id, order_count_expected, order_count_actual, order_count_diff,
		  amount_expected, amount_actual, amount_diff, has_anomalies, anomaly_details)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		settlementID, expectedCount, actualCount, orderCountDiff,
		expectedAmount, actualAmount, amountDiff, hasAnomalies, anomalyJSON,
	).Scan(&reconciliationID)

	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to create reconciliation", err, map[string]interface{}{
			"settlement_id": settlementID,
		})
		return nil, fmt.Errorf("failed to create reconciliation: %w", err)
	}

	reconciliation := &ReconciliationData{
		ID:                 reconciliationID,
		SettlementID:       settlementID,
		OrderCountExpected: expectedCount,
		OrderCountActual:   actualCount,
		OrderCountDiff:     orderCountDiff,
		AmountExpected:     expectedAmount,
		AmountActual:       actualAmount,
		AmountDiff:         amountDiff,
		HasAnomalies:       hasAnomalies,
		AnomalyDetails:     anomalyDetails,
		ReconciledAt:       time.Now(),
	}

	logger.LogInfo(ctx, "settlement_service", "Reconciliation completed", map[string]interface{}{
		"settlement_id":    settlementID,
		"has_anomalies":    hasAnomalies,
		"order_count_diff": orderCountDiff,
		"amount_diff":      amountDiff,
	})

	return reconciliation, nil
}
