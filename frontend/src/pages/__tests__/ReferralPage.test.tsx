import { render, screen, fireEvent, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import ReferralPage from '../ReferralPage'
import { useReferralStore } from '@/stores/referralStore'
import { useAuthStore } from '@/stores/authStore'
import { message } from 'antd'

// 模拟 useReferralStore
jest.mock('@/stores/referralStore')

// 模拟 useAuthStore
jest.mock('@/stores/authStore')

// 模拟 CSS 模块
jest.mock('../ReferralPage.module.css', () => ({}))

// 模拟 message
jest.mock('antd', () => {
  const antd = jest.requireActual('antd');
  return {
    ...antd,
    message: {
      success: jest.fn(),
      error: jest.fn(),
    },
    Table: antd.Table,
    Tabs: antd.Tabs,
    TabPane: antd.Tabs.TabPane,
    Input: antd.Input,
    Button: antd.Button,
    Space: antd.Space,
    Spin: antd.Spin,
    Card: antd.Card,
    Row: antd.Row,
    Col: antd.Col,
    Typography: antd.Typography,
    Statistic: antd.Statistic,
  };
});

const mockUseReferralStore = useReferralStore as jest.MockedFunction<typeof useReferralStore>
const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>
const mockMessageSuccess = message.success as jest.MockedFunction<typeof message.success>

// 模拟 clipboard API
Object.defineProperty(navigator, 'clipboard', {
  value: {
    writeText: jest.fn().mockResolvedValue(undefined),
  },
  writable: true,
})

describe('ReferralPage Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders ReferralPage with referral code and stats', () => {
    // 模拟 store 状态
    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: { totalReferrals: 10, totalRewards: 100, pendingRewards: 20, paidRewards: 80 },
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
    expect(screen.getByText('复制邀请码')).toBeInTheDocument()
    expect(screen.getByText('分享链接')).toBeInTheDocument()
  })

  test('handles copy referral code', async () => {
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

    // 点击复制邀请码按钮
    const copyButton = screen.getByText('复制邀请码')
    await act(async () => {
      fireEvent.click(copyButton)
    })

    // 验证 clipboard.writeText 被调用
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('TESTCODE')
    expect(mockMessageSuccess).toHaveBeenCalledWith('邀请码已复制到剪贴板')
  })

  test('handles share referral link', async () => {
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

    // 点击分享链接按钮
    const shareButton = screen.getByText('分享链接')
    await act(async () => {
      fireEvent.click(shareButton)
    })

    // 验证 clipboard.writeText 被调用
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(expect.stringContaining('/register?code=TESTCODE'))
    expect(mockMessageSuccess).toHaveBeenCalledWith('分享链接已复制到剪贴板')
  })

  test('renders referral and reward tables', () => {
    const mockReferrals = [
      {
        id: 1,
        referee_name: '用户1',
        code_used: 'TESTCODE',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
      },
    ]

    const mockRewards = [
      {
        id: 1,
        referee_name: '用户1',
        amount: 100,
        status: 'paid',
        created_at: '2024-01-01T00:00:00Z',
      },
    ]

    // 模拟 store 状态
    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: { totalReferrals: 10, totalRewards: 100 },
      referrals: mockReferrals,
      rewards: mockRewards,
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

    // 检查表格渲染
    expect(screen.getByText('邀请记录')).toBeInTheDocument()
    expect(screen.getByText('返利明细')).toBeInTheDocument()
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

    // 输入邀请码
    const input = screen.getByPlaceholderText('输入好友的邀请码') as HTMLInputElement
    await act(async () => {
      fireEvent.change(input, { target: { value: 'FRIEND12' } })
    })

    // 点击绑定按钮
    const bindButton = screen.getByRole('button', { name: /绑 定/i })
    await act(async () => {
      fireEvent.click(bindButton)
    })

    // 验证绑定函数被调用
    expect(mockBindReferralCode).toHaveBeenCalledWith('FRIEND12')
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
    // 检查是否显示加载状态（通过检查是否存在ant-spin元素）
    expect(document.querySelector('.ant-spin-spinning')).toBeInTheDocument()
  })
})
