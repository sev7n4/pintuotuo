package product

import "time"

// CreateProductRequest represents product creation request
type CreateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	OriginalPrice float64 `json:"original_price"`
	Stock         int     `json:"stock" binding:"required,gte=0"`
}

// UpdateProductRequest represents product update request
type UpdateProductRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	Stock         int     `json:"stock"`
	Status        string  `json:"status"`
}

// ListProductsParams represents list parameters
type ListProductsParams struct {
	Page    int
	PerPage int
	Status  string
}

// SearchProductsParams represents search parameters
type SearchProductsParams struct {
	Query   string
	Page    int
	PerPage int
}

// ListProductsResult represents paginated product list result
type ListProductsResult struct {
	Total   int       `json:"total"`
	Page    int       `json:"page"`
	PerPage int       `json:"per_page"`
	Data    []Product `json:"data"`
}

// Product represents a token product
type Product struct {
	ID            int       `json:"id"`
	MerchantID    int       `json:"merchant_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	OriginalPrice float64   `json:"original_price"`
	Stock         int       `json:"stock"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
