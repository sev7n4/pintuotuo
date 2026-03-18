import api from './api'
import { Product, APIResponse, PaginatedResponse } from '@/types'

interface ProductFilters {
  page?: number
  per_page?: number
  status?: string
  merchant_id?: number
}

interface CreateProductRequest {
  name: string
  description: string
  price: number
  stock: number
}

export const productService = {
  // List products
  listProducts: (filters?: ProductFilters) =>
    api.get<APIResponse<PaginatedResponse<Product>>>('/products', { params: filters }),

  // Get product by ID
  getProductByID: (id: number) =>
    api.get<APIResponse<Product>>(`/products/${id}`),

  // Search products
  searchProducts: (query: string) =>
    api.get<APIResponse<PaginatedResponse<Product>>>('/products/search', {
      params: { q: query },
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
