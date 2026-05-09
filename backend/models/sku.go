package models

import (
	"time"
)

type ProviderModel struct {
	ID                   int        `json:"id"`
	ProviderCode         string     `json:"provider_code"`
	ModelID              string     `json:"model_id"`
	DisplayName          string     `json:"display_name,omitempty"`
	ReferenceInputPrice  *float64   `json:"reference_input_price,omitempty"`
	ReferenceOutputPrice *float64   `json:"reference_output_price,omitempty"`
	ReferenceCurrency    string     `json:"reference_currency,omitempty"`
	IsActive             bool       `json:"is_active"`
	SyncedAt             *time.Time `json:"synced_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ModelProvider struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	APIBaseURL  string `json:"api_base_url,omitempty"`
	APIFormat   string `json:"api_format"`
	BillingType string `json:"billing_type,omitempty"`
	// LiteLLM 网关：与 litellm-catalog-sync / 部署侧 os.environ 对齐（DB 为主 SSOT）
	LitellmModelTemplate    string         `json:"litellm_model_template,omitempty"`
	LitellmGatewayAPIKeyEnv string         `json:"litellm_gateway_api_key_env,omitempty"`
	LitellmGatewayAPIBase   string         `json:"litellm_gateway_api_base,omitempty"`
	CompatPrefixes          []string       `json:"compat_prefixes,omitempty"`
	SegmentConfig           map[string]any `json:"segment_config,omitempty"`
	CacheEnabled            bool           `json:"cache_enabled"`
	CacheDiscount           float64        `json:"cache_discount_rate,omitempty"`
	// 统一配置化路由系统字段
	ProviderRegion string                 `json:"provider_region,omitempty"` // domestic, overseas
	RouteStrategy  map[string]interface{} `json:"route_strategy,omitempty"`  // JSONB: route strategy config
	Endpoints      map[string]interface{} `json:"endpoints,omitempty"`       // JSONB: endpoints config
	Status         string                 `json:"status"`
	SortOrder      int                    `json:"sort_order"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type SPU struct {
	ID      int    `json:"id"`
	SPUCode string `json:"spu_code"`
	Name    string `json:"name"`

	ModelProvider       string  `json:"model_provider"`
	ProviderModelID     string  `json:"provider_model_id,omitempty"`
	ProviderAPIEndpoint string  `json:"provider_api_endpoint,omitempty"`
	ProviderAuthType    string  `json:"provider_auth_type,omitempty"`
	ProviderBillingType string  `json:"provider_billing_type,omitempty"`
	ProviderInputRate   float64 `json:"provider_input_rate,omitempty"`
	ProviderOutputRate  float64 `json:"provider_output_rate,omitempty"`

	ModelName    string `json:"model_name"`
	ModelVersion string `json:"model_version,omitempty"`
	ModelTier    string `json:"model_tier"`

	ContextWindow      int      `json:"context_window,omitempty"`
	MaxOutputTokens    int      `json:"max_output_tokens,omitempty"`
	SupportedFunctions []string `json:"supported_functions,omitempty"`

	BaseComputePoints  float64 `json:"base_compute_points"`
	BillingCoefficient float64 `json:"billing_coefficient"`

	Description  string   `json:"description,omitempty"`
	Features     []string `json:"features,omitempty"`
	ThumbnailURL string   `json:"thumbnail_url,omitempty"`

	InputLengthRanges []InputRange   `json:"input_length_ranges,omitempty"`
	BillingAdapter    BillingAdapter `json:"billing_adapter,omitempty"`
	RoutingRules      RoutingRules   `json:"routing_rules,omitempty"`
	BatchInference    BatchConfig    `json:"batch_inference,omitempty"`

	Status          string  `json:"status"`
	SortOrder       int     `json:"sort_order"`
	TotalSalesCount int64   `json:"total_sales_count"`
	AverageRating   float64 `json:"average_rating,omitempty"`
	// SkuCount / ActiveSkuCount：管理端列表与详情 JOIN 统计，其它接口可能为 0
	SkuCount       int       `json:"sku_count"`
	ActiveSkuCount int       `json:"active_sku_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type InputRange struct {
	MinTokens int     `json:"min_tokens"`
	MaxTokens int     `json:"max_tokens"`
	Label     string  `json:"label"`
	Surcharge float64 `json:"surcharge,omitempty"`
}

type BillingAdapter struct {
	Type          string        `json:"type"`
	SegmentConfig []SegmentRule `json:"segment_config,omitempty"`
	CacheEnabled  bool          `json:"cache_enabled"`
	CacheDiscount float64       `json:"cache_discount_rate,omitempty"`
}

type SegmentRule struct {
	InputRange string  `json:"input_range"`
	Multiplier float64 `json:"multiplier"`
}

type RoutingRules struct {
	AutoRoute    bool              `json:"auto_route"`
	DefaultRange string            `json:"default_range,omitempty"`
	RangeMapping map[string]string `json:"range_mapping,omitempty"`
}

type BatchConfig struct {
	Enabled      bool    `json:"enabled"`
	DiscountRate float64 `json:"discount_rate"`
	AsyncOnly    bool    `json:"async_only"`
}

type SKU struct {
	ID                 int       `json:"id"`
	SPUID              int       `json:"spu_id"`
	SPUCode            string    `json:"spu_code,omitempty"`
	SPUName            string    `json:"spu_name,omitempty"`
	SKUCode            string    `json:"sku_code"`
	MerchantID         *int      `json:"merchant_id,omitempty"`
	SKUType            string    `json:"sku_type"`
	EndpointType       string    `json:"endpoint_type,omitempty"`
	TokenAmount        int64     `json:"token_amount,omitempty"`
	ComputePoints      float64   `json:"compute_points,omitempty"`
	SubscriptionPeriod string    `json:"subscription_period,omitempty"`
	IsUnlimited        bool      `json:"is_unlimited"`
	FairUseLimit       int64     `json:"fair_use_limit,omitempty"`
	TPMLimit           int       `json:"tpm_limit,omitempty"`
	RPMLimit           int       `json:"rpm_limit,omitempty"`
	ConcurrentReqs     int       `json:"concurrent_requests,omitempty"`
	ValidDays          int       `json:"valid_days"`
	RetailPrice        float64   `json:"retail_price"`
	WholesalePrice     float64   `json:"wholesale_price,omitempty"`
	OriginalPrice      float64   `json:"original_price,omitempty"`
	Stock              int       `json:"stock"`
	DailyLimit         int       `json:"daily_limit,omitempty"`
	GroupEnabled       bool      `json:"group_enabled"`
	MinGroupSize       int       `json:"min_group_size"`
	MaxGroupSize       int       `json:"max_group_size"`
	GroupDiscountRate  float64   `json:"group_discount_rate,omitempty"`
	CostInputRate      float64   `json:"cost_input_rate,omitempty"`
	CostOutputRate     float64   `json:"cost_output_rate,omitempty"`
	InheritSPUCost     bool      `json:"inherit_spu_cost"`
	IsTrial            bool      `json:"is_trial"`
	TrialDurationDays  int       `json:"trial_duration_days,omitempty"`
	Status             string    `json:"status"`
	IsPromoted         bool      `json:"is_promoted"`
	SalesCount         int64     `json:"sales_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type SKUWithSPU struct {
	SKU
	SPUName            string   `json:"spu_name"`
	SpuStatus          string   `json:"spu_status"`
	Sellable           bool     `json:"sellable"`
	ModelProvider      string   `json:"model_provider"`
	ModelName          string   `json:"model_name"`
	ModelTier          string   `json:"model_tier"`
	SPUTotalSalesCount int64    `json:"spu_total_sales_count,omitempty"`
	SPUAverageRating   *float64 `json:"spu_average_rating,omitempty"`
}

type ComputePointAccount struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Balance      float64   `json:"balance"`
	TotalEarned  float64   `json:"total_earned"`
	TotalUsed    float64   `json:"total_used"`
	TotalExpired float64   `json:"total_expired"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ComputePointTransaction struct {
	ID           int            `json:"id"`
	UserID       int            `json:"user_id"`
	Type         string         `json:"type"`
	Amount       float64        `json:"amount"`
	BalanceAfter float64        `json:"balance_after"`
	OrderID      *int           `json:"order_id,omitempty"`
	SKUID        *int           `json:"sku_id,omitempty"`
	Description  string         `json:"description,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type UserSubscription struct {
	ID                  int        `json:"id"`
	UserID              int        `json:"user_id"`
	SKUID               int        `json:"sku_id"`
	StartDate           time.Time  `json:"start_date"`
	EndDate             time.Time  `json:"end_date"`
	UsedTokens          int64      `json:"used_tokens"`
	UsedComputePoints   float64    `json:"used_compute_points"`
	Status              string     `json:"status"`
	AutoRenew           bool       `json:"auto_renew"`
	PricingVersionID    *int       `json:"pricing_version_id,omitempty"`
	EntitlementAnchorAt *time.Time `json:"entitlement_anchor_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type UserSubscriptionWithSKU struct {
	UserSubscription
	SKUCode     string  `json:"sku_code"`
	SPUName     string  `json:"spu_name"`
	RetailPrice float64 `json:"retail_price"`
}

type SPUCreateRequest struct {
	SPUCode             string   `json:"spu_code" binding:"required"`
	Name                string   `json:"name" binding:"required"`
	ModelProvider       string   `json:"model_provider" binding:"required"`
	ModelName           string   `json:"model_name" binding:"required"`
	ModelVersion        string   `json:"model_version"`
	ModelTier           string   `json:"model_tier" binding:"required"`
	ContextWindow       int      `json:"context_window"`
	MaxOutputTokens     int      `json:"max_output_tokens"`
	SupportedFunctions  []string `json:"supported_functions"`
	BaseComputePoints   float64  `json:"base_compute_points"`
	Description         string   `json:"description"`
	Features            []string `json:"features"`
	ThumbnailURL        string   `json:"thumbnail_url"`
	ProviderInputRate   float64  `json:"provider_input_rate"`
	ProviderOutputRate  float64  `json:"provider_output_rate"`
	SyncBaselinePricing *bool    `json:"sync_baseline_pricing,omitempty"`
	SyncPricingVersions []int    `json:"sync_pricing_versions,omitempty"`
	Status              string   `json:"status"`
	SortOrder           int      `json:"sort_order"`
}

type SPUUpdateRequest struct {
	Name                *string  `json:"name"`
	ModelProvider       *string  `json:"model_provider"`
	ModelName           *string  `json:"model_name"`
	ModelVersion        *string  `json:"model_version"`
	ModelTier           *string  `json:"model_tier"`
	ContextWindow       *int     `json:"context_window"`
	MaxOutputTokens     *int     `json:"max_output_tokens"`
	SupportedFunctions  []string `json:"supported_functions"`
	BaseComputePoints   *float64 `json:"base_compute_points"`
	Description         *string  `json:"description"`
	Features            []string `json:"features"`
	ThumbnailURL        *string  `json:"thumbnail_url"`
	ProviderInputRate   *float64 `json:"provider_input_rate"`
	ProviderOutputRate  *float64 `json:"provider_output_rate"`
	SyncBaselinePricing *bool    `json:"sync_baseline_pricing,omitempty"`
	SyncPricingVersions []int    `json:"sync_pricing_versions,omitempty"`
	Status              *string  `json:"status"`
	SortOrder           *int     `json:"sort_order"`
}

type SKUCreateRequest struct {
	SPUID              int     `json:"spu_id" binding:"required"`
	SKUCode            string  `json:"sku_code" binding:"required"`
	SKUType            string  `json:"sku_type" binding:"required"`
	EndpointType       string  `json:"endpoint_type"`
	TokenAmount        int64   `json:"token_amount"`
	ComputePoints      float64 `json:"compute_points"`
	SubscriptionPeriod string  `json:"subscription_period"`
	IsUnlimited        bool    `json:"is_unlimited"`
	FairUseLimit       int64   `json:"fair_use_limit"`
	TPMLimit           int     `json:"tpm_limit"`
	RPMLimit           int     `json:"rpm_limit"`
	ConcurrentReqs     int     `json:"concurrent_requests"`
	ValidDays          int     `json:"valid_days"`
	RetailPrice        float64 `json:"retail_price" binding:"required"`
	WholesalePrice     float64 `json:"wholesale_price"`
	OriginalPrice      float64 `json:"original_price"`
	Stock              FlexInt `json:"stock"`
	DailyLimit         FlexInt `json:"daily_limit"`
	GroupEnabled       bool    `json:"group_enabled"`
	MinGroupSize       FlexInt `json:"min_group_size"`
	MaxGroupSize       FlexInt `json:"max_group_size"`
	GroupDiscountRate  float64 `json:"group_discount_rate"`
	CostInputRate      float64 `json:"cost_input_rate"`
	CostOutputRate     float64 `json:"cost_output_rate"`
	InheritSPUCost     bool    `json:"inherit_spu_cost"`
	IsTrial            bool    `json:"is_trial"`
	TrialDurationDays  int     `json:"trial_duration_days"`
	Status             string  `json:"status"`
	IsPromoted         bool    `json:"is_promoted"`
}

type SKUUpdateRequest struct {
	Name               string  `json:"name"`
	EndpointType       string  `json:"endpoint_type"`
	TokenAmount        int64   `json:"token_amount"`
	ComputePoints      float64 `json:"compute_points"`
	SubscriptionPeriod string  `json:"subscription_period"`
	IsUnlimited        bool    `json:"is_unlimited"`
	FairUseLimit       int64   `json:"fair_use_limit"`
	TPMLimit           int     `json:"tpm_limit"`
	RPMLimit           int     `json:"rpm_limit"`
	ConcurrentReqs     int     `json:"concurrent_requests"`
	RetailPrice        float64 `json:"retail_price"`
	WholesalePrice     float64 `json:"wholesale_price"`
	OriginalPrice      float64 `json:"original_price"`
	Stock              FlexInt `json:"stock"`
	DailyLimit         FlexInt `json:"daily_limit"`
	GroupEnabled       bool    `json:"group_enabled"`
	MinGroupSize       FlexInt `json:"min_group_size"`
	MaxGroupSize       FlexInt `json:"max_group_size"`
	GroupDiscountRate  float64 `json:"group_discount_rate"`
	CostInputRate      float64 `json:"cost_input_rate"`
	CostOutputRate     float64 `json:"cost_output_rate"`
	InheritSPUCost     *bool   `json:"inherit_spu_cost"`
	ValidDays          FlexInt `json:"valid_days"`
	Status             string  `json:"status"`
	IsPromoted         bool    `json:"is_promoted"`
}

type ComputePointBalanceResponse struct {
	Balance      float64 `json:"balance"`
	TotalEarned  float64 `json:"total_earned"`
	TotalUsed    float64 `json:"total_used"`
	TotalExpired float64 `json:"total_expired"`
}
