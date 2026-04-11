package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/models"
)

// SettlementService derives merchant-side figures from api_usage_logs (per-request usage and recorded cost).
// It does not use user order SKU token_amount or tokens.balance; those are retail fulfillment / user ledger fields.
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

	logger.LogInfo(ctx, "settlement_service", "Starting monthly settlement generation (billing-based)", map[string]interface{}{
		"period_start": periodStart.Format("2006-01-02"),
		"period_end":   periodEnd.Format("2006-01-02"),
	})

	query := `
		SELECT 
			mak.merchant_id,
			COUNT(aul.id) as total_requests,
			SUM(aul.input_tokens + aul.output_tokens) as total_tokens,
			SUM(aul.cost) as total_cost
		FROM api_usage_logs aul
		JOIN merchant_api_keys mak ON mak.id = aul.key_id
		JOIN merchants m ON m.id = mak.merchant_id
		WHERE aul.created_at >= $1 
			AND aul.created_at <= $2
			AND aul.status_code = 200
			AND m.status = 'approved'
		GROUP BY mak.merchant_id
		HAVING COUNT(aul.id) > 0
	`

	rows, err := s.db.Query(query, periodStart, periodEnd)
	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to query merchant billing data", err, nil)
		return nil, apperrors.NewAppError(
			"SETTLEMENT_QUERY_FAILED",
			"Failed to query merchant billing data",
			http.StatusInternalServerError,
			err,
		)
	}
	defer rows.Close()

	var settlements []models.MerchantSettlement
	platformFeeRate := 0.05

	for rows.Next() {
		var data SettlementData
		var totalTokens sql.NullInt64
		err := rows.Scan(&data.MerchantID, &data.TotalOrders, &totalTokens, &data.TotalSales)
		if err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to scan merchant data", err, nil)
			continue
		}

		settlement, err := s.createSettlementWithItems(data.MerchantID, periodStart, periodEnd, data.TotalSales, platformFeeRate)
		if err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to create settlement", err, map[string]interface{}{
				"merchant_id": data.MerchantID,
			})
			continue
		}

		settlements = append(settlements, *settlement)

		logger.LogInfo(ctx, "settlement_service", "Settlement created", map[string]interface{}{
			"settlement_id":     settlement.ID,
			"merchant_id":       data.MerchantID,
			"total_sales":       data.TotalSales,
			"settlement_amount": settlement.SettlementAmount,
			"total_requests":    data.TotalOrders,
			"total_tokens":      totalTokens,
		})
	}

	logger.LogInfo(ctx, "settlement_service", "Monthly settlement generation completed", map[string]interface{}{
		"total_settlements": len(settlements),
	})

	return settlements, nil
}

func (s *SettlementService) GenerateSettlementForMerchant(merchantID int, periodStart, periodEnd time.Time) (*models.MerchantSettlement, error) {
	ctx := context.Background()

	logger.LogInfo(ctx, "settlement_service", "Generating settlement for merchant", map[string]interface{}{
		"merchant_id":  merchantID,
		"period_start": periodStart.Format("2006-01-02"),
		"period_end":   periodEnd.Format("2006-01-02"),
	})

	if s.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if merchantID <= 0 {
		return nil, fmt.Errorf("invalid merchant ID: %d", merchantID)
	}

	if periodStart.After(periodEnd) {
		return nil, fmt.Errorf("invalid date range: period_start (%s) is after period_end (%s)", periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))
	}

	query := `
		SELECT 
			COUNT(aul.id) as total_requests,
			SUM(aul.input_tokens + aul.output_tokens) as total_tokens,
			SUM(aul.cost) as total_cost
		FROM api_usage_logs aul
		JOIN merchant_api_keys mak ON mak.id = aul.key_id
		WHERE mak.merchant_id = $1
			AND aul.created_at >= $2 
			AND aul.created_at <= $3
			AND aul.status_code = 200
	`

	var totalRequests int
	var totalTokens sql.NullInt64
	var totalCost sql.NullFloat64
	err := s.db.QueryRow(query, merchantID, periodStart, periodEnd).Scan(&totalRequests, &totalTokens, &totalCost)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no billing data found for merchant %d in the specified period", merchantID)
		}
		return nil, fmt.Errorf("failed to query billing data: %w", err)
	}

	if totalRequests == 0 {
		return nil, fmt.Errorf("no billing data found for merchant %d in the specified period", merchantID)
	}

	actualTotalTokens := int64(0)
	if totalTokens.Valid {
		actualTotalTokens = totalTokens.Int64
	}

	actualTotalCost := 0.0
	if totalCost.Valid {
		actualTotalCost = totalCost.Float64
	}

	platformFeeRate := 0.05
	settlement, err := s.createSettlementWithItems(merchantID, periodStart, periodEnd, actualTotalCost, platformFeeRate)
	if err != nil {
		return nil, err
	}

	logger.LogInfo(ctx, "settlement_service", "Settlement created for merchant", map[string]interface{}{
		"settlement_id":     settlement.ID,
		"merchant_id":       merchantID,
		"total_sales":       actualTotalCost,
		"settlement_amount": settlement.SettlementAmount,
		"total_requests":    totalRequests,
		"total_tokens":      actualTotalTokens,
	})

	return settlement, nil
}

