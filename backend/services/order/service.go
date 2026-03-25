package order

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/pintuotuo/backend/cache"
)

// Service defines the order service interface
type Service interface {
	// Read operations
	ListOrders(ctx context.Context, userID int, params *ListOrdersParams) (*ListOrdersResult, error)
	GetOrderByID(ctx context.Context, userID int, orderID int) (*Order, error)
	GetOrdersByStatus(ctx context.Context, userID int, status string) ([]Order, error)

	// Write operations
	CreateOrder(ctx context.Context, userID int, req *CreateOrderRequest) (*Order, error)
	CancelOrder(ctx context.Context, userID int, orderID int) (*Order, error)
	UpdateOrderStatus(ctx context.Context, userID int, orderID int, newStatus string) (*Order, error)
}

// service implements the Service interface
type service struct {
	db  *sql.DB
	log *log.Logger
}

// NewService creates a new order service
func NewService(db *sql.DB, logger *log.Logger) Service {
	if logger == nil {
		logger = log.New(os.Stderr, "[OrderService] ", log.LstdFlags)
	}

	return &service{
		db:  db,
		log: logger,
	}
}

// CreateOrder creates a new order
func (s *service) CreateOrder(ctx context.Context, userID int, req *CreateOrderRequest) (*Order, error) {
	// Validate input
	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// Get product info
	var productPrice float64
	var productStock int

	err := s.db.QueryRowContext(
		ctx,
		"SELECT price, stock FROM products WHERE id = $1",
		req.ProductID,
	).Scan(&productPrice, &productStock)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, wrapError("CreateOrder", "getProduct", err)
	}

	// Check stock
	if productStock < req.Quantity {
		return nil, ErrInsufficientStock
	}

	// Calculate total price
	totalPrice := productPrice * float64(req.Quantity)

	// Create order
	var order Order
	err = s.db.QueryRowContext(
		ctx,
		"INSERT INTO orders (user_id, product_id, group_id, quantity, unit_price, total_price, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at",
		userID, req.ProductID, req.GroupID, req.Quantity, productPrice, totalPrice, "pending",
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, wrapError("CreateOrder", "insert", err)
	}

	// Invalidate user's order list cache
	_ = cache.InvalidatePatterns(ctx, "orders:user:"+string(rune(userID))+":*")

	s.log.Printf("Order created: id=%d, user_id=%d, product_id=%d, quantity=%d, total_price=%.2f", order.ID, userID, req.ProductID, req.Quantity, totalPrice)
	return &order, nil
}

// ListOrders retrieves orders for a user with pagination
func (s *service) ListOrders(ctx context.Context, userID int, params *ListOrdersParams) (*ListOrdersResult, error) {
	// Validate parameters
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	// Query database
	offset := (params.Page - 1) * params.PerPage

	var rows *sql.Rows
	var err error

	if params.Status == "" || params.Status == "all" {
		rows, err = s.db.QueryContext(
			ctx,
			"SELECT id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
			userID, params.PerPage, offset,
		)
	} else {
		rows, err = s.db.QueryContext(
			ctx,
			"SELECT id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at FROM orders WHERE user_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4",
			userID, params.Status, params.PerPage, offset,
		)
	}

	if err != nil {
		return nil, wrapError("ListOrders", "query", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.GroupID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, wrapError("ListOrders", "scan", err)
		}
		orders = append(orders, o)
	}

	// Get total count
	var total int
	if params.Status == "" || params.Status == "all" {
		s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&total)
	} else {
		s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status = $2", userID, params.Status).Scan(&total)
	}

	result := &ListOrdersResult{
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
		Data:    orders,
	}

	s.log.Printf("Listed orders: user_id=%d, page=%d, per_page=%d, status=%s, total=%d", userID, params.Page, params.PerPage, params.Status, total)
	return result, nil
}

// GetOrderByID retrieves a single order by ID (ownership verified)
func (s *service) GetOrderByID(ctx context.Context, userID int, orderID int) (*Order, error) {
	var order Order
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at FROM orders WHERE id = $1 AND user_id = $2",
		orderID, userID,
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, wrapError("GetOrderByID", "query", err)
	}

	return &order, nil
}

// GetOrdersByStatus retrieves all orders with a specific status for a user
func (s *service) GetOrdersByStatus(ctx context.Context, userID int, status string) ([]Order, error) {
	if status == "" {
		return nil, ErrInvalidStatus
	}

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at FROM orders WHERE user_id = $1 AND status = $2 ORDER BY created_at DESC",
		userID, status,
	)

	if err != nil {
		return nil, wrapError("GetOrdersByStatus", "query", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.GroupID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, wrapError("GetOrdersByStatus", "scan", err)
		}
		orders = append(orders, o)
	}

	s.log.Printf("Retrieved orders by status: user_id=%d, status=%s, count=%d", userID, status, len(orders))
	return orders, nil
}

// CancelOrder cancels a pending order
func (s *service) CancelOrder(ctx context.Context, userID int, orderID int) (*Order, error) {
	// Check order exists and belongs to user
	var currentStatus string
	err := s.db.QueryRowContext(
		ctx,
		"SELECT status FROM orders WHERE id = $1 AND user_id = $2",
		orderID, userID,
	).Scan(&currentStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, wrapError("CancelOrder", "checkOrder", err)
	}

	// Can only cancel pending orders
	if currentStatus != "pending" {
		return nil, ErrCannotCancelOrder
	}

	// Update order status
	var order Order
	err = s.db.QueryRowContext(
		ctx,
		"UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3 RETURNING id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at",
		"cancelled", orderID, userID,
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, wrapError("CancelOrder", "update", err)
	}

	// Invalidate cache
	_ = cache.InvalidatePatterns(ctx, "orders:user:"+string(rune(userID))+":*")

	s.log.Printf("Order cancelled: id=%d, user_id=%d", orderID, userID)
	return &order, nil
}

// UpdateOrderStatus updates order status (admin/internal use)
func (s *service) UpdateOrderStatus(ctx context.Context, userID int, orderID int, newStatus string) (*Order, error) {
	// Validate new status
	validStatuses := map[string]bool{
		"pending":   true,
		"paid":      true,
		"completed": true,
		"cancelled": true,
		"failed":    true,
	}

	if !validStatuses[newStatus] {
		return nil, ErrInvalidStatus
	}

	// Update order status
	var order Order
	err := s.db.QueryRowContext(
		ctx,
		"UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2 AND user_id = $3 RETURNING id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at",
		newStatus, orderID, userID,
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.GroupID, &order.Quantity, &order.UnitPrice, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrderNotFound
		}
		return nil, wrapError("UpdateOrderStatus", "update", err)
	}

	// Invalidate cache
	_ = cache.InvalidatePatterns(ctx, "orders:user:"+string(rune(userID))+":*")

	s.log.Printf("Order status updated: id=%d, user_id=%d, new_status=%s", orderID, userID, newStatus)
	return &order, nil
}
