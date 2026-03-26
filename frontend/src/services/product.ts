import api from './api'
import { Product, APIResponse, PaginatedResponse, HomeData, Category, Group } from '@/types'

interface ProductFilters {
  page?: number
  per_page?: number
  status?: string
  merchant_id?: number
  category?: string
  sort?: 'hot' | 'new' | 'price_asc' | 'price_desc'
}

interface CreateProductRequest {
  name: string
  description: string
  price: number
  original_price?: number
  stock: number
  category?: string
}

export const productService = {
  // Get home page data
  getHomeData: () =>
    api.get<HomeData>('/products/home'),

  // Get hot products
  getHotProducts: (limit?: number) =>
    api.get<APIResponse<Product[]>>('/products/hot', { params: { limit } }),

  // Get new products
  getNewProducts: (limit?: number) =>
    api.get<APIResponse<Product[]>>('/products/new', { params: { limit } }),

  // Get categories
  getCategories: () =>
    api.get<APIResponse<Category[]>>('/products/categories'),

  // List products
  listProducts: (filters?: ProductFilters) =>
    api.get<APIResponse<PaginatedResponse<Product>>>('/products', { params: filters }),

  // Get product by ID
  getProductByID: (id: number) =>
    api.get<APIResponse<Product>>(`/products/${id}`),

  // Get active groups for a product
  getProductGroups: (productId: number) =>
    api.get<APIResponse<Group[]>>(`/products/${productId}/groups`),

  // Search products
  searchProducts: (query: string, page?: number, perPage?: number) =>
    api.get<APIResponse<PaginatedResponse<Product>>>('/products/search', {
      params: { q: query, page, per_page: perPage },
    }),

  // Create product (merchant)
  createProduct: (data: CreateProductRequest) =>
    api.post<APIResponse<Product>>('/products/merchants', data),

  // Update product (merchant)
  updateProduct: (id: number, data: Partial<CreateProductRequest>) =>
    api.put<APIResponse<Product>>(`/products/merchants/${id}`, data),

  // Delete product (merchant)
  deleteProduct: (id: number) =>
    api.delete<APIResponse<void>>(`/products/merchants/${id}`),
}
