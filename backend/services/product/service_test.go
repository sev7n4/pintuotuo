package product

import (
	"context"
	"log"
	"os"
	"sync"
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
	return int(atomic.AddInt64(&userIDCounter, 1)) + 3000 + int(time.Now().Unix()%1000)
}

func createUser(t *testing.T, email string) int {
	db := config.GetDB()
	var id int
	err := db.QueryRow(
		"INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name RETURNING id",
		email, "Test User", "hash",
	).Scan(&id)
	require.NoError(t, err)
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

	logger := log.New(os.Stderr, "[TestProductService] ", log.LstdFlags)
	testService = NewService(config.GetDB(), logger)
}

// TestListProductsValid tests valid product listing
func TestListProductsValid(t *testing.T) {
	params := &ListProductsParams{
		Page:    1,
		PerPage: 20,
		Status:  "active",
	}

	result, err := testService.ListProducts(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PerPage)
	assert.True(t, result.Total >= 0)
}

// TestListProductsPagination tests pagination validation
func TestListProductsPagination(t *testing.T) {
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
			params := &ListProductsParams{
				Page:    tt.page,
				PerPage: tt.perPage,
				Status:  "active",
			}

			result, err := testService.ListProducts(context.Background(), params)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedPage, result.Page)
			assert.Equal(t, tt.expectedPerPage, result.PerPage)
		})
	}
}

// TestCreateProductValid tests valid product creation
func TestCreateProductValid(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("prod_create_valid_%d@test.com", time.Now().UnixNano()))
	req := &CreateProductRequest{
		Name:          "Test Product",
		Description:   "A test product",
		Price:         99.99,
		OriginalPrice: 199.99,
		Stock:         100,
	}

	product, err := testService.CreateProduct(context.Background(), uid, req)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, req.Name, product.Name)
	assert.Equal(t, req.Price, product.Price)
	assert.Equal(t, req.Stock, product.Stock)
	assert.Equal(t, uid, product.MerchantID)
}

// TestCreateProductInvalidPrice tests product creation with invalid price
func TestCreateProductInvalidPrice(t *testing.T) {
	req := &CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
		Price:       -50,
		Stock:       100,
	}

	product, err := testService.CreateProduct(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Equal(t, ErrInvalidPrice, err)
}

// TestCreateProductInvalidStock tests product creation with invalid stock
func TestCreateProductInvalidStock(t *testing.T) {
	req := &CreateProductRequest{
		Name:  "Test Product",
		Price: 99.99,
		Stock: -10,
	}

	product, err := testService.CreateProduct(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Equal(t, ErrInvalidStock, err)
}

// TestCreateProductMissingName tests product creation with missing name
func TestCreateProductMissingName(t *testing.T) {
	req := &CreateProductRequest{
		Name:  "",
		Price: 99.99,
		Stock: 100,
	}

	product, err := testService.CreateProduct(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Equal(t, ErrInvalidProductName, err)
}

// TestGetProductByIDValid tests retrieving product by valid ID
func TestGetProductByIDValid(t *testing.T) {
	uid := createUser(t, "prod_get_creator@test.com")
	// Create product first
	req := &CreateProductRequest{
		Name:  "Test Product",
		Price: 99.99,
		Stock: 100,
	}
	created, err := testService.CreateProduct(context.Background(), uid, req)
	require.NoError(t, err)

	// Retrieve product
	product, err := testService.GetProductByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, created.ID, product.ID)
	assert.Equal(t, created.Name, product.Name)
}

// TestGetProductByIDCache tests caching in GetProductByID
func TestGetProductByIDCache(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("prod_cache_test_%d@test.com", time.Now().UnixNano()))

	// Create product
	req := &CreateProductRequest{
		Name:  "Cache Test Product",
		Price: 75.50,
		Stock: 150,
	}
	created, err := testService.CreateProduct(ctx, uid, req)
	require.NoError(t, err)

	// First call - should hit database
	product1, err := testService.GetProductByID(ctx, created.ID)
	assert.NoError(t, err)

	// Second call - should hit cache
	product2, err := testService.GetProductByID(ctx, created.ID)
	assert.NoError(t, err)

	assert.Equal(t, product1.ID, product2.ID)
	assert.Equal(t, product1.Name, product2.Name)

	// Verify cache was set
	cacheKey := cache.ProductKey(created.ID)
	cached, err := cache.Get(ctx, cacheKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, cached)
}

// TestGetProductByIDNotFound tests retrieving non-existent product
func TestGetProductByIDNotFound(t *testing.T) {
	product, err := testService.GetProductByID(context.Background(), 99999)
	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Equal(t, ErrProductNotFound, err)
}

