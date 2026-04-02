package models

import "time"

// User represents a user in the system
type User struct {
	ID             int       `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Password       string    `json:"-"`
	Role           string    `json:"role"` // user, merchant, admin
	ReferralCode   string    `json:"referral_code,omitempty"`
	ReferredBy     int       `json:"referred_by,omitempty"`
	TotalReferrals int       `json:"total_referrals,omitempty"`
	TotalRewards   float64   `json:"total_rewards,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Product represents a token product
type Product struct {
	ID            int       `json:"id"`
	MerchantID    int       `json:"merchant_id"`
	SpuID         int       `json:"spu_id,omitempty"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	OriginalPrice float64   `json:"original_price,omitempty"`
	Stock         int       `json:"stock"`
	SoldCount     int       `json:"sold_count"`
	Category      string    `json:"category,omitempty"`
	Status        string    `json:"status"` // active, inactive, archived
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Order represents a user's order
type Order struct {
	ID         int         `json:"id"`
	UserID     int         `json:"user_id"`
	ProductID  *int        `json:"product_id,omitempty"` // NULL for SKU-only orders (migration 020)
	SKUID      int         `json:"sku_id,omitempty"`
	SPUID      int         `json:"spu_id,omitempty"`
	GroupID    interface{} `json:"group_id"` // Can be NULL
	Quantity   int         `json:"quantity"`
	UnitPrice  float64     `json:"unit_price"`
	TotalPrice float64     `json:"total_price"`
	Status     string      `json:"status"` // pending, paid, completed, failed
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// Group represents a group purchase
type Group struct {
	ID           int       `json:"id"`
	ProductID    *int      `json:"product_id,omitempty"` // legacy products FK; NULL for SKU-only groups (migration 020)
	SKUID        int       `json:"sku_id,omitempty"`
	SPUID        int       `json:"spu_id,omitempty"`
	CreatorID    int       `json:"creator_id"`
	TargetCount  int       `json:"target_count"`
	CurrentCount int       `json:"current_count"`
	Status       string    `json:"status"` // active, completed, failed
	Deadline     time.Time `json:"deadline"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Token represents user token balance
type Token struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Payment represents a payment transaction
type Payment struct {
	ID            int       `json:"id"`
	OrderID       int       `json:"order_id"`
	UserID        int       `json:"user_id"`
	Amount        float64   `json:"amount"`
	PayMethod     string    `json:"pay_method"`
	OutTradeNo    string    `json:"out_trade_no"`
	TransactionID string    `json:"transaction_id,omitempty"`
	Status        string    `json:"status"`
	PaidAt        time.Time `json:"paid_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// APIKey represents an API key for B2B users
type APIKey struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Key       string    `json:"-"`
	KeyHash   string    `json:"-"`
	Name      string    `json:"name"`
	Status    string    `json:"status"` // active, inactive
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReferralCode represents a user's referral code
type ReferralCode struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Referral represents a referral relationship
type Referral struct {
	ID         int       `json:"id"`
	ReferrerID int       `json:"referrer_id"`
	RefereeID  int       `json:"referee_id"`
	CodeUsed   string    `json:"code_used"`
	Status     string    `json:"status"` // active, canceled
	CreatedAt  time.Time `json:"created_at"`
}

// ReferralReward represents a referral reward
type ReferralReward struct {
	ID         int       `json:"id"`
	ReferrerID int       `json:"referrer_id"`
	RefereeID  int       `json:"referee_id"`
	OrderID    int       `json:"order_id,omitempty"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"` // pending, paid, canceled
	CreatedAt  time.Time `json:"created_at"`
	PaidAt     time.Time `json:"paid_at,omitempty"`
}

// Merchant represents a merchant/seller
type Merchant struct {
	ID                 int        `json:"id"`
	UserID             int        `json:"user_id"`
	CompanyName        string     `json:"company_name"`
	BusinessLicense    *string    `json:"business_license,omitempty"`
	BusinessLicenseURL *string    `json:"business_license_url,omitempty"`
	IDCardFrontURL     *string    `json:"id_card_front_url,omitempty"`
	IDCardBackURL      *string    `json:"id_card_back_url,omitempty"`
	Attachments        *string    `json:"attachments,omitempty"` // JSON array of file URLs
	ContactName        *string    `json:"contact_name,omitempty"`
	ContactPhone       *string    `json:"contact_phone,omitempty"`
	ContactEmail       *string    `json:"contact_email,omitempty"`
	Address            *string    `json:"address,omitempty"`
	Description        *string    `json:"description,omitempty"`
	LogoURL            *string    `json:"logo_url,omitempty"`
	BusinessCategory   *string    `json:"business_category,omitempty"` // 经营类目
	AdminNotes         *string    `json:"admin_notes,omitempty"`       // 管理员内部备注
	Status             string     `json:"status"`                      // pending, reviewing, active, suspended, rejected
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty"`
	ReviewNote         *string    `json:"review_note,omitempty"`
	RejectionReason    *string    `json:"rejection_reason,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// MerchantAuditLog is an admin action record for merchant review and metadata changes.
type MerchantAuditLog struct {
	ID                  int       `json:"id"`
	MerchantID          int       `json:"merchant_id"`
	AdminUserID         *int      `json:"admin_user_id,omitempty"`
	AdminEmail          *string   `json:"admin_email,omitempty"`
	Action              string    `json:"action"`
	CompanyNameSnapshot *string   `json:"company_name_snapshot,omitempty"`
	Reason              *string   `json:"reason,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

// MerchantAPIKey represents a merchant's API key for token托管
type MerchantAPIKey struct {
	ID                 int       `json:"id"`
	MerchantID         int       `json:"merchant_id"`
	Name               string    `json:"name"`
	Provider           string    `json:"provider"` // openai, anthropic, etc.
	APIKeyEncrypted    string    `json:"-"`
	APISecretEncrypted string    `json:"-"`
	QuotaLimit         float64   `json:"quota_limit"`
	QuotaUsed          float64   `json:"quota_used"`
	Status             string    `json:"status"` // active, inactive
	LastUsedAt         time.Time `json:"last_used_at,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// MerchantSettlement represents a merchant's settlement record
type MerchantSettlement struct {
	ID               int       `json:"id"`
	MerchantID       int       `json:"merchant_id"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	TotalSales       float64   `json:"total_sales"`
	PlatformFee      float64   `json:"platform_fee"`
	SettlementAmount float64   `json:"settlement_amount"`
	Status           string    `json:"status"` // pending, processing, completed
	SettledAt        time.Time `json:"settled_at,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// MerchantStats represents daily merchant statistics
type MerchantStats struct {
	ID              int       `json:"id"`
	MerchantID      int       `json:"merchant_id"`
	StatDate        time.Time `json:"stat_date"`
	TotalOrders     int       `json:"total_orders"`
	TotalSales      float64   `json:"total_sales"`
	TotalTokensSold float64   `json:"total_tokens_sold"`
	NewCustomers    int       `json:"new_customers"`
	CreatedAt       time.Time `json:"created_at"`
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	SKUID     int       `json:"sku_id,omitempty"`
	SPUID     int       `json:"spu_id,omitempty"`
	GroupID   int       `json:"group_id,omitempty"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartResponse represents the cart response with product details
type CartResponse struct {
	ID       int     `json:"id"`
	SKUID    int     `json:"sku_id"`
	Product  Product `json:"product"`
	GroupID  int     `json:"group_id,omitempty"`
	Quantity int     `json:"quantity"`
}

// CartSummary represents the cart summary
type CartSummary struct {
	Items      []CartResponse `json:"items"`
	TotalItems int            `json:"total_items"`
	TotalPrice float64        `json:"total_price"`
}

// Favorite represents a user's favorite product
type Favorite struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

// FavoriteResponse represents a favorite with product details
type FavoriteResponse struct {
	ID        int     `json:"id"`
	SKUID     int     `json:"sku_id"`
	Product   Product `json:"product"`
	CreatedAt string  `json:"created_at"`
}

// BrowseHistory represents a user's browse history
type BrowseHistory struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	ViewCount int       `json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BrowseHistoryResponse represents a browse history item with product details
type BrowseHistoryResponse struct {
	ID        int     `json:"id"`
	SKUID     int     `json:"sku_id"`
	Product   Product `json:"product"`
	ViewCount int     `json:"view_count"`
	ViewedAt  string  `json:"viewed_at"`
}
