// User related types
export interface User {
  id: number;
  email: string;
  name: string;
  /** 绑定手机号（若有） */
  phone?: string;
  role: 'user' | 'merchant' | 'admin';
  created_at: string;
  updated_at: string;
}

/** GET /users/me/identities */
export interface UserIdentity {
  provider: string;
  external_id: string;
  display_name?: string;
}

// Product related types
export interface Product {
  id: number;
  merchant_id: number;
  /** SPU id，用于加载同系列 SKU */
  spu_id?: number;
  name: string;
  description: string;
  price: number;
  original_price?: number;
  stock: number;
  sold_count?: number;
  category?: string;
  /** 列表/卡片主图（若后端返回） */
  image_url?: string;
  thumbnail_url?: string;
  status: 'active' | 'inactive' | 'archived';
  created_at: string;
  updated_at: string;
  token_count?: number;
  token_type?: string;
  models?: string[];
  validity_period?: string;
  context_length?: string;
  rating?: number;
  review_count?: number;
  group_prices?: GroupPrice[];
}

export interface GroupPrice {
  min_members: number;
  price_per_person: number;
  discount_percent: number;
}

export interface ProductReview {
  id: number;
  user_id: number;
  user_name: string;
  user_avatar?: string;
  rating: number;
  content: string;
  created_at: string;
}

// Category type（卖场/首页辅料：按 model_tier 聚合）
export interface Category {
  name: string;
  count: number;
}

/** 首页主分类：使用场景（与 usage_scenarios + 在售 SKU 统计一致） */
export interface ScenarioCategoryItem {
  code: string;
  name: string;
  count: number;
}

// Banner type
export interface Banner {
  id: number;
  title: string;
  image: string;
  link: string;
}

// Home data type
export interface HomeData {
  banners: Banner[];
  hot: Product[];
  new: Product[];
  categories: Category[];
  scenario_categories?: ScenarioCategoryItem[];
}

// Order related types
export interface Order {
  id: number;
  user_id: number;
  /** 纯 SKU 订单可为空 */
  product_id?: number | null;
  sku_id?: number;
  spu_id?: number;
  group_id: number | null;
  /** 拼团当前状态（来自 groups.status） */
  group_status?: 'active' | 'completed' | 'failed';
  quantity: number;
  unit_price: number;
  total_price: number;
  items?: OrderItem[];
  status:
    | 'pending'
    | 'paid'
    | 'processing'
    | 'completed'
    | 'failed'
    | 'cancelled'
    | 'refunding'
    | 'refunded';
  /** 套餐包一键下单时由后端写入，用于销量统计 */
  entitlement_package_id?: number | null;
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  id: number;
  order_id: number;
  sku_id: number;
  spu_id: number;
  spu_name?: string;
  sku_code?: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  sku_type?: string;
  token_amount?: number;
  compute_points?: number;
  fulfilled_at?: string;
  pricing_version_id?: number;
  created_at: string;
  updated_at: string;
}

// Group purchase related types
export interface Group {
  id: number;
  product_id?: number | null;
  sku_id?: number;
  spu_id?: number;
  creator_id: number;
  target_count: number;
  current_count: number;
  status: 'active' | 'completed' | 'failed';
  deadline: string;
  created_at: string;
  updated_at: string;
  sku_name?: string;
  sku_type?: string;
  sku_specs?: string;
  group_discount_rate?: number;
}

// Token related types
export interface Token {
  id: number;
  user_id: number;
  balance: number;
  total_used?: number;
  total_earned?: number;
  created_at: string;
  updated_at: string;
}

export interface TokenTransaction {
  id: number;
  user_id: number;
  type: 'purchase' | 'usage' | 'transfer' | 'reward' | 'refund' | 'recharge' | 'expired';
  amount: number;
  reason?: string;
  order_id?: number;
  created_at: string;
}

/** 余额批次（后端 FIFO 扣减、按到期优先）；加油包入账多为带 expires_at 的批次 */
export interface TokenLot {
  id: number;
  remaining_amount: number;
  lot_type: string;
  expires_at: string | null;
  order_item_id?: number | null;
  created_at: string;
}

export interface RechargeOrder {
  id: number;
  user_id: number;
  amount: number;
  payment_method: 'alipay' | 'wechat' | 'balance';
  payment_id?: number;
  status: 'pending' | 'success' | 'failed';
  out_trade_no: string;
  created_at: string;
  updated_at: string;
}