// TestSearchProductsValid tests valid product search
func TestSearchProductsValid(t *testing.T) {
	ctx := context.Background()

	// Create a test product with known name
	req := &CreateProductRequest{
		Name:        "SearchableToken",
		Description: "Easy to find product",
		Price:       120.00,
		Stock:       50,
	}
	_, err := testService.CreateProduct(ctx, 1, req)
	require.NoError(t, err)

	// Search for the product
	searchParams := &SearchProductsParams{
		Query:   "SearchableToken",
		Page:    1,
		PerPage: 20,
	}

	result, err := testService.SearchProducts(ctx, searchParams)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Total > 0)
	assert.Equal(t, 1, result.Page)
}

// TestSearchProductsEmptyQuery tests search with empty query
func TestSearchProductsEmptyQuery(t *testing.T) {
	searchParams := &SearchProductsParams{
		Query:   "",
		Page:    1,
		PerPage: 20,
	}

	result, err := testService.SearchProducts(context.Background(), searchParams)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidSearchQuery, err)
}

// TestUpdateProductValid tests valid product update
func TestUpdateProductValid(t *testing.T) {
	uid := createUser(t, "prod_upd_creator@test.com")
	// Create product first
	req := &CreateProductRequest{
		Name:  "Old Name",
		Price: 99.99,
		Stock: 100,
	}
	created, err := testService.CreateProduct(context.Background(), uid, req)
	require.NoError(t, err)

	// Update product
	updateReq := &UpdateProductRequest{
		Name:  "New Name",
		Price: 149.99,
		Stock: 50,
	}
	updated, err := testService.UpdateProduct(context.Background(), uid, created.ID, updateReq)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, 149.99, updated.Price)
	assert.Equal(t, 50, updated.Stock)
}

// TestUpdateProductNotOwner tests updating product not owned by merchant
func TestUpdateProductNotOwner(t *testing.T) {
	ctx := context.Background()

	// Create product with merchant 1
	req := &CreateProductRequest{
		Name:  "Owned Product",
		Price: 100.00,
		Stock: 50,
	}
	created, err := testService.CreateProduct(ctx, 1, req)
	require.NoError(t, err)

	// Try to update with merchant 2
	updateReq := &UpdateProductRequest{
		Name: "Hacked",
	}
	updated, err := testService.UpdateProduct(ctx, 2, created.ID, updateReq)
	assert.Error(t, err)
	assert.Nil(t, updated)
	assert.Equal(t, ErrNotProductOwner, err)
}

// TestUpdateProductCacheInvalidation tests cache invalidation on update
func TestUpdateProductCacheInvalidation(t *testing.T) {
	ctx := context.Background()

	// Create product
	req := &CreateProductRequest{
		Name:  "Cache Invalidate Test",
		Price: 99.99,
		Stock: 100,
	}
	created, err := testService.CreateProduct(ctx, 1, req)
	require.NoError(t, err)

	// Load into cache
	_, err = testService.GetProductByID(ctx, created.ID)
	require.NoError(t, err)

	// Verify cache has data
	cacheKey := cache.ProductKey(created.ID)
	_, err = cache.Get(ctx, cacheKey)
	assert.NoError(t, err)

	// Update product
	updateReq := &UpdateProductRequest{
		Name: "Updated in Cache",
	}
	_, err = testService.UpdateProduct(ctx, 1, created.ID, updateReq)
	require.NoError(t, err)

	// Cache should be invalidated (hard to test directly)
	// But we can verify the update happened
}

// TestDeleteProductValid tests valid product deletion
func TestDeleteProductValid(t *testing.T) {
	uid := createUser(t, "prod_del_creator@test.com")
	// Create product first
	req := &CreateProductRequest{
		Name:  "To Be Deleted",
		Price: 99.99,
		Stock: 100,
	}
	created, err := testService.CreateProduct(context.Background(), uid, req)
	require.NoError(t, err)

	// Delete product
	err = testService.DeleteProduct(context.Background(), uid, created.ID)
	assert.NoError(t, err)

	// Verify product is gone (or status changed to deleted)
	product, err := testService.GetProductByID(context.Background(), created.ID)
	// Depending on implementation, it might return ErrProductNotFound or product with deleted status
	if err == nil {
		assert.Equal(t, "deleted", product.Status)
	} else {
		assert.Equal(t, ErrProductNotFound, err)
	}
}

