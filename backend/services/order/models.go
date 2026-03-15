package order

import "time"

// CreateOrderRequest represents order creation request
type CreateOrderRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	GroupID   int `json:"group_id"`
	Quantity  int `json:"quantity" binding:"required,gt=0"`
}

// ListOrdersParams represents list parameters
type ListOrdersParams struct {
	Page    int
	PerPage int
	Status  string
}

// ListOrdersResult represents paginated order list result
type ListOrdersResult struct {
	Total   int     `json:"total"`
	Page    int     `json:"page"`
	PerPage int     `json:"per_page"`
	Data    []Order `json:"data"`
}

// Order represents a user's order
type Order struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	ProductID  int       `json:"product_id"`
	GroupID    int       `json:"group_id"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// OrderDetail represents detailed order information with product details
type OrderDetail struct {
	*Order
	ProductName  string `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
}
