import { render, screen, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

jest.mock('@/stores/merchantStore', () => ({
  useMerchantStore: jest.fn(() => ({
    profile: {
      company_name: 'Test Company',
      contact_name: 'John Doe',
      contact_phone: '13800138000',
      contact_email: 'test@example.com',
      address: 'Test Address',
      description: 'Test Description',
      status: 'active',
      business_license: 'license.pdf',
      verified_at: '2024-01-15T00:00:00Z',
      logo_url: 'https://example.com/logo.png',
    },
    fetchProfile: jest.fn(),
    updateProfile: jest.fn(),
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

describe('MerchantSettings', () => {
  let MerchantSettings: React.FC

  beforeEach(async () => {
    jest.clearAllMocks()
    MerchantSettings = (await import('../MerchantSettings')).default
  })

  it('renders settings page with title', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantSettings />
        </MemoryRouter>
      )
    })
    expect(screen.getByText('店铺设置')).toBeInTheDocument()
  })

  it('displays company name field', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantSettings />
        </MemoryRouter>
      )
    })
    expect(screen.getByText('公司名称')).toBeInTheDocument()
  })

  it('displays save button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantSettings />
        </MemoryRouter>
      )
    })
    expect(screen.getByText('保存设置')).toBeInTheDocument()
  })

  it('displays upload logo button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantSettings />
        </MemoryRouter>
      )
    })
    expect(screen.getByText('更换Logo')).toBeInTheDocument()
  })

  it('displays certification status card', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantSettings />
        </MemoryRouter>
      )
    })
    expect(screen.getByText('认证状态')).toBeInTheDocument()
    expect(screen.getByText('已认证')).toBeInTheDocument()
  })
})
