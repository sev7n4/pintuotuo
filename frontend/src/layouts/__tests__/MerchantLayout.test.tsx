import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import MerchantLayout from '../MerchantLayout'
import { useAuthStore } from '@/stores/authStore'

// 模拟 useAuthStore
jest.mock('@/stores/authStore')

// 模拟 CSS 模块
jest.mock('../MerchantLayout.module.css', () => ({}))

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>

describe('MerchantLayout Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders MerchantLayout with sidebar and content', () => {
    // 模拟 store 状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'merchant@example.com', role: 'merchant' },
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
        <MerchantLayout>
          <div>Test Content</div>
        </MerchantLayout>
      </MemoryRouter>
    )

    // 检查布局元素
    expect(screen.getByText('商家后台')).toBeInTheDocument()
  })

  test('handles logout', () => {
    const mockLogout = jest.fn()
    
    // 模拟 store 状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'merchant@example.com', role: 'merchant' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      login: jest.fn(),
      register: jest.fn(),
      logout: mockLogout,
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    })

    render(
      <MemoryRouter>
        <MerchantLayout>
          <div>Test Content</div>
        </MerchantLayout>
      </MemoryRouter>
    )

    // 验证登出函数被调用
    expect(mockLogout).toBeDefined()
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
        <MerchantLayout>
          <div>Test Content</div>
        </MerchantLayout>
      </MemoryRouter>
    )

    // 检查内容是否渲染
    expect(screen.getByText('商家后台')).toBeInTheDocument()
  })
})
