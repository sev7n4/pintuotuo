package payment

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/token"
)

// Service defines the payment service interface
type Service interface {
	// Payment operations
	InitiatePayment(ctx context.Context, userID int, req *InitiatePaymentRequest) (*Payment, error)
	GetPaymentByID(ctx context.Context, userID int, paymentID int) (*Payment, error)
	GetPaymentsByOrder(ctx context.Context, userID int, orderID int) ([]Payment, error)
	ListPayments(ctx context.Context, userID int, params *ListPaymentsParams) (*PaymentListResult, error)

	// Webhook handling
	HandleAlipayCallback(ctx context.Context, payload *AlipayCallback) (*Payment, error)
	HandleWechatCallback(ctx context.Context, payload *WechatCallback) (*Payment, error)

	// Refunds
	RefundPayment(ctx context.Context, userID int, paymentID int, reason string) (*Payment, error)

	// Revenue tracking
	GetMerchantRevenue(ctx context.Context, merchantID int, period string) (*MerchantRevenue, error)
	CalculateCommission(amount float64, commissionRate float64) float64
}

// service implements the Service interface
type service struct {
	db           *sql.DB
	log          *log.Logger
	orderService order.Service
	tokenService token.Service
}

// NewService creates a new payment service
func NewService(db *sql.DB, orderService order.Service, logger *log.Logger, tokenService token.Service) Service {
	if logger == nil {
		logger = log.New(os.Stderr, "[PaymentService] ", log.LstdFlags)
	}

	if tokenService == nil {
		tokenService = token.NewService(db, logger)
	}

	return &service{
		db:           db,
		log:          logger,
		orderService: orderService,
		tokenService: tokenService,
	}
}

// InitiatePayment initiates a payment for an order
func (s *service) InitiatePayment(ctx context.Context, userID int, req *InitiatePaymentRequest) (*Payment, error) {
	// Validate payment method
	if req.PaymentMethod != "alipay" && req.PaymentMethod != "wechat" {
		return nil, ErrInvalidPaymentMethod
	}

	// Get order details
	order, err := s.orderService.GetOrderByID(ctx, userID, req.OrderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	// Validate order status
	if order.Status != "pending" {
		if order.Status == "paid" || order.Status == "completed" {
			return nil, ErrOrderAlreadyPaid
		}
		return nil, ErrOrderCancelled
	}

	// Check if there's already a pending payment for this order
	var existingPaymentStatus string
	err = s.db.QueryRowContext(
		ctx,
		"SELECT status FROM payments WHERE order_id = $1 AND status = 'pending' LIMIT 1",
		req.OrderID,
	).Scan(&existingPaymentStatus)

	if err == nil {
		// Found an existing pending payment
		return nil, ErrPaymentAlreadyPending
	} else if err != sql.ErrNoRows {
		s.log.Printf("Failed to check existing payments: %v", err)
		return nil, wrapError("InitiatePayment", "check_existing", err)
	}
	// sql.ErrNoRows is expected, meaning no pending payment exists

	// Create payment record
	var payment Payment
	err = s.db.QueryRowContext(
		ctx,
		"INSERT INTO payments (user_id, order_id, amount, method, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id, user_id, order_id, amount, method, status, created_at, updated_at",
		userID, req.OrderID, order.TotalPrice, req.PaymentMethod,
		"pending",
	).Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		s.log.Printf("Failed to create payment: %v", err)
		return nil, wrapError("InitiatePayment", "insert", err)
	}

	// Invalidate cache for payments
	_ = cache.InvalidatePatterns(ctx, "payments:user:"+strconv.Itoa(userID)+":*")

	s.log.Printf("Payment initiated: id=%d, user_id=%d, order_id=%d, amount=%.2f, method=%s", payment.ID, userID, req.OrderID, order.TotalPrice, req.PaymentMethod)
	return &payment, nil
}

// GetPaymentByID retrieves a payment by ID
func (s *service) GetPaymentByID(ctx context.Context, userID int, paymentID int) (*Payment, error) {
	// Try cache first
	cacheKey := cache.PaymentKey(paymentID)
	if cachedPayment, err := cache.Get(ctx, cacheKey); err == nil {
		var payment Payment
		if err := cache.UnmarshalCachedValue(cachedPayment, &payment); err == nil {
			// Verify ownership
			if payment.UserID != userID {
				return nil, ErrPaymentNotFound
			}
			return &payment, nil
		}
	}

	var payment Payment
	var transactionID sql.NullString
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at FROM payments WHERE id = $1 AND user_id = $2",
		paymentID, userID,
	).Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &transactionID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		s.log.Printf("Failed to get payment: %v", err)
		return nil, wrapError("GetPaymentByID", "select", err)
	}

	if transactionID.Valid {
		tempID := transactionID.String
		payment.TransactionID = &tempID
	}

	// Cache the result
	_ = cache.Set(ctx, cacheKey, cache.MarshalCachedValue(payment), cache.PaymentTTL)

	return &payment, nil
}

