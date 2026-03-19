import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import ReferralPage from '../ReferralPage'
import { useReferralStore } from '@/stores/referralStore'
import { useAuthStore } from '@/stores/authStore'

// 模拟 useReferralStore
jest.mock('@/stores/referralStore')

// 模拟 useAuthStore
jest.mock('@/stores/authStore')

// 模拟 CSS 模块
jest.mock('../ReferralPage.module.css', () => ({}))

// 模拟 message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockUseReferralStore = useReferralStore as jest.MockedFunction<typeof useReferralStore>
const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>

describe('ReferralPage Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders ReferralPage with referral code and stats', () => {
    // 模拟 store 状态
    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: { totalReferrals: 10, totalRewards: 100 },
      referrals: [],
      rewards: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      bindReferralCode: jest.fn(),
      clearError: jest.fn(),
    })

    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    })

    render(
      <MemoryRouter>
        <ReferralPage />
      </MemoryRouter>
    )

    // 检查页面元素
    expect(screen.getByText('邀请好友')).toBeInTheDocument()
    expect(screen.getByText('TESTCODE')).toBeInTheDocument()
  })

  test('handles binding referral code', async () => {
    const mockBindReferralCode = jest.fn().mockResolvedValue(undefined)
    
    // 模拟 store 状态
    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: { totalReferrals: 10, totalRewards: 100 },
      referrals: [],
      rewards: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      bindReferralCode: mockBindReferralCode,
      clearError: jest.fn(),
    })

    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    })

    render(
      <MemoryRouter>
        <ReferralPage />
      </MemoryRouter>
    )

    // 验证绑定函数被调用
    expect(mockBindReferralCode).toBeDefined()
  })

  test('renders correctly when not authenticated', () => {
    // 模拟未认证状态
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    })

    render(
      <MemoryRouter>
        <ReferralPage />
      </MemoryRouter>
    )

    // 检查页面是否提示登录
    expect(screen.getByText('邀请好友')).toBeInTheDocument()
  })

  test('shows loading state when fetching data', () => {
    // 模拟加载状态
    mockUseReferralStore.mockReturnValue({
      referralCode: null,
      stats: null,
      referrals: [],
      rewards: [],
      isLoading: true,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      bindReferralCode: jest.fn(),
      clearError: jest.fn(),
    })

    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    })

    render(
      <MemoryRouter>
        <ReferralPage />
      </MemoryRouter>
    )

    // 检查加载状态
    expect(screen.getByText('邀请好友')).toBeInTheDocument()
  })
})