func (s *SettlementService) createSettlementWithItems(merchantID int, periodStart, periodEnd time.Time, totalSales float64, platformFeeRate float64) (*models.MerchantSettlement, error) {
	ctx := context.Background()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existingID int
	err = tx.QueryRow(
		"SELECT id FROM merchant_settlements WHERE merchant_id = $1 AND period_start = $2 AND period_end = $3 FOR UPDATE",
		merchantID, periodStart, periodEnd,
	).Scan(&existingID)

	if err == nil {
		return nil, fmt.Errorf("settlement already exists: %d", existingID)
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing settlement: %w", err)
	}

	platformFee := totalSales * platformFeeRate
	settlementAmount := totalSales - platformFee

	var settlementID int
	err = tx.QueryRow(
		`INSERT INTO merchant_settlements 
		 (merchant_id, period_start, period_end, total_sales, platform_fee, settlement_amount, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		merchantID, periodStart, periodEnd, totalSales, platformFee, settlementAmount, "pending",
	).Scan(&settlementID)

	if err != nil {
		return nil, fmt.Errorf("failed to create settlement: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO settlement_items (settlement_id, api_usage_log_id, user_id, merchant_id, provider, model, input_tokens, output_tokens, cost)
		 SELECT $1, aul.id, aul.user_id, mak.merchant_id, aul.provider, aul.model, aul.input_tokens, aul.output_tokens, aul.cost
		 FROM api_usage_logs aul
		 JOIN merchant_api_keys mak ON mak.id = aul.key_id
		 WHERE mak.merchant_id = $2
		   AND aul.created_at >= $3 
		   AND aul.created_at <= $4
		   AND aul.status_code = 200`,
		settlementID, merchantID, periodStart, periodEnd,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create settlement items: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	settlement := &models.MerchantSettlement{
		ID:               settlementID,
		MerchantID:       merchantID,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		TotalSales:       totalSales,
		PlatformFee:      platformFee,
		SettlementAmount: settlementAmount,
		Status:           "pending",
	}

	logger.LogInfo(ctx, "settlement_service", "Settlement with items created", map[string]interface{}{
		"settlement_id": settlementID,
		"merchant_id":   merchantID,
	})

	return settlement, nil
}

type BillingRecord struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	MerchantID   int       `json:"merchant_id"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Cost         float64   `json:"cost"`
	RequestID    string    `json:"request_id"`
	StatusCode   int       `json:"status_code"`
	LatencyMs    int       `json:"latency_ms"`
	CreatedAt    time.Time `json:"created_at"`
}

func (s *SettlementService) GetBillingRecords(merchantID int, startDate, endDate time.Time, page, pageSize int) ([]BillingRecord, error) {
	ctx := context.Background()

	offset := (page - 1) * pageSize

	query := `
		SELECT 
			aul.id,
			aul.user_id,
			mak.merchant_id,
			aul.provider,
			aul.model,
			aul.input_tokens,
			aul.output_tokens,
			aul.cost,
			aul.request_id,
			aul.status_code,
			aul.latency_ms,
			aul.created_at
		FROM api_usage_logs aul
		JOIN merchant_api_keys mak ON mak.id = aul.key_id
		WHERE aul.created_at >= $1 
			AND aul.created_at <= $2
	`
	args := []interface{}{startDate, endDate}
	argIndex := 3

	if merchantID > 0 {
		query += " AND mak.merchant_id = $" + strconv.Itoa(argIndex)
		args = append(args, merchantID)
		argIndex++
	}

	query += " ORDER BY aul.created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to query billing records", err, nil)
		return nil, fmt.Errorf("failed to query billing records: %w", err)
	}
	defer rows.Close()

	var records []BillingRecord
	for rows.Next() {
		var r BillingRecord
		err := rows.Scan(
			&r.ID, &r.UserID, &r.MerchantID, &r.Provider, &r.Model,
			&r.InputTokens, &r.OutputTokens, &r.Cost, &r.RequestID,
			&r.StatusCode, &r.LatencyMs, &r.CreatedAt,
		)
		if err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to scan billing record", err, nil)
			continue
		}
		records = append(records, r)
	}

	return records, nil
}

type SettlementItem struct {
	ID           int       `json:"id"`
	SettlementID int       `json:"settlement_id"`
	APILogID     int       `json:"api_usage_log_id"`
	UserID       int       `json:"user_id"`
	MerchantID   int       `json:"merchant_id"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Cost         float64   `json:"cost"`
	CreatedAt    time.Time `json:"created_at"`
}