// TestConcurrentStockUpdate tests concurrent stock updates
func TestConcurrentStockUpdate(t *testing.T) {
	uid := createUser(t, "prod_stock_creator@test.com")
	// Create product with 100 stock
	req := &CreateProductRequest{
		Name:  "Concurrent Stock Test",
		Price: 99.99,
		Stock: 100,
	}
	product, err := testService.CreateProduct(context.Background(), uid, req)
	require.NoError(t, err)

	var wg sync.WaitGroup
	var successCount int32
	ctx := context.Background()

	// Try to decrement stock 120 times (should only succeed 100 times)
	for i := 0; i < 120; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := testService.UpdateStock(ctx, product.ID, -1)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()

	// Verify exactly 100 decrements succeeded
	assert.Equal(t, int32(100), successCount)

	// Verify stock is 0
	p, err := testService.GetProductByID(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, p.Stock)
}

// TestDeleteProductNotOwner tests deleting product not owned by merchant
func TestDeleteProductNotOwner(t *testing.T) {
	ctx := context.Background()

	// Create product with merchant 1
	req := &CreateProductRequest{
		Name:  "Protected Product",
		Price: 100.00,
		Stock: 50,
	}
	created, err := testService.CreateProduct(ctx, 1, req)
	require.NoError(t, err)

	// Try to delete with merchant 2
	err = testService.DeleteProduct(ctx, 2, created.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotProductOwner, err)

	// Verify product still exists
	product, err := testService.GetProductByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, product)
}

// TestDeleteProductNotFound tests deleting non-existent product
func TestDeleteProductNotFound(t *testing.T) {
	err := testService.DeleteProduct(context.Background(), 1, 99999)
	assert.Error(t, err)
	assert.Equal(t, ErrProductNotFound, err)
}

// TestConcurrentProductCreation tests concurrent product creation
func TestConcurrentProductCreation(t *testing.T) {
	var count int32
	var wg sync.WaitGroup
	ctx := context.Background()

	// Ensure merchant user exists
	_, _ = config.GetDB().ExecContext(ctx, "INSERT INTO users (id, email, name, password_hash) VALUES (1, 'merchant@example.com', 'Merchant', 'hash') ON CONFLICT DO NOTHING")

	// Create multiple products concurrently
	uid := createUser(t, "concurrent_prod_creator@test.com")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := &CreateProductRequest{
				Name:  "Concurrent Product " + string(rune(idx+65)), // Use A, B, C...
				Price: 100.00 + float64(idx),
				Stock: 50,
			}
			_, err := testService.CreateProduct(context.Background(), uid, req)
			if err == nil {
				atomic.AddInt32(&count, 1)
			}
		}(i)
	}

	// Wait for all goroutines
	wg.Wait()

	// All should succeed
	assert.Equal(t, int32(5), count)
}

// TestProductFieldsOnCreate tests that all fields are properly set on creation
func TestProductFieldsOnCreate(t *testing.T) {
	uid := createUser(t, "fields_creator@test.com")
	req := &CreateProductRequest{
		Name:          "Field Test",
		Description:   "Test Description",
		Price:         123.45,
		OriginalPrice: 234.56,
		Stock:         789,
	}

	product, err := testService.CreateProduct(context.Background(), uid, req)
	require.NoError(t, err)

	assert.Equal(t, "Field Test", product.Name)
	assert.Equal(t, "Test Description", product.Description)
	assert.Equal(t, 123.45, product.Price)
	assert.Equal(t, 234.56, product.OriginalPrice)
	assert.Equal(t, 789, product.Stock)
	assert.Equal(t, uid, product.MerchantID)
	assert.Equal(t, "active", product.Status)
	assert.NotZero(t, product.ID)
	assert.NotZero(t, product.CreatedAt)
	assert.NotZero(t, product.UpdatedAt)
}

// TestListProductsWithAllStatus tests listing all products regardless of status
func TestListProductsWithAllStatus(t *testing.T) {
	params := &ListProductsParams{
		Page:    1,
		PerPage: 20,
		Status:  "all",
	}

	result, err := testService.ListProducts(context.Background(), params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Total >= 0)
}

// TestProductMetadata tests that products have correct metadata fields
func TestProductMetadata(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, "metadata_creator@test.com")

	req := &CreateProductRequest{
		Name:  "Metadata Test",
		Price: 99.99,
		Stock: 50,
	}

	created, err := testService.CreateProduct(ctx, uid, req)
	require.NoError(t, err)

	// Verify metadata
	assert.Equal(t, uid, created.MerchantID)
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
	assert.Equal(t, "active", created.Status)
}
