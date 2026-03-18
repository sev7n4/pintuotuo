package order

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
)

var testService Service
var userIDCounter int64

func generateUserID() int {
	return int(atomic.AddInt64(&userIDCounter, 1)) + 1000 + int(time.Now().Unix()%1000)
}

func createUser(t *testing.T, email string) int {
	db := config.GetDB()
	var id int
	err := db.QueryRow("INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3) RETURNING id", email, "Test User", "hash").Scan(&id)
	require.NoError(t, err)

	// Initialize tokens
	_, _ = db.Exec("INSERT INTO tokens (user_id, balance) VALUES ($1, 1000.00) ON CONFLICT DO NOTHING", id)
	return id
}

func init() {
	// Initialize test database
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to init test DB: %v", err)
	}

	// Initialize cache
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to init cache: %v", err)
	}

	logger := log.New(os.Stderr, "[TestOrderService] ", log.LstdFlags)
	testService = NewService(config.GetDB(), logger)
}

// TestCreateOrderValid tests valid order creation
func TestCreateOrderValid(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("valid%d@test.com", generateUserID()))
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  2,
		GroupID:   0,
	}

	order, err := testService.CreateOrder(context.Background(), uid, req)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, req.ProductID, order.ProductID)
	assert.Equal(t, req.Quantity, order.Quantity)
	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, uid, order.UserID)
	assert.True(t, order.ID > 0)
	assert.True(t, order.TotalPrice > 0)
	assert.True(t, order.UnitPrice > 0)
}

// TestCreateOrderInvalidQuantity tests order creation with invalid quantity
func TestCreateOrderInvalidQuantity(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("invalid_qty%d@test.com", generateUserID()))
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  0,
		GroupID:   0,
	}

	order, err := testService.CreateOrder(context.Background(), uid, req)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, ErrInvalidQuantity, err)
}

// TestCreateOrderInsufficientStock tests order creation with insufficient stock
func TestCreateOrderInsufficientStock(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("stock%d@test.com", generateUserID()))
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  99999, // Assume product doesn't have this much stock
		GroupID:   0,
	}

	order, err := testService.CreateOrder(context.Background(), uid, req)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, ErrInsufficientStock, err)
}

// TestCreateOrderProductNotFound tests order creation with non-existent product
func TestCreateOrderProductNotFound(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("notfound%d@test.com", generateUserID()))
	req := &CreateOrderRequest{
		ProductID: 99999,
		Quantity:  1,
		GroupID:   0,
	}

	order, err := testService.CreateOrder(context.Background(), uid, req)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, ErrProductNotFound, err)
}

// TestListOrdersValid tests valid order listing
func TestListOrdersValid(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("list%d@test.com", generateUserID()))
	params := &ListOrdersParams{
		Page:    1,
		PerPage: 20,
		Status:  "all",
	}

	result, err := testService.ListOrders(context.Background(), uid, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PerPage)
	assert.True(t, result.Total >= 0)
}

// TestListOrdersPagination tests pagination validation
func TestListOrdersPagination(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("pagination%d@test.com", generateUserID()))
	tests := []struct {
		name            string
		page            int
		perPage         int
		expectedPage    int
		expectedPerPage int
	}{
		{name: "Valid first page", page: 1, perPage: 20, expectedPage: 1, expectedPerPage: 20},
		{name: "Invalid page 0", page: 0, perPage: 20, expectedPage: 1, expectedPerPage: 20},
		{name: "Invalid perPage 0", page: 1, perPage: 0, expectedPage: 1, expectedPerPage: 20},
		{name: "PerPage exceeds max", page: 1, perPage: 150, expectedPage: 1, expectedPerPage: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &ListOrdersParams{
				Page:    tt.page,
				PerPage: tt.perPage,
				Status:  "all",
			}

			result, err := testService.ListOrders(context.Background(), uid, params)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedPage, result.Page)
			assert.Equal(t, tt.expectedPerPage, result.PerPage)
		})
	}
}

// TestListOrdersByStatus tests filtering orders by status
func TestListOrdersByStatus(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("status%d@test.com", generateUserID()))

	// Create an order
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	_, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)

	// List pending orders
	params := &ListOrdersParams{
		Page:    1,
		PerPage: 20,
		Status:  "pending",
	}

	result, err := testService.ListOrders(ctx, uid, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Total > 0)

	// Verify orders are pending
	for _, order := range result.Data {
		assert.Equal(t, "pending", order.Status)
	}
}