// GetPaymentsByOrder retrieves all payments for an order
func (s *service) GetPaymentsByOrder(ctx context.Context, userID int, orderID int) ([]Payment, error) {
	rows, err := s.db.QueryContext(
		ctx,
		"SELECT id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at FROM payments WHERE order_id = $1 AND user_id = $2 ORDER BY created_at DESC",
		orderID, userID,
	)
	if err != nil {
		s.log.Printf("Failed to get payments by order: %v", err)
		return nil, wrapError("GetPaymentsByOrder", "select", err)
	}
	defer rows.Close()

	payments := []Payment{}
	for rows.Next() {
		var payment Payment
		var transactionID sql.NullString
		err := rows.Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &transactionID, &payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			s.log.Printf("Failed to scan payment: %v", err)
			return nil, wrapError("GetPaymentsByOrder", "scan", err)
		}
		if transactionID.Valid {
			tempID := transactionID.String
			payment.TransactionID = &tempID
		}
		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		s.log.Printf("Row iteration error: %v", err)
		return nil, wrapError("GetPaymentsByOrder", "iteration", err)
	}

	return payments, nil
}

// ListPayments retrieves payments for a user with pagination
func (s *service) ListPayments(ctx context.Context, userID int, params *ListPaymentsParams) (*PaymentListResult, error) {
	// Validate parameters
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	offset := (params.Page - 1) * params.PerPage

	// Build query
	whereClause := "user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if params.Status != "" {
		whereClause += " AND status = $" + strconv.Itoa(argIndex)
		args = append(args, params.Status)
		argIndex++
	}

	if params.Method != "" {
		whereClause += " AND method = $" + strconv.Itoa(argIndex)
		args = append(args, params.Method)
		argIndex++
	}

	// Get total count
	var total int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM payments WHERE "+whereClause, args...).Scan(&total)
	if err != nil {
		s.log.Printf("Failed to count payments: %v", err)
		return nil, wrapError("ListPayments", "count", err)
	}

	// Add LIMIT and OFFSET for paginated data
	limit := params.PerPage
	query := "SELECT id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at FROM payments WHERE " + whereClause + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		s.log.Printf("Failed to list payments: %v", err)
		return nil, wrapError("ListPayments", "select", err)
	}
	defer rows.Close()

	payments := make([]Payment, 0, params.PerPage)
	for rows.Next() {
		var payment Payment
		var transactionID sql.NullString
		err := rows.Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &transactionID, &payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			s.log.Printf("Failed to scan payment: %v", err)
			return nil, wrapError("ListPayments", "scan", err)
		}
		if transactionID.Valid {
			tempID := transactionID.String
			payment.TransactionID = &tempID
		}
		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		s.log.Printf("Row iteration error: %v", err)
		return nil, wrapError("ListPayments", "iteration", err)
	}

	return &PaymentListResult{
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
		Data:    payments,
	}, nil
}

