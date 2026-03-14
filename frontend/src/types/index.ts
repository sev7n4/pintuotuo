// User related types
export interface User {
  id: number
  email: string
  name: string
  role: 'user' | 'merchant' | 'admin'
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
  stock: number
  status: 'active' | 'inactive' | 'archived'
  created_at: string
  updated_at: string
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
