package integration

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/payment"
	"github.com/pintuotuo/backend/services/product"
	"github.com/pintuotuo/backend/services/user"
)

// Test constants
const (
	TestTimeout      = 5
	TestUserID       = 9999
	TestMerchantID   = 100
	TestProductID    = 1
	TestProductPrice = 99.99
)

// TestServices holds all service instances for testing
type TestServices struct {
	DB             *sql.DB
	UserService    user.Service
	ProductService product.Service
	OrderService   order.Service
	PaymentService payment.Service
	Logger         *log.Logger
}

// SetupPaymentTest initializes all services for integration testing
func SetupPaymentTest(t *testing.T) *TestServices {
	// Initialize database
	if err := config.InitDB(); err != nil {
		t.Fatalf("Failed to init test DB: %v", err)
	}

	// Initialize cache
	if err := cache.Init(); err != nil {
		t.Fatalf("Failed to init cache: %v", err)
	}

	db := config.GetDB()
	logger := log.New(os.Stderr, "[TestIntegration] ", log.LstdFlags)

	// Initialize services
	userSvc := user.NewService(db, logger)
	productSvc := product.NewService(db, logger)
	orderSvc := order.NewService(db, logger)
	paymentSvc := payment.NewService(db, orderSvc, logger)

	return &TestServices{
		DB:             db,
		UserService:    userSvc,
		ProductService: productSvc,
		OrderService:   orderSvc,
		PaymentService: paymentSvc,
		Logger:         logger,
	}
}

// TeardownPaymentTest cleans up test resources
func TeardownPaymentTest(t *testing.T, ts *TestServices) {
	// Clean up cache
	cache.Close()
	// Connection pools close automatically
}

// SeedTestUser creates a test user
func SeedTestUser(t *testing.T, db *sql.DB, userID int) int {
	ctx := context.Background()
	logger := log.New(os.Stderr, "[SeedUser] ", log.LstdFlags)
	userService := user.NewService(db, logger)

	req := &user.RegisterRequest{
		Email:    "test" + string(rune(userID)) + "@example.com",
		Password: "TestPassword123!",
		Name:     "Test User",
	}

	registeredUser, err := userService.RegisterUser(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, registeredUser)

	return registeredUser.ID
}

// SeedTestProduct creates a test product
func SeedTestProduct(t *testing.T, db *sql.DB, productID int) int {
	ctx := context.Background()

	// Insert product directly
	var id int
	err := db.QueryRowContext(
		ctx,
		"INSERT INTO products (name, description, price, stock, merchant_id, category, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		"Test Product "+string(rune(productID)),
		"Test product description",
		TestProductPrice,
		1000, // stock
		TestMerchantID,
		"test",
		"active",
	).Scan(&id)

	require.NoError(t, err)
	return id
}

// SeedTestOrder creates a test order
func SeedTestOrder(t *testing.T, db *sql.DB, userID int, productID int) int {
	ctx := context.Background()
	logger := log.New(os.Stderr, "[SeedOrder] ", log.LstdFlags)
	orderService := order.NewService(db, logger)

	req := &order.CreateOrderRequest{
		ProductID: productID,
		Quantity:  1,
	}

	createdOrder, err := orderService.CreateOrder(ctx, userID, req)
	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	return createdOrder.ID
}

// AssertPaymentStatus verifies payment status in database
func AssertPaymentStatus(t *testing.T, db *sql.DB, paymentID int, expected string) {
	var status string
	err := db.QueryRow("SELECT status FROM payments WHERE id = $1", paymentID).Scan(&status)
	require.NoError(t, err)
	require.Equal(t, expected, status, "Payment status should be %s", expected)
}

// AssertOrderStatus verifies order status in database
func AssertOrderStatus(t *testing.T, db *sql.DB, orderID int, expected string) {
	var status string
	err := db.QueryRow("SELECT status FROM orders WHERE id = $1", orderID).Scan(&status)
	require.NoError(t, err)
	require.Equal(t, expected, status, "Order status should be %s", expected)
}

// AssertCacheKeyExists verifies cache key exists
func AssertCacheKeyExists(t *testing.T, ctx context.Context, key string) {
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	require.True(t, exists, "Cache key %s should exist", key)
}

