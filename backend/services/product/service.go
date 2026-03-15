package product

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/pintuotuo/backend/cache"
)

// Service defines the product service interface
type Service interface {
	// Read operations
	ListProducts(ctx context.Context, params *ListProductsParams) (*ListProductsResult, error)
	GetProductByID(ctx context.Context, productID int) (*Product, error)
	SearchProducts(ctx context.Context, params *SearchProductsParams) (*ListProductsResult, error)

	// Write operations (merchant only)
	CreateProduct(ctx context.Context, merchantID int, req *CreateProductRequest) (*Product, error)
	UpdateProduct(ctx context.Context, merchantID int, productID int, req *UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, merchantID int, productID int) error
}

// service implements the Service interface
type service struct {
	db  *sql.DB
	log *log.Logger
}

// NewService creates a new product service
func NewService(db *sql.DB, logger *log.Logger) Service {
	if logger == nil {
		logger = log.New(os.Stderr, "[ProductService] ", log.LstdFlags)
	}

	return &service{
		db:  db,
		log: logger,
	}
}

// ListProducts retrieves products with pagination and caching
func (s *service) ListProducts(ctx context.Context, params *ListProductsParams) (*ListProductsResult, error) {
	// Validate parameters
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}
	if params.Status == "" {
		params.Status = "active"
	}

	// Try cache first
	cacheKey := cache.ProductListKey(params.Page, params.PerPage, params.Status)
	if cachedData, err := cache.Get(ctx, cacheKey); err == nil {
		var result ListProductsResult
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			return &result, nil
		}
	}

	// Query database
	offset := (params.Page - 1) * params.PerPage

	var rows *sql.Rows
	var err error

	if params.Status == "all" {
		rows, err = s.db.QueryContext(
			ctx,
			"SELECT id, merchant_id, name, description, price, COALESCE(original_price, 0), stock, status, created_at, updated_at FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			params.PerPage, offset,
		)
	} else {
		rows, err = s.db.QueryContext(
			ctx,
			"SELECT id, merchant_id, name, description, price, COALESCE(original_price, 0), stock, status, created_at, updated_at FROM products WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
			params.Status, params.PerPage, offset,
		)
	}

	if err != nil {
		return nil, wrapError("ListProducts", "query", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.OriginalPrice, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, wrapError("ListProducts", "scan", err)
		}
		products = append(products, p)
	}

	// Get total count
	var total int
	var countQuery string

	if params.Status == "all" {
		countQuery = "SELECT COUNT(*) FROM products"
		s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	} else {
		countQuery = "SELECT COUNT(*) FROM products WHERE status = $1"
		s.db.QueryRowContext(ctx, countQuery, params.Status).Scan(&total)
	}

	result := &ListProductsResult{
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
		Data:    products,
	}

	// Cache result
	if resultJSON, err := json.Marshal(result); err == nil {
		_ = cache.Set(ctx, cacheKey, string(resultJSON), cache.ProductListTTL)
	}

	s.log.Printf("Listed products: page=%d, per_page=%d, status=%s, total=%d", params.Page, params.PerPage, params.Status, total)
	return result, nil
}

// GetProductByID retrieves a single product by ID with caching
func (s *service) GetProductByID(ctx context.Context, productID int) (*Product, error) {
	// Try cache first
	cacheKey := cache.ProductKey(productID)
	if cachedData, err := cache.Get(ctx, cacheKey); err == nil {
		var product Product
		if err := json.Unmarshal([]byte(cachedData), &product); err == nil {
			return &product, nil
		}
	}

	// Query database
	var product Product
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, merchant_id, name, description, price, COALESCE(original_price, 0), stock, status, created_at, updated_at FROM products WHERE id = $1",
		productID,
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, wrapError("GetProductByID", "query", err)
	}

	// Cache result
	if productJSON, err := json.Marshal(product); err == nil {
		_ = cache.Set(ctx, cacheKey, string(productJSON), cache.ProductCacheTTL)
	}

	return &product, nil
}

