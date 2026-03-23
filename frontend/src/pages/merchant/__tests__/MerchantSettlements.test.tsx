import { render, screen, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

jest.mock('@/stores/merchantStore', () => ({
  useMerchantStore: jest.fn(() => ({
    settlements: [],
    stats: { total_settlements: 10000, pending_settlements: 5000 },
    fetchSettlements: jest.fn(),
    fetchStats: jest.fn(),
    requestSettlement: jest.fn(),
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

describe('MerchantSettlements', () => {
  let MerchantSettlements: React.FC

  beforeEach(async () => {
    MerchantSettlements = (await import('../MerchantSettlements')).default
  })

  it('renders settlements page with title', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantSettlements /></MemoryRouter>)
    })
    expect(screen.getByText('结算管理')).toBeInTheDocument()
  })

  it('displays settlement stats', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantSettlements /></MemoryRouter>)
    })
    await waitFor(() => {
      expect(screen.getByText('累计结算')).toBeInTheDocument()
    })
  })

  it('displays pending settlements', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantSettlements /></MemoryRouter>)
    })
    await waitFor(() => {
      expect(screen.getByText('待结算')).toBeInTheDocument()
    })
  })

  it('displays request settlement button', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantSettlements /></MemoryRouter>)
    })
    const requestButtons = screen.getAllByRole('button')
    const requestButton = requestButtons.find(btn => btn.textContent?.includes('申请结算'))
    expect(requestButton).toBeInTheDocument()
  })

  it('displays settlement records section', async () => {
    await act(async () => {
      render(<MemoryRouter><MerchantSettlements /></MemoryRouter>)
    })
    await waitFor(() => {
      expect(screen.getByText('结算记录')).toBeInTheDocument()
    })
  })
})