// HandleAlipayCallback handles Alipay webhook callback
func (s *service) HandleAlipayCallback(ctx context.Context, payload *AlipayCallback) (*Payment, error) {
	// Verify signature (in production, validate with Alipay's public key)
	// For now, we skip signature verification in this implementation

	// Extract payment ID from out_trade_no
	paymentID, err := strconv.Atoi(payload.OutTradeNo)
	if err != nil {
		s.log.Printf("Failed to parse payment ID from out_trade_no: %v", err)
		return nil, ErrInvalidSignature
	}

	// Get current payment
	var currentStatus string
	var userID int
	err = s.db.QueryRowContext(
		ctx,
		"SELECT status, user_id FROM payments WHERE id = $1",
		paymentID,
	).Scan(&currentStatus, &userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		s.log.Printf("Failed to get payment status: %v", err)
		return nil, wrapError("HandleAlipayCallback", "select", err)
	}

	// Check for idempotency - don't process if already processed
	if currentStatus == "success" || currentStatus == "failed" {
		// Return existing payment
		var payment Payment
		var transactionID sql.NullString
		err = s.db.QueryRowContext(
			ctx,
			"SELECT id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at FROM payments WHERE id = $1",
			paymentID,
		).Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &transactionID, &payment.CreatedAt, &payment.UpdatedAt)

		if err != nil {
			return nil, wrapError("HandleAlipayCallback", "select_existing", err)
		}

		if transactionID.Valid {
			tempID := transactionID.String
			payment.TransactionID = &tempID
		}
		return &payment, nil
	}

	// Update payment status based on Alipay response
	newStatus := "failed"
	if payload.TradeStatus == "TRADE_SUCCESS" || payload.TradeStatus == "TRADE_FINISHED" {
		newStatus = "success"
	}

	var payment Payment
	var transactionID sql.NullString
	err = s.db.QueryRowContext(
		ctx,
		"UPDATE payments SET status = $1, transaction_id = $2, updated_at = NOW() WHERE id = $3 RETURNING id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at",
		newStatus, payload.TradeNo, paymentID,
	).Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &transactionID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		s.log.Printf("Failed to update payment: %v", err)
		return nil, wrapError("HandleAlipayCallback", "update", err)
	}

	if transactionID.Valid {
		tempID := transactionID.String
		payment.TransactionID = &tempID
	}

	// If payment successful, update order status and recharge tokens
	if newStatus == "success" {
		// Get order ID to update
		_, err := s.orderService.UpdateOrderStatus(ctx, payment.UserID, payment.OrderID, "paid")
		if err != nil {
			s.log.Printf("Failed to update order status: %v", err)
			// Log but don't fail - payment was already recorded
		}

		// Recharge tokens for user based on payment amount
		if s.tokenService != nil {
			tokenReq := &token.RechargeTokensRequest{
				UserID: payment.UserID,
				Amount: payment.Amount,
				Reason: "Payment successful - Order " + strconv.Itoa(payment.OrderID),
			}

			_, err := s.tokenService.RechargeTokens(ctx, tokenReq)
			if err != nil {
				s.log.Printf("Failed to recharge tokens after payment: %v", err)
				// Log but don't fail - payment was already recorded
			}
		}
	}

	// Invalidate cache
	_ = cache.Delete(ctx, cache.PaymentKey(paymentID))
	_ = cache.InvalidatePatterns(ctx, "payments:user:"+strconv.Itoa(payment.UserID)+":*")

	s.log.Printf("Alipay callback processed: payment_id=%d, status=%s, transaction_id=%s", paymentID, newStatus, payload.TradeNo)
	return &payment, nil
}

// HandleWechatCallback handles WeChat Pay webhook callback
func (s *service) HandleWechatCallback(ctx context.Context, payload *WechatCallback) (*Payment, error) {
	// Verify signature (in production, validate with WeChat's key)
	// For now, we skip signature verification in this implementation

	// Extract payment ID from out_trade_no
	paymentID, err := strconv.Atoi(payload.OutTradeNo)
	if err != nil {
		s.log.Printf("Failed to parse payment ID from out_trade_no: %v", err)
		return nil, ErrInvalidSignature
	}

	// Get current payment
	var currentStatus string
	var userID int
	err = s.db.QueryRowContext(
		ctx,
		"SELECT status, user_id FROM payments WHERE id = $1",
		paymentID,
	).Scan(&currentStatus, &userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		s.log.Printf("Failed to get payment status: %v", err)
		return nil, wrapError("HandleWechatCallback", "select", err)
	}

	// Check for idempotency - don't process if already processed
	if currentStatus == "success" || currentStatus == "failed" {
		// Return existing payment
		var payment Payment
		var transactionID sql.NullString
		err = s.db.QueryRowContext(
			ctx,
			"SELECT id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at FROM payments WHERE id = $1",
			paymentID,
		).Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &transactionID, &payment.CreatedAt, &payment.UpdatedAt)

		if err != nil {
			return nil, wrapError("HandleWechatCallback", "select_existing", err)
		}

		if transactionID.Valid {
			tempID := transactionID.String
			payment.TransactionID = &tempID
		}
		return &payment, nil
	}

	// Update payment status based on WeChat response
	newStatus := "failed"
	if payload.ResultCode == "SUCCESS" {
		newStatus = "success"
	}

	var payment Payment
	var transactionID sql.NullString
	err = s.db.QueryRowContext(
		ctx,
		"UPDATE payments SET status = $1, transaction_id = $2, updated_at = NOW() WHERE id = $3 RETURNING id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at",
		newStatus, payload.TransactionID, paymentID,
	).Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &transactionID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		s.log.Printf("Failed to update payment: %v", err)
		return nil, wrapError("HandleWechatCallback", "update", err)
	}

	if transactionID.Valid {
		tempID := transactionID.String
		payment.TransactionID = &tempID
	}

	// If payment successful, update order status and recharge tokens
	if newStatus == "success" {
		_, err := s.orderService.UpdateOrderStatus(ctx, payment.UserID, payment.OrderID, "paid")
		if err != nil {
			s.log.Printf("Failed to update order status: %v", err)
			// Log but don't fail - payment was already recorded
		}

		// Recharge tokens for user based on payment amount
		if s.tokenService != nil {
			tokenReq := &token.RechargeTokensRequest{
				UserID: payment.UserID,
				Amount: payment.Amount,
				Reason: "Payment successful - Order " + strconv.Itoa(payment.OrderID),
			}

			_, err := s.tokenService.RechargeTokens(ctx, tokenReq)
			if err != nil {
				s.log.Printf("Failed to recharge tokens after payment: %v", err)
				// Log but don't fail - payment was already recorded
			}
		}
	}

	// Invalidate cache
	_ = cache.Delete(ctx, cache.PaymentKey(paymentID))
	_ = cache.InvalidatePatterns(ctx, "payments:user:"+strconv.Itoa(payment.UserID)+":*")

	s.log.Printf("WeChat callback processed: payment_id=%d, status=%s, transaction_id=%s", paymentID, newStatus, payload.TransactionID)
	return &payment, nil
}

