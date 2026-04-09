export interface ModelProvider {
  id: number;
  code: string;
  name: string;
  api_base_url?: string;
  api_format: string;
  billing_type?: string;
  /** OpenAI 兼容入口：无前缀 model 名按最长前缀匹配到本厂商 */
  compat_prefixes?: string[];
  cache_enabled: boolean;
  cache_discount_rate?: number;
  status: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface InputRange {
  min_tokens: number;
  max_tokens: number;
  label: string;
  surcharge?: number;
}

export interface SegmentRule {
  input_range: string;
  multiplier: number;
}

export interface BillingAdapter {
  type: 'flat' | 'segment' | 'tiered';
  segment_config?: SegmentRule[];
  cache_enabled: boolean;
  cache_discount_rate?: number;
}

export interface RoutingRules {
  auto_route: boolean;
  default_range?: string;
  range_mapping?: Record<string, string>;
}

export interface BatchConfig {
  enabled: boolean;
  discount_rate: number;
  async_only: boolean;
}

export interface SPU {
  id: number;
  spu_code: string;
  name: string;

  model_provider: string;
  provider_model_id?: string;
  provider_api_endpoint?: string;
  provider_auth_type?: string;
  provider_billing_type?: string;
  provider_input_rate?: number;
  provider_output_rate?: number;

  model_name: string;
  model_version?: string;
  model_tier: 'pro' | 'lite' | 'mini' | 'vision';

  context_window?: number;
  max_output_tokens?: number;
  supported_functions?: string[];

  base_compute_points: number;
  billing_coefficient?: number;

  description?: string;
  features?: string[];
  thumbnail_url?: string;

  input_length_ranges?: InputRange[];
  billing_adapter?: BillingAdapter;
  routing_rules?: RoutingRules;
  batch_inference?: BatchConfig;

  status: string;
  sort_order: number;
  total_sales_count: number;
  average_rating?: number;
  /** 管理端列表/详情：关联 SKU 总数 */
  sku_count?: number;
  /** 管理端列表/详情：在售 SKU 数 */
  active_sku_count?: number;
  created_at: string;
  updated_at: string;
}

export interface SKU {
  id: number;
  spu_id: number;
  spu_code?: string;
  spu_name?: string;
  sku_code: string;
  merchant_id?: number;
  sku_type: 'token_pack' | 'subscription' | 'concurrent' | 'trial';
  token_amount?: number;
  compute_points?: number;
  subscription_period?: 'monthly' | 'quarterly' | 'yearly';
  is_unlimited: boolean;
  fair_use_limit?: number;
  tpm_limit?: number;
  rpm_limit?: number;
  concurrent_requests?: number;
  valid_days: number;
  retail_price: number;
  wholesale_price?: number;
  original_price?: number;
  stock: number;
  daily_limit?: number;
  group_enabled: boolean;
  min_group_size: number;
  max_group_size: number;
  group_discount_rate?: number;
  is_trial: boolean;
  trial_duration_days?: number;
  status: string;
  is_promoted: boolean;
  promotion_labels?: string[];
  new_user_offer?: string;
  full_reduction?: string;
  coupons?: Array<{
    name: string;
    threshold?: number;
    discount?: number;
  }>;
  sales_count: number;
  created_at: string;
  updated_at: string;
  model_provider?: string;
  model_name?: string;
  /** 上游 model 参数示例，优先于 model_name */
  provider_model_id?: string;
  model_tier?: string;
  /** SPU 维度累计销量（公开列表/详情由后端 JOIN 返回） */
  spu_total_sales_count?: number;
  /** SPU 维度评分，详情页可与 SKU 销量组合展示 */
  spu_average_rating?: number;
  /** SPU 缩略图（若后端返回） */
  thumbnail_url?: string;
}

export interface SKUWithSPU extends SKU {
  spu_name: string;
  /** 关联 SPU 上下架状态 */
  spu_status?: string;
  /** SKU+SPU 均在售时商户端可选，与 /merchants/skus/available 一致 */
  sellable?: boolean;
  model_provider: string;
  model_name: string;
  model_tier: string;
}

export interface ComputePointAccount {
  id: number;
  user_id: number;
  balance: number;
  total_earned: number;
  total_used: number;
  total_expired: number;
  created_at: string;
  updated_at: string;
}

export interface ComputePointTransaction {
  id: number;
  user_id: number;
  type: 'purchase' | 'reward' | 'usage' | 'refund' | 'expire' | 'group_bonus';
  amount: number;
  balance_after: number;
  order_id?: number;
  sku_id?: number;
  description?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface UserSubscription {
  id: number;
  user_id: number;
  sku_id: number;
  start_date: string;
  end_date: string;
  used_tokens: number;
  used_compute_points: number;
  status: 'active' | 'expired' | 'cancelled';
  auto_renew: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserSubscriptionWithSKU extends UserSubscription {
  sku_code: string;
  spu_name: string;
  retail_price: number;
}

export interface SPUCreateRequest {
  spu_code: string;
  name: string;
  model_provider: string;
  model_name: string;
  model_version?: string;
  model_tier: 'pro' | 'lite' | 'mini' | 'vision';
  context_window?: number;
  max_output_tokens?: number;
  supported_functions?: string[];
  base_compute_points?: number;
  provider_input_rate?: number;
  provider_output_rate?: number;
  description?: string;
  features?: string[];
  thumbnail_url?: string;
  status?: string;
  sort_order?: number;
}

export interface SKUCreateRequest {
  spu_id: number;
  sku_code: string;
  sku_type: 'token_pack' | 'subscription' | 'concurrent' | 'trial';
  token_amount?: number;
  compute_points?: number;
  subscription_period?: 'monthly' | 'quarterly' | 'yearly';
  is_unlimited?: boolean;
  fair_use_limit?: number;
  tpm_limit?: number;
  rpm_limit?: number;
  concurrent_requests?: number;
  valid_days?: number;
  retail_price: number;
  wholesale_price?: number;
  original_price?: number;
  stock?: number;
  daily_limit?: number;
  group_enabled?: boolean;
  min_group_size?: number;
  max_group_size?: number;
  group_discount_rate?: number;
  is_trial?: boolean;
  trial_duration_days?: number;
  status?: string;
  is_promoted?: boolean;
}

export interface SKUUpdateRequest {
  retail_price?: number;
  wholesale_price?: number;
  original_price?: number;
  stock?: number;
  daily_limit?: number;
  group_enabled?: boolean;
  min_group_size?: number;
  max_group_size?: number;
  group_discount_rate?: number;
  status?: string;
  is_promoted?: boolean;
}

export const MODEL_TIER_LABELS: Record<string, string> = {
  pro: '旗舰版',
  lite: '标准版',
  mini: '轻量版',
  vision: '多模态版',
};

export const SKU_TYPE_LABELS: Record<string, string> = {
  token_pack: 'Token包',
  subscription: '订阅套餐',
  concurrent: '并发套餐',
  trial: '试用套餐',
};

export const SUBSCRIPTION_PERIOD_LABELS: Record<string, string> = {
  monthly: '月度',
  quarterly: '季度',
  yearly: '年度',
};