// AssertCacheKeyNotExists verifies cache key doesn't exist
func AssertCacheKeyNotExists(t *testing.T, ctx context.Context, key string) {
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	require.False(t, exists, "Cache key %s should not exist", key)
}

// GetPaymentFromDB retrieves payment directly from database
func GetPaymentFromDB(t *testing.T, db *sql.DB, paymentID int) *payment.Payment {
	var p payment.Payment
	err := db.QueryRow(
		"SELECT id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at FROM payments WHERE id = $1",
		paymentID,
	).Scan(&p.ID, &p.UserID, &p.OrderID, &p.Amount, &p.Method, &p.Status, &p.TransactionID, &p.CreatedAt, &p.UpdatedAt)

	require.NoError(t, err)
	return &p
}

// GetOrderFromDB retrieves order directly from database
func GetOrderFromDB(t *testing.T, db *sql.DB, orderID int) *order.Order {
	var o order.Order
	err := db.QueryRow(
		"SELECT id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at FROM orders WHERE id = $1",
		orderID,
	).Scan(&o.ID, &o.UserID, &o.ProductID, &o.GroupID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt, &o.UpdatedAt)

	require.NoError(t, err)
	return &o
}

// CreateTestPaymentFlow creates a complete payment flow for testing
func CreateTestPaymentFlow(t *testing.T, ts *TestServices) (userID int, orderID int, paymentID int) {
	ctx := context.Background()

	// Create user
	userID = SeedTestUser(t, ts.DB, 1)

	// Create product
	productID := SeedTestProduct(t, ts.DB, 1)

	// Create order
	orderID = SeedTestOrder(t, ts.DB, userID, productID)

	// Initiate payment
	paymentReq := &payment.InitiatePaymentRequest{
		OrderID:       orderID,
		PaymentMethod: "alipay",
	}
	p, err := ts.PaymentService.InitiatePayment(ctx, userID, paymentReq)
	require.NoError(t, err)
	require.NotNil(t, p)

	return userID, orderID, p.ID
}

// SimulateAlipayCallback simulates a successful Alipay webhook callback
func SimulateAlipayCallback(t *testing.T, ctx context.Context, db *sql.DB, paymentService payment.Service, paymentID int) *payment.Payment {
	// Get payment to extract order details
	p := GetPaymentFromDB(t, db, paymentID)

	callback := &payment.AlipayCallback{
		OutTradeNo:  string(rune(paymentID)),
		TradeNo:     "alipay_test_" + string(rune(paymentID)),
		TotalAmount: p.Amount,
		TradeStatus: "TRADE_SUCCESS",
		Timestamp:   "2026-03-15 12:00:00",
		Sign:        "test_signature",
	}

	result, err := paymentService.HandleAlipayCallback(ctx, callback)
	require.NoError(t, err)
	require.NotNil(t, result)

	return result
}

// SimulateWechatCallback simulates a successful WeChat webhook callback
func SimulateWechatCallback(t *testing.T, ctx context.Context, db *sql.DB, paymentService payment.Service, paymentID int) *payment.Payment {
	// Get payment to extract order details
	p := GetPaymentFromDB(t, db, paymentID)

	callback := &payment.WechatCallback{
		OutTradeNo:    string(rune(paymentID)),
		TransactionID: "wechat_test_" + string(rune(paymentID)),
		TotalFee:      int(p.Amount * 100), // Convert to cents
		ResultCode:    "SUCCESS",
		Sign:          "test_signature",
	}

	result, err := paymentService.HandleWechatCallback(ctx, callback)
	require.NoError(t, err)
	require.NotNil(t, result)

	return result
}

// CleanupTestData removes test data from database
func CleanupTestData(t *testing.T, db *sql.DB, userID int) {
	ctx := context.Background()

	// Delete payments
	_, err := db.ExecContext(ctx, "DELETE FROM payments WHERE user_id = $1", userID)
	require.NoError(t, err)

	// Delete orders
	_, err = db.ExecContext(ctx, "DELETE FROM orders WHERE user_id = $1", userID)
	require.NoError(t, err)

	// Delete user tokens
	_, err = db.ExecContext(ctx, "DELETE FROM user_tokens WHERE user_id = $1", userID)
	require.NoError(t, err)

	// Delete user
	_, err = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
	require.NoError(t, err)
}
