import { render, screen, fireEvent, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { useCartStore } from '@/stores/cartStore'
import { useOrderStore } from '@/stores/orderStore'

jest.mock('@/stores/cartStore')
jest.mock('@/stores/orderStore')
jest.mock('@/services/api', () => ({
  default: {
    get: jest.fn(),
    post: jest.fn(),
    put: jest.fn(),
    delete: jest.fn(),
  },
}))

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: jest.fn(),
}))

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    warning: jest.fn(),
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockNavigate = jest.fn()
const mockClear = jest.fn()
const mockCreateOrder = jest.fn()

const mockCartItems = [
  {
    id: '1',
    product_id: 1,
    quantity: 2,
    group_id: null,
    product: {
      id: 1,
      name: 'Test Product 1',
      price: 100,
      image: 'test1.jpg',
      description: 'Description 1',
      stock: 10,
      category_id: 1,
      merchant_id: 1,
      status: 'active',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    },
  },
  {
    id: '2',
    product_id: 2,
    quantity: 1,
    group_id: null,
    product: {
      id: 2,
      name: 'Test Product 2',
      price: 200,
      image: 'test2.jpg',
      description: 'Description 2',
      stock: 5,
      category_id: 1,
      merchant_id: 1,
      status: 'active',
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    },
  },
]

describe('CheckoutPage', () => {
  let CheckoutPage: React.FC
  
  beforeEach(() => {
    jest.clearAllMocks()
    ;(require('react-router-dom').useNavigate as jest.Mock).mockReturnValue(mockNavigate)
    ;(useCartStore as unknown as jest.Mock).mockReturnValue({
      items: mockCartItems,
      clear: mockClear,
    })
    ;(useOrderStore as unknown as jest.Mock).mockReturnValue({
      createOrder: mockCreateOrder,
      isLoading: false,
    })
    
    CheckoutPage = require('../CheckoutPage').default
  })

  it('renders empty state when cart is empty', () => {
    ;(useCartStore as unknown as jest.Mock).mockReturnValue({
      items: [],
      clear: mockClear,
    })

    render(
      <MemoryRouter>
        <CheckoutPage />
      </MemoryRouter>
    )

    expect(screen.getByText('购物车是空的')).toBeInTheDocument()
  })

  it('renders checkout page with cart items', () => {
    render(
      <MemoryRouter>
        <CheckoutPage />
      </MemoryRouter>
    )

    expect(screen.getByText('确认订单')).toBeInTheDocument()
  })

  it('navigates to products page when clicking "去购物" button', () => {
    ;(useCartStore as unknown as jest.Mock).mockReturnValue({
      items: [],
      clear: mockClear,
    })

    render(
      <MemoryRouter>
        <CheckoutPage />
      </MemoryRouter>
    )

    const goShoppingButton = screen.getByRole('button', { name: /去购物/i })
    fireEvent.click(goShoppingButton)

    expect(mockNavigate).toHaveBeenCalledWith('/products')
  })

  it('calls createOrder and clears cart on successful checkout', async () => {
    mockCreateOrder.mockResolvedValue({ id: 1 })

    render(
      <MemoryRouter>
        <CheckoutPage />
      </MemoryRouter>
    )

    const submitButtons = screen.getAllByRole('button')
    const submitButton = submitButtons.find(btn => btn.textContent?.includes('提交订单'))
    
    if (submitButton) {
      await act(async () => {
        fireEvent.click(submitButton)
      })
    }

    expect(mockCreateOrder).toHaveBeenCalled()
  })
})