// TestGetOrderByIDValid tests retrieving order by valid ID
func TestGetOrderByIDValid(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("get%d@test.com", generateUserID()))

	// Create order first
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	created, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)

	// Retrieve order
	order, err := testService.GetOrderByID(ctx, uid, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, created.ID, order.ID)
	assert.Equal(t, uid, order.UserID)
}

// TestGetOrderByIDNotOwner tests retrieving order by non-owner
func TestGetOrderByIDNotOwner(t *testing.T) {
	ctx := context.Background()
	uid1 := createUser(t, fmt.Sprintf("owner1%d@test.com", generateUserID()))
	uid2 := createUser(t, fmt.Sprintf("owner2%d@test.com", generateUserID()))

	// Create order as user 1
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	created, err := testService.CreateOrder(ctx, uid1, req)
	require.NoError(t, err)

	// Try to retrieve as user 2
	order, err := testService.GetOrderByID(ctx, uid2, created.ID)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, ErrOrderNotFound, err)
}

// TestGetOrderByIDNotFound tests retrieving non-existent order
func TestGetOrderByIDNotFound(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("getnotfound%d@test.com", generateUserID()))
	order, err := testService.GetOrderByID(context.Background(), uid, 99999)
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Equal(t, ErrOrderNotFound, err)
}

// TestGetOrdersByStatusValid tests retrieving orders by status
func TestGetOrdersByStatusValid(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("statusvalid%d@test.com", generateUserID()))

	// Create a pending order
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	_, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)

	// Get pending orders
	orders, err := testService.GetOrdersByStatus(ctx, uid, "pending")
	assert.NoError(t, err)
	assert.NotNil(t, orders)
	assert.True(t, len(orders) > 0)

	// Verify all are pending
	for _, order := range orders {
		assert.Equal(t, "pending", order.Status)
		assert.Equal(t, uid, order.UserID)
	}
}

// TestGetOrdersByStatusEmptyStatus tests with invalid status
func TestGetOrdersByStatusEmptyStatus(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("statusinvalid%d@test.com", generateUserID()))
	orders, err := testService.GetOrdersByStatus(context.Background(), uid, "")
	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.Equal(t, ErrInvalidStatus, err)
}

// TestCancelOrderValid tests valid order cancellation
func TestCancelOrderValid(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("cancel%d@test.com", generateUserID()))

	// Create order
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	created, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)
	assert.Equal(t, "pending", created.Status)

	// Cancel order
	cancelled, err := testService.CancelOrder(ctx, uid, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, cancelled)
	assert.Equal(t, "cancelled", cancelled.Status)
	assert.Equal(t, created.ID, cancelled.ID)
}

// TestCancelOrderNotFound tests cancelling non-existent order
func TestCancelOrderNotFound(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("cancelnotfound%d@test.com", generateUserID()))
	cancelled, err := testService.CancelOrder(context.Background(), uid, 99999)
	assert.Error(t, err)
	assert.Nil(t, cancelled)
	assert.Equal(t, ErrOrderNotFound, err)
}

// TestCancelOrderNotPending tests cancelling non-pending order
func TestCancelOrderNotPending(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("cancelnotpending%d@test.com", generateUserID()))

	// Create order
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	created, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)

	// Update to paid
	_, err = testService.UpdateOrderStatus(ctx, uid, created.ID, "paid")
	require.NoError(t, err)

	// Try to cancel
	cancelled, err := testService.CancelOrder(ctx, uid, created.ID)
	assert.Error(t, err)
	assert.Nil(t, cancelled)
	assert.Equal(t, ErrCannotCancelOrder, err)
}

// TestUpdateOrderStatusValid tests valid status update
func TestUpdateOrderStatusValid(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("update%d@test.com", generateUserID()))

	// Create order
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	created, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)

	// Update status to paid
	updated, err := testService.UpdateOrderStatus(ctx, uid, created.ID, "paid")
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "paid", updated.Status)
	assert.Equal(t, created.ID, updated.ID)
}

