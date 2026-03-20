import { render, screen, fireEvent, act, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import PaymentPage from '../PaymentPage'
import { useOrderStore } from '@/stores/orderStore'
import { paymentService } from '@/services/payment'

jest.mock('@/stores/orderStore')
jest.mock('@/services/payment')

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockUseOrderStore = useOrderStore as jest.MockedFunction<typeof useOrderStore>

describe('Payment Flow Integration Tests - User Experience Flow', () => {
  const mockOrder = {
    id: 1,
    product_id: 101,
    user_id: 1,
    quantity: 2,
    total_price: 200,
    status: 'pending',
    payment_method: null,
    payment_status: 'unpaid',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('TC-PAY-001: 选择支付方式', () => {
    test('should display available payment methods', async () => {
      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: mockOrder,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/1']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/选择支付方式/i)).toBeInTheDocument()
      })

      expect(screen.getByText(/支付宝/i)).toBeInTheDocument()
      expect(screen.getByText(/微信支付/i)).toBeInTheDocument()
    })

    test('should show order summary before payment', async () => {
      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: mockOrder,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/1']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/支付金额/i)).toBeInTheDocument()
      })
    })
  })

  describe('TC-PAY-002: 支付宝支付流程', () => {
    test('should process alipay payment successfully', async () => {
      const mockInitiatePayment = jest.fn().mockResolvedValue({
        data: {
          data: {
            id: 1,
            order_id: 1,
            method: 'alipay',
            amount: 200,
            status: 'success',
            transaction_id: 'trans_123',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        },
      })

      ;(paymentService as any).initiatePayment = mockInitiatePayment

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: mockOrder,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/1']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/支付宝/i)).toBeInTheDocument()
      })

      const alipayRadio = screen.getByText(/支付宝/i)
      await act(async () => {
        fireEvent.click(alipayRadio)
      })

      const payButton = screen.getByText(/立即支付/i)
      await act(async () => {
        fireEvent.click(payButton)
      })

      await waitFor(() => {
        expect(mockInitiatePayment).toHaveBeenCalledWith({
          order_id: 1,
          method: 'alipay',
        })
      })
    })
  })

  describe('TC-PAY-003: 微信支付流程', () => {
    test('should process wechat payment successfully', async () => {
      const mockInitiatePayment = jest.fn().mockResolvedValue({
        data: {
          data: {
            id: 1,
            order_id: 1,
            method: 'wechat',
            amount: 200,
            status: 'success',
            transaction_id: 'trans_456',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        },
      })

      ;(paymentService as any).initiatePayment = mockInitiatePayment

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: mockOrder,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/1']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/微信支付/i)).toBeInTheDocument()
      })

      const wechatRadio = screen.getByText(/微信支付/i)
      await act(async () => {
        fireEvent.click(wechatRadio)
      })

      const payButton = screen.getByText(/立即支付/i)
      await act(async () => {
        fireEvent.click(payButton)
      })

      await waitFor(() => {
        expect(mockInitiatePayment).toHaveBeenCalledWith({
          order_id: 1,
          method: 'wechat',
        })
      })
    })
  })

  describe('TC-PAY-004: 支付失败处理', () => {
    test('should handle payment failure and show retry option', async () => {
      const mockInitiatePayment = jest.fn().mockResolvedValue({
        data: {
          data: {
            id: 1,
            order_id: 1,
            method: 'alipay',
            amount: 200,
            status: 'failed',
            transaction_id: 'trans_789',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        },
      })

      ;(paymentService as any).initiatePayment = mockInitiatePayment

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: mockOrder,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/1']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      const payButton = await screen.findByText(/立即支付/i)
      await act(async () => {
        fireEvent.click(payButton)
      })

      await waitFor(() => {
        expect(screen.getByText(/支付失败/i)).toBeInTheDocument()
      })

      expect(screen.getByText(/重新支付/i)).toBeInTheDocument()
    })
  })

  describe('TC-PAY-005: 已支付订单处理', () => {
    test('should show already paid message for paid orders', async () => {
      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: {
          ...mockOrder,
          status: 'paid',
          payment_status: 'paid',
        },
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/1']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/订单已支付/i)).toBeInTheDocument()
      })
    })
  })

  describe('TC-PAY-006: 订单不存在处理', () => {
    test('should show empty state when order does not exist', async () => {
      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/999']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/订单不存在/i)).toBeInTheDocument()
      })
    })
  })

  describe('TC-PAY-007: 加载状态处理', () => {
    test('should show loading spinner while fetching order', async () => {
      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: null,
        isLoading: true,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/1']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      const spinner = document.querySelector('.ant-spin')
      expect(spinner).toBeTruthy()
    })
  })

  describe('TC-PAY-008: 完整支付旅程', () => {
    test('should complete full payment journey from order selection to confirmation', async () => {
      const mockInitiatePayment = jest.fn().mockResolvedValue({
        data: {
          data: {
            id: 1,
            order_id: 1,
            method: 'wechat',
            amount: 200,
            status: 'success',
            transaction_id: 'trans_123',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        },
      })

      ;(paymentService as any).initiatePayment = mockInitiatePayment

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: mockOrder,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        updateOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/1']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/选择支付方式/i)).toBeInTheDocument()
      })

      expect(screen.getByText(/支付金额/i)).toBeInTheDocument()

      const wechatRadio = screen.getByText(/微信支付/i)
      await act(async () => {
        fireEvent.click(wechatRadio)
      })

      const payButton = screen.getByText(/立即支付/i)
      await act(async () => {
        fireEvent.click(payButton)
      })

      await waitFor(() => {
        expect(screen.getByText(/支付成功/i)).toBeInTheDocument()
      }, { timeout: 5000 })

      const viewOrdersButton = screen.getByText(/查看订单/i)
      expect(viewOrdersButton).toBeInTheDocument()
    })
  })
})