export interface UserAPIKey {
  id: number;
  user_id: number;
  name: string;
  key_preview?: string;
  status: 'active' | 'inactive';
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

/** GET /tokens/api-usage-guide — 基于订阅与已支付订单的调用示例 */
export interface APIUsageGuideItem {
  source: string;
  spu_name?: string;
  sku_code?: string;
  provider_code: string;
  model_example: string;
  provider_slash_example?: string;
}

export interface APIUsageGuideResponse {
  items: APIUsageGuideItem[];
  default_model_example?: string;
  disclaimer: string;
}

// Payment related types
export interface Payment {
  id: number;
  order_id: number;
  amount: number;
  method: 'alipay' | 'wechat';
  status: 'pending' | 'success' | 'failed';
  created_at: string;
  updated_at: string;
}

// API Key related types
export interface APIKey {
  id: number;
  user_id: number;
  name: string;
  status: 'active' | 'inactive';
  created_at: string;
  updated_at: string;
}

// Referral related types
export interface ReferralCode {
  id: number;
  user_id: number;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface Referral {
  id: number;
  referrer_id: number;
  referee_id: number;
  code_used: string;
  status: 'active' | 'cancelled';
  created_at: string;
  referee_name?: string;
}

export interface ReferralReward {
  id: number;
  referrer_id: number;
  referee_id: number;
  order_id?: number;
  amount: number;
  status: 'pending' | 'paid' | 'cancelled';
  created_at: string;
  paid_at?: string;
  referee_name?: string;
}

export interface ReferralStats {
  total_referrals: number;
  total_rewards: number;
  pending_rewards: number;
  paid_rewards: number;
  available_rewards: number;
}

export interface ReferralWithdrawal {
  id: number;
  user_id: number;
  amount: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  method: 'alipay' | 'wechat' | 'bank';
  account_info: string;
  request_note?: string;
  reject_reason?: string;
  created_at: string;
  processed_at?: string;
  completed_at?: string;
}

export interface ReferralWithdrawalRequest {
  amount: number;
  method: string;
  account_info: string;
  request_note?: string;
}

// Merchant related types
export interface Merchant {
  id: number;
  user_id: number;
  company_name: string;
  business_license?: string;
  business_license_url?: string;
  id_card_front_url?: string;
  id_card_back_url?: string;
  attachments?: string;
  contact_name?: string;
  contact_phone?: string;
  contact_email?: string;
  address?: string;
  description?: string;
  logo_url?: string;
  status: 'pending' | 'reviewing' | 'active' | 'suspended' | 'rejected';
  /** 运营生命周期：trial | active | suspended */
  lifecycle_status?: 'trial' | 'active' | 'suspended';
  reviewed_at?: string;
  review_note?: string;
  created_at: string;
  updated_at: string;
}

export interface MerchantStats {
  total_products: number;
  active_products: number;
  total_sales: number;
  month_sales: number;
  total_orders: number;
  month_orders: number;
  group_success_rate?: number;
  success_groups?: number;
  pending_groups?: number;
  failed_groups?: number;
  week_sales?: number;
  week_growth?: number;
  new_customers?: number;
}

export interface MerchantSettlement {
  id: number;
  merchant_id: number;
  company_name?: string;
  period_start: string;
  period_end: string;
  total_sales: number;
  total_sales_cny: number;
  total_tokens: number;
  /** 周期内采购成本合计（人民币元），来自 api_usage_logs.procurement_cost_cny */
  total_procurement_cny?: number;
  platform_fee: number;
  settlement_amount: number;
  status: 'pending' | 'processing' | 'completed';
  settled_at?: string;
  created_at: string;
  updated_at: string;
  merchant_confirmed: boolean;
  merchant_confirmed_at?: string;
  finance_approved: boolean;
  finance_approved_at?: string;
}

export interface SettlementDispute {
  id: number;
  settlement_id: number;
  reason: string;
  status: 'pending' | 'resolved' | 'rejected';
  resolved_at?: string;
  resolved_by?: number;
  resolution_notes?: string;
  created_at: string;
}

export interface SettlementReconciliation {
  id: number;
  settlement_id: number;
  order_count: number;
  total_usage: number;
  total_amount: number;
  anomalies: string;
  reconciled_by: number;
  reconciled_at: string;
  created_at: string;
}

export interface SettlementItem {
  id: number;
  settlement_id: number;
  api_usage_log_id: number;
  user_id: number;
  merchant_id: number;
  provider: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  cost: number;
  created_at: string;
}

export interface BillingRecord {
  id: number;
  user_id: number;
  merchant_id: number;
  provider: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  token_usage: number;
  cost: number;
  request_id: string;
  status_code: number;
  latency_ms: number;
  created_at: string;
}

export interface BillingStats {
  total_cost: number;
  total_requests: number;
  total_tokens: number;
  average_latency: number;
  success_rate: number;
}

export interface BillingFilter {
  start_date?: string;
  end_date?: string;
  merchant_id?: number;
  year?: number;
  month?: number;
}

export interface MerchantOrder extends Order {
  product_name: string;
}

export interface MerchantAPIKey {
  id: number;
  merchant_id: number;
  name: string;
  provider: string;
  quota_limit: number | null;
  quota_used: number;
  status: 'active' | 'inactive';
  last_used_at?: string;
  created_at: string;
  updated_at: string;

  health_check_interval?: number;
  health_check_level?: 'high' | 'medium' | 'low' | 'daily';
  endpoint_url?: string;
  health_status?: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  health_error_message?: string;
  last_health_check_at?: string;
  consecutive_failures?: number;

  verified_at?: string;
  /** 与后端 merchant_api_keys.verification_result 对齐：verified | failed | pending 等 */
  verification_result?: 'verified' | 'success' | 'failed' | 'pending' | string;
  verification_message?: string;
  models_supported?: string[];

  cost_input_rate?: number;
  cost_output_rate?: number;
  profit_margin?: number;
}

export interface HealthStatus {
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  last_check_at?: string;
  consecutive_failures: number;
  latency_ms?: number;
}

export interface VerificationResult {
  id: number;
  api_key_id: number;
  verification_type: string;
  status: 'pending' | 'in_progress' | 'success' | 'failed';
  connection_test: boolean;
  connection_latency_ms?: number;
  models_found?: string[];
  models_count: number;
  pricing_verified: boolean;
  pricing_info?: Record<string, unknown>;
  error_code?: string;
  error_message?: string;
  started_at: string;
  completed_at?: string;
  retry_count: number;
}

export interface APIKeyUsage {
  id: number;
  name: string;
  provider: string;
  quota_limit: number | null;
  quota_used: number;
  usage_percentage: number;
}

// Cart item type
export interface CartItem {
  id: string;
  sku_id: number;
  product: Product;
  quantity: number;
  group_id?: number;
  sku_name?: string;
  sku_type?: string;
  sku_specs?: string;
}

// API Response types
export interface APIResponse<T> {
  code: number;
  message: string;
  data?: T;
}

export interface PaginatedResponse<T> {
  total: number;
  page: number;
  per_page: number;
  data: T[];
}
