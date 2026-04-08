import api from './api';
import { Product, APIResponse, PaginatedResponse, HomeData, Category, Group } from '@/types';

interface ProductFilters {
  page?: number;
  per_page?: number;
  status?: string;
  merchant_id?: number;
  category?: string;
  sort?: 'hot' | 'new' | 'price_asc' | 'price_desc';
}

interface CreateProductRequest {
  name: string;
  description: string;
  price: number;
  original_price?: number;
  stock: number;
  category?: string;
}

export const productService = {
  // Get home page data
  getHomeData: () => api.get<HomeData>('/catalog/home'),

  // Get hot products
  getHotProducts: (limit?: number) =>
    api.get<APIResponse<Product[]>>('/catalog/hot', { params: { limit } }),

  // Get new products
  getNewProducts: (limit?: number) =>
    api.get<APIResponse<Product[]>>('/catalog/new', { params: { limit } }),

  // Get categories
  getCategories: () => api.get<APIResponse<Category[]>>('/catalog/categories'),

  // Usage scenarios (C 端卖场筛选)
  getCatalogScenarios: () =>
    api.get<{ scenarios: Array<{ id: number; code: string; name: string; spu_count?: number }> }>(
      '/catalog/scenarios'
    ),

  // List products
  listProducts: (filters?: ProductFilters) =>
    api.get<APIResponse<PaginatedResponse<Product>>>('/catalog', { params: filters }),

  // Get product by ID
  getProductByID: (id: number) => api.get<APIResponse<Product>>(`/catalog/${id}`),

  // Get active groups for a product
  getProductGroups: (productId: number) =>
    api.get<APIResponse<Group[]>>(`/catalog/${productId}/groups`),

  // Search products
  searchProducts: (query: string, page?: number, perPage?: number) =>
    api.get<APIResponse<PaginatedResponse<Product>>>('/catalog/search', {
      params: { q: query, page, per_page: perPage },
    }),

  // Create product (merchant)
  createProduct: (data: CreateProductRequest) =>
    api.post<APIResponse<Product>>('/catalog/merchants', data),

  // Update product (merchant)
  updateProduct: (id: number, data: Partial<CreateProductRequest>) =>
    api.put<APIResponse<Product>>(`/catalog/merchants/${id}`, data),

  // Delete product (merchant)
  deleteProduct: (id: number) => api.delete<APIResponse<void>>(`/catalog/merchants/${id}`),
};
