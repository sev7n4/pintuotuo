import { render, screen, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

jest.mock('@/stores/merchantStore', () => ({
  useMerchantStore: jest.fn(() => ({
    products: [
      { id: 1, name: 'Test Product 1', price: 100, stock: 10, status: 'active' },
      { id: 2, name: 'Test Product 2', price: 200, stock: 5, status: 'active' },
    ],
    fetchProducts: jest.fn(),
    createProduct: jest.fn(),
    updateProduct: jest.fn(),
    deleteProduct: jest.fn(),
    isLoading: false,
  })),
}))

jest.mock('@/stores/authStore', () => ({
  useAuthStore: jest.fn(() => ({
    user: { id: 1, name: 'Test Merchant', role: 'merchant' },
  })),
}))

jest.mock('@/services/api', () => ({
  default: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}))

describe('MerchantProducts', () => {
  let MerchantProducts: React.FC

  beforeEach(async () => {
    MerchantProducts = (await import('../MerchantProducts')).default
  })

  it('renders products page with title', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantProducts /></MemoryRouter>)
    })
    expect(screen.getByText('商品管理')).toBeInTheDocument()
  })

  it('displays add product button', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantProducts /></MemoryRouter>)
    })
    const addButtons = screen.getAllByRole('button')
    const addButton = addButtons.find(btn => btn.textContent?.includes('添加商品'))
    expect(addButton).toBeInTheDocument()
  })

  it('displays product table container', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantProducts /></MemoryRouter>)
    })
    const tableElement = document.querySelector('.ant-table')
    expect(tableElement).toBeInTheDocument()
  })
})
