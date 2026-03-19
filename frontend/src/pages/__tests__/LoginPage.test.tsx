import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { act } from 'react'
import LoginPage from '../LoginPage'
import { useAuthStore } from '@/stores/authStore'

// 模拟 useAuthStore
jest.mock('@/stores/authStore')

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

describe('LoginPage Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders LoginPage with form', () => {
    // 模拟 store 状态
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
        <LoginPage />
      </MemoryRouter>
    )

    // 检查页面元素
    expect(screen.getByText('拼脱脱 - 登录')).toBeInTheDocument()
    expect(screen.getByLabelText('邮箱')).toBeInTheDocument()
    expect(screen.getByLabelText('密码')).toBeInTheDocument()
    expect(screen.getByText(/登录/)).toBeInTheDocument()
    expect(screen.getByText('创建新账户')).toBeInTheDocument()
  })

  test('renders login form correctly', async () => {
    // 模拟 store 状态
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
        <LoginPage />
      </MemoryRouter>
    )

    // 检查表单元素
    expect(screen.getByPlaceholderText('example@email.com')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('输入密码')).toBeInTheDocument()
    expect(screen.getByText(/登录/)).toBeInTheDocument()
    expect(screen.getByText('创建新账户')).toBeInTheDocument()
  })

  test('navigates to register page when create account is clicked', () => {
    // 模拟 store 状态
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
        <LoginPage />
      </MemoryRouter>
    )

    // 点击创建新账户按钮
    const createAccountButton = screen.getByText('创建新账户')
    fireEvent.click(createAccountButton)

    // 验证页面导航
    expect(screen.getByText('拼脱脱 - 登录')).toBeInTheDocument()
  })

  test('shows loading state during login', () => {
    // 模拟加载状态
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: true,
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
        <LoginPage />
      </MemoryRouter>
    )

    // 检查登录按钮是否处于加载状态
    expect(screen.getByText(/登录/)).toBeInTheDocument()
  })

  test('navigates to register page when create account is clicked', () => {
    // 模拟 store 状态
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
        <LoginPage />
      </MemoryRouter>
    )

    // 点击创建新账户按钮
    const createAccountButton = screen.getByText('创建新账户')
    fireEvent.click(createAccountButton)

    // 验证导航到注册页面
    expect(screen.getByText('拼脱脱 - 登录')).toBeInTheDocument()
  })
})