func (s *SettlementService) GetSettlementItems(settlementID int) ([]SettlementItem, error) {
	ctx := context.Background()

	query := `
		SELECT 
			id,
			settlement_id,
			api_usage_log_id,
			user_id,
			merchant_id,
			provider,
			model,
			input_tokens,
			output_tokens,
			cost,
			created_at
		FROM settlement_items
		WHERE settlement_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, settlementID)
	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to query settlement items", err, map[string]interface{}{
			"settlement_id": settlementID,
		})
		return nil, fmt.Errorf("failed to query settlement items: %w", err)
	}
	defer rows.Close()

	var items []SettlementItem
	for rows.Next() {
		var item SettlementItem
		err := rows.Scan(
			&item.ID, &item.SettlementID, &item.APILogID, &item.UserID,
			&item.MerchantID, &item.Provider, &item.Model,
			&item.InputTokens, &item.OutputTokens, &item.Cost, &item.CreatedAt,
		)
		if err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to scan settlement item", err, nil)
			continue
		}
		items = append(items, item)
	}

	return items, nil
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

	tx, err := s.db.Begin()
	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to begin transaction", err, nil)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var merchantConfirmed, financeApproved bool
	err = tx.QueryRow(
		"SELECT merchant_confirmed, finance_approved FROM merchant_settlements WHERE id = $1 FOR UPDATE",
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

	_, err = tx.Exec(
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

	if err := tx.Commit(); err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to commit transaction", err, nil)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.LogInfo(ctx, "settlement_service", "Settlement approved by finance", map[string]interface{}{
		"settlement_id":   settlementID,
		"finance_user_id": financeUserID,
	})

	return nil
}

func (s *SettlementService) MarkAsPaid(settlementID, userID int) error {
	ctx := context.Background()

	tx, err := s.db.Begin()
	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to begin transaction", err, nil)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var financeApproved bool
	err = tx.QueryRow(
		"SELECT finance_approved FROM merchant_settlements WHERE id = $1 FOR UPDATE",
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

	_, err = tx.Exec(
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

	if err := tx.Commit(); err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to commit transaction", err, nil)
		return fmt.Errorf("failed to commit transaction: %w", err)
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

	tx, err := s.db.Begin()
	if err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to begin transaction", err, nil)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var dispute DisputeData
	var evidenceURLs []byte
	err = tx.QueryRow(
		"SELECT id, settlement_id, merchant_id, dispute_type, dispute_reason, evidence_urls, original_amount, disputed_amount, status FROM settlement_disputes WHERE id = $1 FOR UPDATE",
		disputeID,
	).Scan(&dispute.ID, &dispute.SettlementID, &dispute.MerchantID, &dispute.DisputeType, &dispute.DisputeReason, &evidenceURLs, &dispute.OriginalAmount, &dispute.DisputedAmount, &dispute.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("dispute not found")
		}
		return fmt.Errorf("failed to query dispute: %w", err)
	}

	_, err = tx.Exec(
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
		_, err = tx.Exec(
			`UPDATE merchant_settlements 
			 SET settlement_amount = $1,
			     updated_at = CURRENT_TIMESTAMP
			 WHERE id = $2`,
			adjustedAmount, dispute.SettlementID,
		)
		if err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to adjust settlement amount", err, map[string]interface{}{
				"settlement_id":   dispute.SettlementID,
				"adjusted_amount": adjustedAmount,
			})
			return fmt.Errorf("failed to adjust settlement amount: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		logger.LogError(ctx, "settlement_service", "Failed to commit transaction", err, nil)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.LogInfo(ctx, "settlement_service", "Dispute processed", map[string]interface{}{
		"dispute_id":      disputeID,
		"handler_id":      handlerID,
		"adjusted_amount": adjustedAmount,
	})

	return nil
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

		if alertErr := s.sendReconciliationAlert(settlementID, merchantID, anomalyDetails); alertErr != nil {
			logger.LogError(ctx, "settlement_service", "Failed to send reconciliation alert", alertErr, map[string]interface{}{
				"settlement_id": settlementID,
				"merchant_id":   merchantID,
			})
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

	if hasAnomalies {
		if err := s.markSettlementAnomaly(settlementID, anomalyDetails); err != nil {
			logger.LogError(ctx, "settlement_service", "Failed to mark settlement anomaly", err, map[string]interface{}{
				"settlement_id": settlementID,
			})
		}
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

func (s *SettlementService) sendReconciliationAlert(settlementID, merchantID int, anomalyDetails map[string]interface{}) error {
	ctx := context.Background()

	logger.LogInfo(ctx, "settlement_service", "Sending reconciliation alert", map[string]interface{}{
		"settlement_id":   settlementID,
		"merchant_id":     merchantID,
		"anomaly_details": anomalyDetails,
	})

	return nil
}

func (s *SettlementService) markSettlementAnomaly(settlementID int, anomalyDetails map[string]interface{}) error {
	ctx := context.Background()

	_, err := s.db.Exec(
		`UPDATE merchant_settlements 
		 SET status = 'anomaly_detected',
		     updated_at = NOW()
		 WHERE id = $1`,
		settlementID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark settlement anomaly: %w", err)
	}

	logger.LogInfo(ctx, "settlement_service", "Settlement anomaly marked", map[string]interface{}{
		"settlement_id":   settlementID,
		"anomaly_details": anomalyDetails,
	})

	return nil
}
