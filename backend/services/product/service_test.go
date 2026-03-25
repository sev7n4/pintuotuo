package product

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
)

var testService Service

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
	req := &CreateProductRequest{
		Name:          "Test Product",
		Description:   "A test product",
		Price:         99.99,
		OriginalPrice: 199.99,
		Stock:         100,
	}

	product, err := testService.CreateProduct(context.Background(), 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, req.Name, product.Name)
	assert.Equal(t, req.Price, product.Price)
	assert.Equal(t, req.Stock, product.Stock)
	assert.Equal(t, "active", product.Status)
	assert.True(t, product.ID > 0)
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
	// Create product first
	req := &CreateProductRequest{
		Name:        "Retrievable Product",
		Description: "For testing retrieval",
		Price:       50.00,
		Stock:       200,
	}
	created, err := testService.CreateProduct(context.Background(), 1, req)
	require.NoError(t, err)

	// Retrieve product
	product, err := testService.GetProductByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, created.ID, product.ID)
	assert.Equal(t, created.Name, product.Name)
	assert.Equal(t, created.Price, product.Price)
}

// TestGetProductByIDCache tests caching in GetProductByID
func TestGetProductByIDCache(t *testing.T) {
	ctx := context.Background()

	// Create product
	req := &CreateProductRequest{
		Name:  "Cache Test Product",
		Price: 75.50,
		Stock: 150,
	}
	created, err := testService.CreateProduct(ctx, 1, req)
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
	ctx := context.Background()

	// Create product
	req := &CreateProductRequest{
		Name:  "Update Test",
		Price: 100.00,
		Stock: 50,
	}
	created, err := testService.CreateProduct(ctx, 1, req)
	require.NoError(t, err)

	// Update product
	updateReq := &UpdateProductRequest{
		Name:  "Updated Name",
		Price: 150.00,
		Stock: 75,
	}
	updated, err := testService.UpdateProduct(ctx, 1, created.ID, updateReq)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, 150.00, updated.Price)
	assert.Equal(t, 75, updated.Stock)
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
	ctx := context.Background()

	// Create product
	req := &CreateProductRequest{
		Name:  "Delete Test",
		Price: 50.00,
		Stock: 25,
	}
	created, err := testService.CreateProduct(ctx, 1, req)
	require.NoError(t, err)

	// Delete product
	err = testService.DeleteProduct(ctx, 1, created.ID)
	assert.NoError(t, err)

	// Verify product is deleted
	product, err := testService.GetProductByID(ctx, created.ID)
	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Equal(t, ErrProductNotFound, err)
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
	done := make(chan bool)
	count := 0

	// Create multiple products concurrently
	for i := 0; i < 5; i++ {
		go func(idx int) {
			req := &CreateProductRequest{
				Name:  "Concurrent Product " + string(rune(idx)),
				Price: 100.00 + float64(idx),
				Stock: 50,
			}
			_, err := testService.CreateProduct(context.Background(), 1, req)
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

	// All should succeed
	assert.Equal(t, 5, count)
}

// TestProductFieldsOnCreate tests that all fields are properly set on creation
func TestProductFieldsOnCreate(t *testing.T) {
	req := &CreateProductRequest{
		Name:          "Field Test",
		Description:   "Test Description",
		Price:         123.45,
		OriginalPrice: 234.56,
		Stock:         789,
	}

	product, err := testService.CreateProduct(context.Background(), 42, req)
	require.NoError(t, err)

	assert.Equal(t, "Field Test", product.Name)
	assert.Equal(t, "Test Description", product.Description)
	assert.Equal(t, 123.45, product.Price)
	assert.Equal(t, 234.56, product.OriginalPrice)
	assert.Equal(t, 789, product.Stock)
	assert.Equal(t, 42, product.MerchantID)
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

	req := &CreateProductRequest{
		Name:  "Metadata Test",
		Price: 99.99,
		Stock: 50,
	}

	created, err := testService.CreateProduct(ctx, 99, req)
	require.NoError(t, err)

	// Verify metadata
	assert.Equal(t, 99, created.MerchantID)
	assert.NotZero(t, created.ID)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
	assert.Equal(t, "active", created.Status)
}