// SearchProducts searches for products with caching
func (s *service) SearchProducts(ctx context.Context, params *SearchProductsParams) (*ListProductsResult, error) {
	// Validate search query
	if params.Query == "" {
		return nil, ErrInvalidSearchQuery
	}

	// Validate pagination parameters
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	// Try cache first
	cacheKey := cache.ProductSearchKey(params.Query, params.Page, params.PerPage)
	if cachedData, err := cache.Get(ctx, cacheKey); err == nil {
		var result ListProductsResult
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			return &result, nil
		}
	}

	// Query database
	offset := (params.Page - 1) * params.PerPage
	searchQuery := "%" + params.Query + "%"

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT id, merchant_id, name, description, price, COALESCE(original_price, 0), stock, status, created_at, updated_at FROM products WHERE status = 'active' AND (name ILIKE $1 OR description ILIKE $1) ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		searchQuery, params.PerPage, offset,
	)

	if err != nil {
		return nil, wrapError("SearchProducts", "query", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Name, &p.Description, &p.Price, &p.OriginalPrice, &p.Stock, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, wrapError("SearchProducts", "scan", err)
		}
		products = append(products, p)
	}

	// Get total count
	var total int
	s.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM products WHERE status = 'active' AND (name ILIKE $1 OR description ILIKE $1)",
		searchQuery,
	).Scan(&total)

	result := &ListProductsResult{
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
		Data:    products,
	}

	// Cache result
	if resultJSON, err := json.Marshal(result); err == nil {
		_ = cache.Set(ctx, cacheKey, string(resultJSON), cache.SearchResultsTTL)
	}

	s.log.Printf("Searched products: query=%s, page=%d, per_page=%d, total=%d", params.Query, params.Page, params.PerPage, total)
	return result, nil
}

// CreateProduct creates a new product (merchant only)
func (s *service) CreateProduct(ctx context.Context, merchantID int, req *CreateProductRequest) (*Product, error) {
	// Validate input
	if req.Name == "" {
		return nil, ErrInvalidProductName
	}
	if req.Price <= 0 {
		return nil, ErrInvalidPrice
	}
	if req.Stock < 0 {
		return nil, ErrInvalidStock
	}

	// Create product
	var product Product
	err := s.db.QueryRowContext(
		ctx,
		"INSERT INTO products (merchant_id, name, description, price, original_price, stock, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, merchant_id, name, description, price, COALESCE(original_price, 0), stock, status, created_at, updated_at",
		merchantID, req.Name, req.Description, req.Price, req.OriginalPrice, req.Stock, "active",
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return nil, wrapError("CreateProduct", "insert", err)
	}

	// Invalidate list cache
	_ = cache.InvalidatePatterns(ctx, "products:list:*")
	_ = cache.InvalidatePatterns(ctx, "products:search:*")

	s.log.Printf("Product created: id=%d, merchant_id=%d, name=%s", product.ID, merchantID, req.Name)
	return &product, nil
}

// UpdateProduct updates a product (merchant only)
func (s *service) UpdateProduct(ctx context.Context, merchantID int, productID int, req *UpdateProductRequest) (*Product, error) {
	// Verify ownership
	var ownerID int
	err := s.db.QueryRowContext(ctx, "SELECT merchant_id FROM products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, wrapError("UpdateProduct", "checkOwnership", err)
	}

	if ownerID != merchantID {
		return nil, ErrNotProductOwner
	}

	// Build update query (only update non-zero fields)
	var product Product
	err = s.db.QueryRowContext(
		ctx,
		"UPDATE products SET name = COALESCE(NULLIF($1, ''), name), description = COALESCE(NULLIF($2, ''), description), price = CASE WHEN $3 > 0 THEN $3 ELSE price END, original_price = CASE WHEN $4 > 0 THEN $4 ELSE original_price END, stock = CASE WHEN $5 >= 0 THEN $5 ELSE stock END, status = COALESCE(NULLIF($6, ''), status), updated_at = NOW() WHERE id = $7 RETURNING id, merchant_id, name, description, price, COALESCE(original_price, 0), stock, status, created_at, updated_at",
		req.Name, req.Description, req.Price, req.OriginalPrice, req.Stock, req.Status, productID,
	).Scan(&product.ID, &product.MerchantID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.Stock, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return nil, wrapError("UpdateProduct", "update", err)
	}

	// Invalidate caches
	_ = cache.Delete(ctx, cache.ProductKey(productID))
	_ = cache.InvalidatePatterns(ctx, "products:list:*")
	_ = cache.InvalidatePatterns(ctx, "products:search:*")

	s.log.Printf("Product updated: id=%d, merchant_id=%d", productID, merchantID)
	return &product, nil
}

// DeleteProduct deletes a product (merchant only)
func (s *service) DeleteProduct(ctx context.Context, merchantID int, productID int) error {
	// Verify ownership
	var ownerID int
	err := s.db.QueryRowContext(ctx, "SELECT merchant_id FROM products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrProductNotFound
		}
		return wrapError("DeleteProduct", "checkOwnership", err)
	}

	if ownerID != merchantID {
		return ErrNotProductOwner
	}

	// Delete product
	result, err := s.db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		return wrapError("DeleteProduct", "delete", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return wrapError("DeleteProduct", "rowsAffected", err)
	}

	if affected == 0 {
		return ErrProductNotFound
	}

	// Invalidate caches
	_ = cache.Delete(ctx, cache.ProductKey(productID))
	_ = cache.InvalidatePatterns(ctx, "products:list:*")
	_ = cache.InvalidatePatterns(ctx, "products:search:*")

	s.log.Printf("Product deleted: id=%d, merchant_id=%d", productID, merchantID)
	return nil
}
