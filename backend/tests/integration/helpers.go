package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services/order"
	"github.com/pintuotuo/backend/services/payment"
	"github.com/pintuotuo/backend/services/product"
	"github.com/pintuotuo/backend/services/token"
	"github.com/pintuotuo/backend/services/user"
	"github.com/stretchr/testify/require"
)

// Test constants
const (
	TestTimeout      = 5
	TestUserID       = 9999
	TestMerchantID   = 100
	TestProductID    = 1
	TestProductPrice = 99.99
)

// Atomic counter for generating unique IDs (thread-safe, no race condition)
var uniqueIDCounter int64

// GenerateUniqueID generates a unique ID for test isolation in parallel tests
// Uses atomic operations to be thread-safe without requiring mutexes
func GenerateUniqueID() int {
	// Use atomic increment to generate unique IDs safely across parallel tests
	return int(atomic.AddInt64(&uniqueIDCounter, 1))
}

// TestServices holds all service instances for testing
type TestServices struct {
	DB             *sql.DB
	UserService    user.Service
	ProductService product.Service
	OrderService   order.Service
	PaymentService payment.Service
	TokenService   token.Service
	Logger         *log.Logger
}

// SetupPaymentTest initializes all services for integration testing
func SetupPaymentTest(t *testing.T) *TestServices {
	// Initialize database
	if err := config.InitDB(); err != nil {
		t.Fatalf("Failed to init test DB: %v", err)
	}

	// Initialize cache (idempotent - only initializes once)
	if err := cache.Init(); err != nil {
		t.Logf("Warning: Cache already initialized or failed to init: %v", err)
	}

	db := config.GetDB()

	logger := log.New(os.Stderr, "[TestIntegration] ", log.LstdFlags)

	// Initialize services
	tokenSvc := token.NewService(db, logger)
	userSvc := user.NewService(db, logger, tokenSvc)
	productSvc := product.NewService(db, logger)
	orderSvc := order.NewService(db, logger)
	paymentSvc := payment.NewService(db, orderSvc, logger, tokenSvc)

	return &TestServices{
		DB:             db,
		UserService:    userSvc,
		ProductService: productSvc,
		OrderService:   orderSvc,
		PaymentService: paymentSvc,
		TokenService:   tokenSvc,
		Logger:         logger,
	}
}

// TeardownPaymentTest cleans up test resources
func TeardownPaymentTest(t *testing.T, ts *TestServices) {
	// Don't close cache as it's a shared global resource
	// Connection pools close automatically
}

// SeedTestUser creates a test user
func SeedTestUser(t *testing.T, db *sql.DB, uniqueID int) int {
	// Add a random component to the email to ensure it's truly unique across parallel tests
	// even if the uniqueID somehow collides (e.g. across different test packages)
	email := fmt.Sprintf("test%d_%d_%d@example.com", uniqueID, time.Now().UnixNano(), rand.Intn(1000000))
	name := fmt.Sprintf("Test User %d", uniqueID)
	passwordHash := "hashed_password"
	var userID int
	err := db.QueryRow("INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3) RETURNING id", email, name, passwordHash).Scan(&userID)
	require.NoError(t, err, "SeedTestUser failed")
	return userID
}

// SeedTestProduct creates a test product and returns its ID and the merchant ID
func SeedTestProduct(t *testing.T, db *sql.DB, productID int) (int, int) {
	ctx := context.Background()

	// Ensure merchant exists first to avoid foreign key violation
	merchantID := SeedTestUser(t, db, GenerateUniqueID())

	// Insert product directly
	var id int
	err := db.QueryRowContext(
		ctx,
		"INSERT INTO products (name, description, price, stock, merchant_id, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		fmt.Sprintf("Test Product %d", productID),
		"Test product description",
		TestProductPrice,
		1000, // stock
		merchantID,
		"active",
	).Scan(&id)

	require.NoError(t, err)
	return id, merchantID
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
	var transactionID sql.NullString
	err := db.QueryRow(
		"SELECT id, user_id, order_id, amount, method, status, transaction_id, created_at, updated_at FROM payments WHERE id = $1",
		paymentID,
	).Scan(&p.ID, &p.UserID, &p.OrderID, &p.Amount, &p.Method, &p.Status, &transactionID, &p.CreatedAt, &p.UpdatedAt)

	if transactionID.Valid {
		tempID := transactionID.String
		p.TransactionID = &tempID
	}

	require.NoError(t, err)
	return &p
}

// GetOrderFromDB retrieves order directly from database
func GetOrderFromDB(t *testing.T, db *sql.DB, orderID int) *order.Order {
	var o order.Order
	var groupID sql.NullInt64
	err := db.QueryRow(
		"SELECT id, user_id, product_id, group_id, quantity, unit_price, total_price, status, created_at, updated_at FROM orders WHERE id = $1",
		orderID,
	).Scan(&o.ID, &o.UserID, &o.ProductID, &groupID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.CreatedAt, &o.UpdatedAt)

	if groupID.Valid {
		o.GroupID = int(groupID.Int64)
	}

	require.NoError(t, err)
	return &o
}

// CreateTestPaymentFlow creates a complete payment flow for testing
func CreateTestPaymentFlow(t *testing.T, ts *TestServices) (userID int, orderID int, paymentID int) {
	ctx := context.Background()

	// Generate unique ID for this test to avoid conflicts in parallel tests
	// Use GenerateUniqueID() which is thread-safe with atomic operations
	uniqueID := GenerateUniqueID()

	// Create user with unique ID
	userID = SeedTestUser(t, ts.DB, uniqueID)

	// Create product with unique ID
	productID, _ := SeedTestProduct(t, ts.DB, uniqueID)

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
		OutTradeNo:  fmt.Sprintf("%d", paymentID),
		TradeNo:     fmt.Sprintf("alipay_test_%d", paymentID),
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
		OutTradeNo:    fmt.Sprintf("%d", paymentID),
		TransactionID: fmt.Sprintf("wechat_test_%d", paymentID),
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
	// Skip cleanup in tests to allow debugging and avoid race conditions
	// unless we are in CI environment
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return
	}

	ctx := context.Background()
	// Delete data in reverse order of foreign key dependencies
	_, _ = db.ExecContext(ctx, "DELETE FROM group_members WHERE user_id = $1", userID)
	_, _ = db.ExecContext(ctx, "DELETE FROM payments WHERE user_id = $1", userID)
	_, _ = db.ExecContext(ctx, "DELETE FROM orders WHERE user_id = $1", userID)
	_, _ = db.ExecContext(ctx, "DELETE FROM groups WHERE creator_id = $1", userID)
	_, _ = db.ExecContext(ctx, "DELETE FROM products WHERE merchant_id = $1", userID)
	_, _ = db.ExecContext(ctx, "DELETE FROM tokens WHERE user_id = $1", userID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
}
