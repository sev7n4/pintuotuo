import { create } from 'zustand'
import { Product, PaginatedResponse, APIResponse } from '@/types'
import { productService } from '@/services/product'

interface ProductFilters {
  status?: string
  merchant_id?: number
  page: number
  per_page: number
}

interface ProductState {
  products: Product[]
  total: number
  filters: ProductFilters
  isLoading: boolean
  error: string | null

  fetchProducts: (filters?: Partial<ProductFilters>) => Promise<void>
  fetchProductByID: (id: number) => Promise<Product | null>
  searchProducts: (query: string) => Promise<void>
  setFilters: (filters: Partial<ProductFilters>) => void
  clearError: () => void
}

export const useProductStore = create<ProductState>((set, get) => ({
  products: [],
  total: 0,
  filters: { page: 1, per_page: 20 },
  isLoading: false,
  error: null,

  fetchProducts: async (newFilters) => {
    set({ isLoading: true, error: null })
    try {
      const filters = { ...get().filters, ...newFilters }
      const response = await productService.listProducts(filters)
      const apiResponse = response.data as APIResponse<PaginatedResponse<Product>>
      set({
        products: apiResponse.data?.data || [],
        total: apiResponse.data?.total || 0,
        filters,
        isLoading: false,
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取产品列表失败'
      set({ error: message, isLoading: false })
    }
  },

  fetchProductByID: async (id) => {
    set({ isLoading: true, error: null })
    try {
      const response = await productService.getProductByID(id)
      const apiResponse = response.data as APIResponse<Product>
      set({ isLoading: false })
      return apiResponse.data || null
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取产品详情失败'
      set({ error: message, isLoading: false })
      return null
    }
  },

  searchProducts: async (query) => {
    set({ isLoading: true, error: null })
    try {
      const response = await productService.searchProducts(query)
      const apiResponse = response.data as APIResponse<PaginatedResponse<Product>>
      set({
        products: apiResponse.data?.data || [],
        total: apiResponse.data?.total || 0,
        isLoading: false,
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : '搜索失败'
      set({ error: message, isLoading: false })
    }
  },

  setFilters: (newFilters) => {
    set((state) => ({
      filters: { ...state.filters, ...newFilters, page: 1 },
    }))
  },

  clearError: () => set({ error: null }),
}))
