export interface MerchantSKU {
  id: number;
  merchant_id: number;
  sku_id: number;
  api_key_id?: number;
  status: 'active' | 'inactive';
  sales_count: number;
  total_sales_amount: number;
  created_at: string;
  updated_at: string;
}

export interface MerchantSKUDetail extends MerchantSKU {
  sku_code: string;
  sku_type: string;
  token_amount?: number;
  compute_points?: number;
  retail_price: number;
  original_price?: number;
  valid_days: number;
  group_enabled: boolean;
  group_discount_rate?: number;
  spu_name: string;
  model_provider: string;
  model_name: string;
  model_tier: string;
  api_key_name?: string;
  api_key_provider?: string;
  cost_input_rate: number;
  cost_output_rate: number;
  profit_margin: number;
  custom_pricing_enabled: boolean;
  spu_input_rate?: number;
  spu_output_rate?: number;
}

export interface AvailableSKU {
  id: number;
  sku_code: string;
  sku_type: string;
  token_amount?: number;
  compute_points?: number;
  retail_price: number;
  original_price?: number;
  valid_days: number;
  group_enabled: boolean;
  group_discount_rate?: number;
  spu_id: number;
  spu_name: string;
  model_provider: string;
  model_name: string;
  model_tier: string;
  spu_input_rate?: number;
  spu_output_rate?: number;
  is_selected: boolean;
}

export interface MerchantSKUCreateRequest {
  sku_id: number;
  api_key_id?: number;
  custom_pricing_enabled?: boolean;
  cost_input_rate?: number;
  cost_output_rate?: number;
  profit_margin?: number;
}

export interface MerchantSKUUpdateRequest {
  api_key_id?: number;
  status?: string;
  custom_pricing_enabled?: boolean;
  cost_input_rate?: number;
  cost_output_rate?: number;
  profit_margin?: number;
}