// TestUpdateOrderStatusInvalid tests updating with invalid status
func TestUpdateOrderStatusInvalid(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("updateinvalid%d@test.com", generateUserID()))

	// Create order
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	created, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)

	// Try invalid status
	updated, err := testService.UpdateOrderStatus(ctx, uid, created.ID, "invalid_status")
	assert.Error(t, err)
	assert.Nil(t, updated)
	assert.Equal(t, ErrInvalidStatus, err)
}

// TestOrderCalculation tests order price calculation
func TestOrderCalculation(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("calc%d@test.com", generateUserID()))
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  3,
		GroupID:   0,
	}

	order, err := testService.CreateOrder(context.Background(), uid, req)
	require.NoError(t, err)

	// Total price should be unit_price * quantity
	expectedTotal := order.UnitPrice * float64(req.Quantity)
	assert.InDelta(t, expectedTotal, order.TotalPrice, 0.001)
	assert.Equal(t, 3, order.Quantity)
}

// TestConcurrentOrderCreation tests concurrent order creation
func TestConcurrentOrderCreation(t *testing.T) {
	done := make(chan bool)
	count := 0

	// Create multiple orders concurrently
	for i := 0; i < 5; i++ {
		go func(idx int) {
			uid := createUser(t, fmt.Sprintf("concurrent%d_%d@test.com", idx, generateUserID()))
			req := &CreateOrderRequest{
				ProductID: 1,
				Quantity:  1,
				GroupID:   0,
			}
			_, err := testService.CreateOrder(context.Background(), uid, req)
			if err == nil {
				count++
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// All should succeed (or some fail due to stock)
	assert.True(t, count > 0)
}

// TestOrderFieldsOnCreate tests that all fields are properly set on creation
func TestOrderFieldsOnCreate(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("fields%d@test.com", generateUserID()))
	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  2,
		GroupID:   5,
	}

	// Note: In a real environment, group ID 5 might not exist, but we assume
	// foreign key check is disabled or we seed it.
	// Actually, I should probably create a group first to be safe.
	db := config.GetDB()
	_, _ = db.Exec("INSERT INTO groups (id, product_id, creator_id, target_count, current_count, deadline) VALUES (5, 1, 1, 10, 1, NOW() + INTERVAL '1 day') ON CONFLICT DO NOTHING")

	order, err := testService.CreateOrder(context.Background(), uid, req)
	require.NoError(t, err)

	assert.Equal(t, 1, order.ProductID)
	assert.Equal(t, uid, order.UserID)
	assert.Equal(t, 5, order.GroupID)
	assert.Equal(t, 2, order.Quantity)
	assert.True(t, order.UnitPrice > 0)
	assert.True(t, order.TotalPrice > 0)
	assert.Equal(t, "pending", order.Status)
	assert.NotZero(t, order.ID)
	assert.NotZero(t, order.CreatedAt)
	assert.NotZero(t, order.UpdatedAt)
}

// TestOrderStatusTransitions tests valid status transitions
func TestOrderStatusTransitions(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("transitions%d@test.com", generateUserID()))

	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   0,
	}
	created, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)

	// Test transition: pending -> paid
	paid, err := testService.UpdateOrderStatus(ctx, uid, created.ID, "paid")
	assert.NoError(t, err)
	assert.Equal(t, "paid", paid.Status)

	// Test transition: paid -> completed
	completed, err := testService.UpdateOrderStatus(ctx, uid, created.ID, "completed")
	assert.NoError(t, err)
	assert.Equal(t, "completed", completed.Status)
}

// TestOrderWithGroupID tests order creation with group ID
func TestOrderWithGroupID(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("groupid%d@test.com", generateUserID()))

	// Create a group first to satisfy foreign key constraint
	db := config.GetDB()
	var groupID int
	err := db.QueryRowContext(ctx, "INSERT INTO groups (product_id, creator_id, target_count, current_count, status, deadline) VALUES (1, 1, 5, 1, 'active', NOW() + INTERVAL '1 day') RETURNING id").Scan(&groupID)
	require.NoError(t, err)

	req := &CreateOrderRequest{
		ProductID: 1,
		Quantity:  1,
		GroupID:   groupID,
	}

	order, err := testService.CreateOrder(ctx, uid, req)
	require.NoError(t, err)

	assert.Equal(t, groupID, order.GroupID)
	assert.Equal(t, "pending", order.Status)
}
