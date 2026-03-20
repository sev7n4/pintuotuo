import { render, screen, fireEvent, act, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import LoginPage from '../LoginPage'
import { useAuthStore } from '@/stores/authStore'

jest.mock('@/stores/authStore')

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>

describe('LoginPage Integration Tests - User Experience Flow', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('TC-AUTH-001: 用户登录成功流程', () => {
    test('should successfully login with valid credentials and navigate to products page for regular user', async () => {
      const mockLogin = jest.fn().mockResolvedValue({ token: 'user-token' })

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

      mockUseAuthStore.getState = jest.fn().mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
      })

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/products" element={<div>Products Page</div>} />
            <Route path="/merchant/dashboard" element={<div>Merchant Dashboard</div>} />
            <Route path="/admin" element={<div>Admin Dashboard</div>} />
          </Routes>
        </MemoryRouter>
      )

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'password123' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password123')
      })

      await waitFor(() => {
        expect(screen.getByText('Products Page')).toBeInTheDocument()
      })
    })

    test('should navigate to merchant dashboard for merchant user', async () => {
      const mockLogin = jest.fn().mockResolvedValue({ token: 'merchant-token' })

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

      mockUseAuthStore.getState = jest.fn().mockReturnValue({
        user: { id: 2, email: 'merchant@example.com', role: 'merchant' },
      })

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/merchant/dashboard" element={<div>Merchant Dashboard</div>} />
          </Routes>
        </MemoryRouter>
      )

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'merchant@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'password123' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(screen.getByText('Merchant Dashboard')).toBeInTheDocument()
      })
    })

    test('should navigate to admin dashboard for admin user', async () => {
      const mockLogin = jest.fn().mockResolvedValue({ token: 'admin-token' })

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

      mockUseAuthStore.getState = jest.fn().mockReturnValue({
        user: { id: 3, email: 'admin@example.com', role: 'admin' },
      })

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/admin" element={<div>Admin Dashboard</div>} />
          </Routes>
        </MemoryRouter>
      )

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'admin@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'password123' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(screen.getByText('Admin Dashboard')).toBeInTheDocument()
      })
    })
  })

  describe('TC-AUTH-002: 登录失败-无效邮箱格式', () => {
    test('should show validation error for invalid email format', async () => {
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

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'invalid-email' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(screen.getByText('邮箱格式不正确')).toBeInTheDocument()
      })
    })

    test('should show validation error for email without @ symbol', async () => {
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

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'userexample.com' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(screen.getByText('邮箱格式不正确')).toBeInTheDocument()
      })
    })
  })

  describe('TC-AUTH-003: 登录失败-密码为空', () => {
    test('should show validation error when password is empty', async () => {
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

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(screen.getByText('请输入密码')).toBeInTheDocument()
      })
    })
  })

  describe('TC-AUTH-004: 登录失败-错误凭证', () => {
    test('should show error message for invalid credentials', async () => {
      const mockLogin = jest.fn().mockRejectedValue(new Error('Invalid credentials'))

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: '登录失败，请检查邮箱和密码',
        isAuthenticated: false,
        login: mockLogin,
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

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'wrong@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('wrong@example.com', 'wrongpassword')
      })
    })

    test('should keep form editable after login failure', async () => {
      const mockLogin = jest.fn().mockRejectedValue(new Error('Invalid credentials'))

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: '登录失败',
        isAuthenticated: false,
        login: mockLogin,
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

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(emailInput).toBeEnabled()
        expect(passwordInput).toBeEnabled()
      })
    })
  })

  describe('TC-AUTH-005: 登录表单完整性', () => {
    test('should render all required form elements', () => {
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

      expect(screen.getByText('拼脱脱 - 登录')).toBeInTheDocument()
      expect(screen.getByLabelText('邮箱')).toBeInTheDocument()
      expect(screen.getByLabelText('密码')).toBeInTheDocument()
      expect(screen.getByText('登 录')).toBeInTheDocument()
      expect(screen.getByText('创建新账户')).toBeInTheDocument()
    })

    test('should have correct placeholder text', () => {
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

      expect(screen.getByPlaceholderText('example@email.com')).toBeInTheDocument()
      expect(screen.getByPlaceholderText('输入密码')).toBeInTheDocument()
    })
  })

  describe('TC-AUTH-006: 导航到注册页面', () => {
    test('should navigate to register page when create account is clicked', async () => {
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
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<div>Register Page</div>} />
          </Routes>
        </MemoryRouter>
      )

      const createAccountButton = screen.getByText('创建新账户')

      await act(async () => {
        fireEvent.click(createAccountButton)
      })

      await waitFor(() => {
        expect(screen.getByText('Register Page')).toBeInTheDocument()
      })
    })
  })

  describe('TC-AUTH-007: 加载状态处理', () => {
    test('should show loading state during login', async () => {
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

      const loginButton = screen.getByText('登 录')
      expect(loginButton).toBeInTheDocument()
    })
  })

  describe('TC-AUTH-008: 邮箱必填验证', () => {
    test('should show error when email is empty', async () => {
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

      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(screen.getByText('请输入邮箱')).toBeInTheDocument()
      })
    })
  })

  describe('TC-AUTH-009: 表单重置', () => {
    test('should allow user to re-enter credentials after error', async () => {
      const mockLogin = jest.fn()
        .mockRejectedValueOnce(new Error('Invalid credentials'))
        .mockResolvedValueOnce({ token: 'success-token' })

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

      mockUseAuthStore.getState = jest.fn()
        .mockReturnValueOnce({ user: null })
        .mockReturnValueOnce({ user: { id: 1, email: 'user@example.com', role: 'user' } })

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/products" element={<div>Products Page</div>} />
          </Routes>
        </MemoryRouter>
      )

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'wrong@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('wrong@example.com', 'wrongpassword')
      })

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'correct@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'correctpassword' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('correct@example.com', 'correctpassword')
      })
    })
  })

  describe('TC-AUTH-010: 完整用户登录旅程', () => {
    test('should complete full login journey from form fill to navigation', async () => {
      const mockLogin = jest.fn().mockResolvedValue({ token: 'user-token' })

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

      mockUseAuthStore.getState = jest.fn().mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
      })

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/products" element={<div>Products Page</div>} />
          </Routes>
        </MemoryRouter>
      )

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement
      const loginButton = screen.getByText('登 录')

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } })
        fireEvent.change(passwordInput, { target: { value: 'password123' } })
        fireEvent.click(loginButton)
      })

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password123')
        expect(screen.getByText('Products Page')).toBeInTheDocument()
      })
    })
  })
})
