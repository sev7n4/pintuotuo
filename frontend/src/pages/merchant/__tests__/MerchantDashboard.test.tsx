import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

jest.mock('@/stores/merchantStore', () => ({
  useMerchantStore: jest.fn(() => ({
    stats: { total_orders: 100, total_revenue: 5000, pending_orders: 10, completed_orders: 90 },
    orders: [],
    fetchStats: jest.fn(),
    fetchOrders: jest.fn(),
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

describe('MerchantDashboard', () => {
  let MerchantDashboard: React.FC

  beforeEach(async () => {
    MerchantDashboard = (await import('../MerchantDashboard')).default
  })

  it('renders dashboard with title', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantDashboard /></MemoryRouter>)
    })
    expect(screen.getByText('本月订单')).toBeInTheDocument()
  })

  it('displays recent orders card', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantDashboard /></MemoryRouter>)
    })
    await waitFor(() => {
      expect(screen.getByText('最近订单')).toBeInTheDocument()
    })
  })
})
