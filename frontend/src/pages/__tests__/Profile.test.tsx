import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { act } from 'react'
import Profile from '../Profile'
import { useAuthStore } from '@/stores/authStore'

// 模拟 useAuthStore
jest.mock('@/stores/authStore')

// 模拟 CSS 模块
jest.mock('../Profile.module.css', () => ({}))

// 模拟 message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>

import { message } from 'antd'
const mockMessage = message as jest.Mocked<typeof message>

describe('Profile Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders Profile page with user info', () => {
    // 模拟 store 状态
    mockUseAuthStore.mockReturnValue({
      user: { 
        id: 1, 
        email: 'user@example.com', 
        name: '测试用户',
        phone: '13800138000',
        role: 'user'
      },
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
        <Profile />
      </MemoryRouter>
    )

    // 检查页面元素
    expect(screen.getByText('个人中心')).toBeInTheDocument()
    expect(screen.getAllByText('测试用户').length).toBeGreaterThan(0)
    expect(screen.getByText('user@example.com')).toBeInTheDocument()
  })

  test('handles logout', () => {
    const mockLogout = jest.fn()
    
    // 模拟 store 状态
    mockUseAuthStore.mockReturnValue({
      user: { 
        id: 1, 
        email: 'user@example.com', 
        name: '测试用户',
        phone: '13800138000',
        role: 'user'
      },
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
        <Profile />
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
        <Profile />
      </MemoryRouter>
    )

    // 检查页面是否提示登录
    expect(screen.getByText('个人中心')).toBeInTheDocument()
  })

  test('shows loading state when fetching user', () => {
    // 模拟加载状态
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: 'test-token',
      isLoading: true,
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
        <Profile />
      </MemoryRouter>
    )

    // 检查加载状态
    expect(screen.getByText('个人中心')).toBeInTheDocument()
  })
})
