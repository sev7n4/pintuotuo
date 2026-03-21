// User related types
export interface User {
  id: number
  email: string
  name: string
  role: 'user' | 'merchant' | 'admin'
  avatar_url?: string
  created_at: string
  updated_at: string
}

// Product related types
export interface Product {
  id: number
  merchant_id: number
  name: string
  description: string
  price: number
  original_price?: number
  stock: number
  sold_count?: number
  category?: string
  status: 'active' | 'inactive' | 'archived'
  created_at: string
  updated_at: string
}

// Category type
export interface Category {
  name: string
  count: number
}

// Banner type
export interface Banner {
  id: number
  title: string
  image: string
  link: string
}

// Home data type
export interface HomeData {
  banners: Banner[]
  hot: Product[]
  new: Product[]
  categories: Category[]
}

// Order related types
export interface Order {
  id: number
  user_id: number
  product_id: number
  group_id: number
  quantity: number
  total_price: number
  status: 'pending' | 'paid' | 'completed' | 'failed'
  created_at: string
  updated_at: string
}

// Group purchase related types
export interface Group {
  id: number
  product_id: number
  creator_id: number
  target_count: number
  current_count: number
  status: 'active' | 'completed' | 'failed'
  deadline: string
  created_at: string
  updated_at: string
}

// Token related types
export interface Token {
  id: number
  user_id: number
  balance: number
  total_used?: number
  total_earned?: number
  created_at: string
  updated_at: string
}

export interface TokenTransaction {
  id: number
  user_id: number
  type: 'purchase' | 'usage' | 'transfer' | 'reward' | 'refund'
  amount: number
  reason?: string
  order_id?: number
  created_at: string
}

export interface UserAPIKey {
  id: number
  user_id: number
  name: string
  key_preview?: string
  status: 'active' | 'inactive'
  last_used_at?: string
  created_at: string
  updated_at: string
}

// Payment related types
export interface Payment {
  id: number
  order_id: number
  amount: number
  method: 'alipay' | 'wechat'
  status: 'pending' | 'success' | 'failed'
  created_at: string
  updated_at: string
}

// API Key related types
export interface APIKey {
  id: number
  user_id: number
  name: string
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

// Referral related types
export interface ReferralCode {
  id: number
  user_id: number
  code: string
  created_at: string
  updated_at: string
}

export interface Referral {
  id: number
  referrer_id: number
  referee_id: number
  code_used: string
  status: 'active' | 'cancelled'
  created_at: string
  referee_name?: string
}

export interface ReferralReward {
  id: number
  referrer_id: number
  referee_id: number
  order_id?: number
  amount: number
  status: 'pending' | 'paid' | 'cancelled'
  created_at: string
  paid_at?: string
  referee_name?: string
}

export interface ReferralStats {
  total_referrals: number
  total_rewards: number
  pending_rewards: number
  paid_rewards: number
  available_rewards: number
}

export interface ReferralWithdrawal {
  id: number
  user_id: number
  amount: number
  status: 'pending' | 'processing' | 'completed' | 'failed'
  method: 'alipay' | 'wechat' | 'bank'
  account_info: string
  request_note?: string
  reject_reason?: string
  created_at: string
  processed_at?: string
  completed_at?: string
}

export interface ReferralWithdrawalRequest {
  amount: number
  method: string
  account_info: string
  request_note?: string
}

// Merchant related types
export interface Merchant {
  id: number
  user_id: number
  company_name: string
  business_license?: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  address?: string
  description?: string
  logo_url?: string
  status: 'pending' | 'active' | 'suspended' | 'rejected'
  verified_at?: string
  created_at: string
  updated_at: string
}

export interface MerchantStats {
  total_products: number
  active_products: number
  total_sales: number
  month_sales: number
  total_orders: number
  month_orders: number
}

export interface MerchantSettlement {
  id: number
  merchant_id: number
  period_start: string
  period_end: string
  total_sales: number
  platform_fee: number
  settlement_amount: number
  status: 'pending' | 'processing' | 'completed'
  settled_at?: string
  created_at: string
  updated_at: string
}

export interface MerchantOrder extends Order {
  product_name: string
}

export interface MerchantAPIKey {
  id: number
  merchant_id: number
  name: string
  provider: string
  quota_limit: number
  quota_used: number
  status: 'active' | 'inactive'
  last_used_at?: string
  created_at: string
  updated_at: string
}

export interface APIKeyUsage {
  id: number
  name: string
  provider: string
  quota_limit: number
  quota_used: number
  usage_percentage: number
}

// Cart item type
export interface CartItem {
  id: string
  product_id: number
  product: Product
  quantity: number
  group_id?: number
}

// API Response types
export interface APIResponse<T> {
  code: number
  message: string
  data?: T
}

export interface PaginatedResponse<T> {
  total: number
  page: number
  per_page: number
  data: T[]
}
