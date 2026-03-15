package payment

import "time"

// InitiatePaymentRequest represents a payment initiation request
type InitiatePaymentRequest struct {
	OrderID       int    `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=alipay wechat"`
}

// Payment represents a payment transaction
type Payment struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	OrderID       int       `json:"order_id"`
	Amount        float64   `json:"amount"`
	Method        string    `json:"method"`       // alipay, wechat
	Status        string    `json:"status"`       // pending, success, failed, refunded
	TransactionID *string   `json:"transaction_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AlipayCallback represents an Alipay webhook callback
type AlipayCallback struct {
	OutTradeNo    string  `json:"out_trade_no"`
	TradeNo       string  `json:"trade_no"`
	TotalAmount   float64 `json:"total_amount"`
	TradeStatus   string  `json:"trade_status"`
	Timestamp     string  `json:"timestamp"`
	Sign          string  `json:"sign"`
}

// WechatCallback represents a WeChat Pay webhook callback
type WechatCallback struct {
	OutTradeNo    string  `json:"out_trade_no"`
	TransactionID string  `json:"transaction_id"`
	TotalFee      int     `json:"total_fee"` // in cents
	ResultCode    string  `json:"result_code"`
	Sign          string  `json:"sign"`
}

// MerchantRevenue represents merchant revenue information
type MerchantRevenue struct {
	MerchantID           int       `json:"merchant_id"`
	Period               string    `json:"period"`
	TotalSales           float64   `json:"total_sales"`
	CommissionRate       float64   `json:"commission_rate"`
	PlatformCommission   float64   `json:"platform_commission"`
	APICallCost          float64   `json:"api_call_cost"`
	MerchantEarnings     float64   `json:"merchant_earnings"`
	TransactionCount     int       `json:"transaction_count"`
	AverageOrderValue    float64   `json:"average_order_value"`
}

// RefundRequest represents a refund request
type RefundRequest struct {
	PaymentID int    `json:"payment_id" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
}

// PaymentListResult represents paginated payment list result
type PaymentListResult struct {
	Total   int        `json:"total"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
	Data    []Payment  `json:"data"`
}

// ListPaymentsParams represents payment list parameters
type ListPaymentsParams struct {
	Page    int
	PerPage int
	Status  string
	Method  string
}