// RefundPayment processes a refund for a payment
func (s *service) RefundPayment(ctx context.Context, userID int, paymentID int, reason string) (*Payment, error) {
	// Get payment details
	var currentStatus string
	err := s.db.QueryRowContext(
		ctx,
		"SELECT status FROM payments WHERE id = $1 AND user_id = $2",
		paymentID, userID,
	).Scan(&currentStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		s.log.Printf("Failed to get payment: %v", err)
		return nil, wrapError("RefundPayment", "select", err)
	}

	// Can only refund successful payments
	if currentStatus != "success" {
		return nil, ErrCannotRefundPendingPayment
	}

	// Update payment status to refunded
	var payment Payment
	var transactionID sql.NullString
	err = s.db.QueryRowContext(
		ctx,
		"UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at",
		"refunded", paymentID,
	).Scan(&payment.ID, &payment.UserID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &transactionID, &payment.CreatedAt, &payment.UpdatedAt)

	if err != nil {
		s.log.Printf("Failed to update payment: %v", err)
		return nil, wrapError("RefundPayment", "update", err)
	}

	if transactionID.Valid {
		tempID := transactionID.String
		payment.TransactionID = &tempID
	}

	// Invalidate cache
	_ = cache.Delete(ctx, cache.PaymentKey(paymentID))
	_ = cache.InvalidatePatterns(ctx, "payments:user:"+strconv.Itoa(userID)+":*")

	s.log.Printf("Refund processed: payment_id=%d, reason=%s", paymentID, reason)
	return &payment, nil
}

// GetMerchantRevenue retrieves merchant revenue information
func (s *service) GetMerchantRevenue(ctx context.Context, merchantID int, period string) (*MerchantRevenue, error) {
	// Query payments for merchant's products with commission calculation
	var totalSales float64
	var transactionCount int

	query := `
		SELECT
			COALESCE(SUM(p.amount), 0) as total_sales,
			COUNT(p.id) as transaction_count
		FROM payments p
		JOIN orders o ON p.order_id = o.id
		JOIN products prod ON o.product_id = prod.id
		WHERE prod.merchant_id = $1 AND p.status = 'success'
	`

	err := s.db.QueryRowContext(ctx, query, merchantID).Scan(&totalSales, &transactionCount)
	if err != nil && err != sql.ErrNoRows {
		s.log.Printf("Failed to get merchant revenue: %v", err)
		return nil, wrapError("GetMerchantRevenue", "select", err)
	}

	// Default commission rate
	commissionRate := 0.30 // 30% platform commission

	// Calculate revenue breakdown
	platformCommission := s.CalculateCommission(totalSales, commissionRate)
	merchantEarnings := totalSales - platformCommission

	var averageOrderValue float64
	if transactionCount > 0 {
		averageOrderValue = totalSales / float64(transactionCount)
	}

	return &MerchantRevenue{
		MerchantID:         merchantID,
		Period:             period,
		TotalSales:         totalSales,
		CommissionRate:     commissionRate,
		PlatformCommission: platformCommission,
		APICallCost:        0, // Would be populated from token transactions
		MerchantEarnings:   merchantEarnings,
		TransactionCount:   transactionCount,
		AverageOrderValue:  averageOrderValue,
	}, nil
}

// CalculateCommission calculates the commission for a payment
func (s *service) CalculateCommission(amount float64, commissionRate float64) float64 {
	return amount * commissionRate
}
