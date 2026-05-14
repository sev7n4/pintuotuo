package models

import "time"

type MerchantSKU struct {
	ID                int       `json:"id"`
	MerchantID        int       `json:"merchant_id"`
	SKUID             int       `json:"sku_id"`
	APIKeyID          *int      `json:"api_key_id,omitempty"`
	AnthropicAPIKeyID *int      `json:"anthropic_api_key_id,omitempty"`
	Status            string    `json:"status"`
	SalesCount        int64     `json:"sales_count"`
	TotalSalesAmount  float64   `json:"total_sales_amount"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type MerchantSKUDetail struct {
	MerchantSKU
	SKUCode                 string  `json:"sku_code"`
	SKUType                 string  `json:"sku_type"`
	TokenAmount             int64   `json:"token_amount,omitempty"`
	ComputePoints           float64 `json:"compute_points,omitempty"`
	RetailPrice             float64 `json:"retail_price"`
	OriginalPrice           float64 `json:"original_price,omitempty"`
	ValidDays               int     `json:"valid_days"`
	GroupEnabled            bool    `json:"group_enabled"`
	GroupDiscountRate       float64 `json:"group_discount_rate,omitempty"`
	SPUName                 string  `json:"spu_name"`
	ModelProvider           string  `json:"model_provider"`
	ModelName               string  `json:"model_name"`
	ModelTier               string  `json:"model_tier"`
	APIKeyName              string  `json:"api_key_name,omitempty"`
	APIKeyProvider          string  `json:"api_key_provider,omitempty"`
	AnthropicAPIKeyName     string  `json:"anthropic_api_key_name,omitempty"`
	AnthropicAPIKeyProvider string  `json:"anthropic_api_key_provider,omitempty"`
	CostInputRate           float64 `json:"cost_input_rate"`
	CostOutputRate          float64 `json:"cost_output_rate"`
	ProfitMargin            float64 `json:"profit_margin"`
	CustomPricing           bool    `json:"custom_pricing_enabled"`
	SPUInputRate            float64 `json:"spu_input_rate,omitempty"`
	SPUOutputRate           float64 `json:"spu_output_rate,omitempty"`
}

type MerchantSKUCreateRequest struct {
	SKUID                int      `json:"sku_id" binding:"required"`
	APIKeyID             *int     `json:"api_key_id"`
	AnthropicAPIKeyID    *int     `json:"anthropic_api_key_id"`
	CustomPricingEnabled bool     `json:"custom_pricing_enabled"`
	CostInputRate        *float64 `json:"cost_input_rate"`
	CostOutputRate       *float64 `json:"cost_output_rate"`
	ProfitMargin         *float64 `json:"profit_margin"`
}

type MerchantSKUUpdateRequest struct {
	APIKeyID             *int     `json:"api_key_id"`
	AnthropicAPIKeyID    *int     `json:"anthropic_api_key_id"`
	Status               string   `json:"status"`
	CustomPricingEnabled *bool    `json:"custom_pricing_enabled"`
	CostInputRate        *float64 `json:"cost_input_rate"`
	CostOutputRate       *float64 `json:"cost_output_rate"`
	ProfitMargin         *float64 `json:"profit_margin"`
}

type AvailableSKU struct {
	ID                int     `json:"id"`
	SKUCode           string  `json:"sku_code"`
	SKUType           string  `json:"sku_type"`
	TokenAmount       int64   `json:"token_amount,omitempty"`
	ComputePoints     float64 `json:"compute_points,omitempty"`
	RetailPrice       float64 `json:"retail_price"`
	OriginalPrice     float64 `json:"original_price,omitempty"`
	ValidDays         int     `json:"valid_days"`
	GroupEnabled      bool    `json:"group_enabled"`
	GroupDiscountRate float64 `json:"group_discount_rate,omitempty"`
	SPUID             int     `json:"spu_id"`
	SPUName           string  `json:"spu_name"`
	ModelProvider     string  `json:"model_provider"`
	ModelName         string  `json:"model_name"`
	ModelTier         string  `json:"model_tier"`
	SPUInputRate      float64 `json:"spu_input_rate,omitempty"`
	SPUOutputRate     float64 `json:"spu_output_rate,omitempty"`
	IsSelected        bool    `json:"is_selected"`
}
