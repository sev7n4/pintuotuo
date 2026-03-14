import { create } from 'zustand'
import { Product, PaginatedResponse } from '@types/index'
import { productService } from '@services/product'

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

  // Actions
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
      set({
        products: response.data.data,
        total: response.data.total,
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
      set({ isLoading: false })
      return response.data
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
      set({
        products: response.data.data,
        total: response.data.total,
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
