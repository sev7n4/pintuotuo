import { useProductStore } from '../productStore'
import { Product } from '@/types'

jest.mock('@/services/product')

const mockProduct: Product = {
  id: 1,
  merchant_id: 1,
  name: 'Test Product',
  description: 'Test Description',
  price: 99.99,
  stock: 100,
  status: 'active',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockProduct2: Product = {
  id: 2,
  merchant_id: 1,
  name: 'Test Product 2',
  description: 'Test Description 2',
  price: 49.99,
  stock: 50,
  status: 'active',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

describe('ProductStore', () => {
  beforeEach(() => {
    useProductStore.setState({
      products: [],
      total: 0,
      filters: { page: 1, per_page: 20 },
      isLoading: false,
      error: null,
    })
    jest.clearAllMocks()
  })

  describe('initial state', () => {
    it('should have correct initial state', () => {
      const state = useProductStore.getState()
      
      expect(state.products).toEqual([])
      expect(state.total).toBe(0)
      expect(state.filters).toEqual({ page: 1, per_page: 20 })
      expect(state.isLoading).toBe(false)
      expect(state.error).toBeNull()
    })
  })

  describe('setFilters', () => {
    it('should update filters and reset page to 1', () => {
      const { setFilters } = useProductStore.getState()
      
      setFilters({ status: 'active', per_page: 10 })
      
      const state = useProductStore.getState()
      expect(state.filters.status).toBe('active')
      expect(state.filters.per_page).toBe(10)
      expect(state.filters.page).toBe(1)
    })

    it('should merge with existing filters', () => {
      useProductStore.setState({ filters: { page: 5, per_page: 50, status: 'active' } })
      
      const { setFilters } = useProductStore.getState()
      setFilters({ merchant_id: 1 })
      
      const state = useProductStore.getState()
      expect(state.filters.merchant_id).toBe(1)
      expect(state.filters.status).toBe('active')
      expect(state.filters.per_page).toBe(50)
      expect(state.filters.page).toBe(1)
    })
  })

  describe('clearError', () => {
    it('should clear error', () => {
      useProductStore.setState({ error: 'Test error' })
      
      const { clearError } = useProductStore.getState()
      clearError()
      
      expect(useProductStore.getState().error).toBeNull()
    })
  })

  describe('products state', () => {
    it('should store products', () => {
      useProductStore.setState({ products: [mockProduct, mockProduct2] })
      
      const state = useProductStore.getState()
      expect(state.products).toHaveLength(2)
      expect(state.products[0].name).toBe('Test Product')
    })

    it('should store total count', () => {
      useProductStore.setState({ total: 100 })
      
      const state = useProductStore.getState()
      expect(state.total).toBe(100)
    })
  })

  describe('loading state', () => {
    it('should track loading state', () => {
      useProductStore.setState({ isLoading: true })
      
      expect(useProductStore.getState().isLoading).toBe(true)
      
      useProductStore.setState({ isLoading: false })
      
      expect(useProductStore.getState().isLoading).toBe(false)
    })
  })

  describe('error state', () => {
    it('should store error message', () => {
      useProductStore.setState({ error: 'Failed to fetch products' })
      
      const state = useProductStore.getState()
      expect(state.error).toBe('Failed to fetch products')
    })
  })
})
