import { render, screen, waitFor, act, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import AdminLayout from '../AdminLayout'
import { useAuthStore } from '@/stores/authStore'
import { message } from 'antd'

jest.mock('@/stores/authStore')
jest.mock('../MerchantLayout.module.css', () => ({}))
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    error: jest.fn(),
    success: jest.fn(),
  },
}))

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>

const mockLocalStorage = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
  }
})()

const mockSessionStorage = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
  }
})()

Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })
Object.defineProperty(window, 'sessionStorage', { value: mockSessionStorage })

describe('AdminLayout Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockLocalStorage.clear()
    mockSessionStorage.clear()
  })

  test('renders AdminLayout with sidebar and content for admin user', async () => {
    mockLocalStorage.setItem('auth_token', 'test-token')
    
    const mockFetchUser = jest.fn().mockResolvedValue(undefined)
    
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'admin@example.com', name: 'Test Admin', role: 'admin', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.getByText('运营管理')).toBeInTheDocument()
    })
  })

  test('redirects to login when no token', async () => {
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.queryByText('运营管理')).not.toBeInTheDocument()
    })
  })

  test('redirects non-admin user to home with error message', async () => {
    mockLocalStorage.setItem('auth_token', 'test-token')
    
    const mockFetchUser = jest.fn().mockResolvedValue(undefined)
    
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', name: 'Test User', role: 'user', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith('无权限访问管理后台')
      expect(screen.queryByText('运营管理')).not.toBeInTheDocument()
    })
  })

  test('shows loading state while checking auth', async () => {
    mockLocalStorage.setItem('auth_token', 'test-token')
    
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn().mockImplementation(() => new Promise(() => {})),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    expect(document.querySelector('.ant-spin-spinning')).toBeInTheDocument()
  })

  test('redirects to login when fetchUser fails', async () => {
    mockLocalStorage.setItem('auth_token', 'test-token')
    
    const mockFetchUser = jest.fn().mockRejectedValue(new Error('Token expired'))
    
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(mockFetchUser).toHaveBeenCalled()
      expect(mockLocalStorage.getItem('auth_token')).toBeNull()
    })
  })

  test('uses sessionStorage token when localStorage token is not present', async () => {
    mockSessionStorage.setItem('auth_token', 'session-token')
    
    const mockFetchUser = jest.fn().mockResolvedValue(undefined)
    
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'admin@example.com', name: 'Test Admin', role: 'admin', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      token: 'session-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.getByText('运营管理')).toBeInTheDocument()
    })
  })

  test('clears sessionStorage token when fetchUser fails', async () => {
    mockSessionStorage.setItem('auth_token', 'session-token')
    
    const mockFetchUser = jest.fn().mockRejectedValue(new Error('Token expired'))
    
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: 'session-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(mockFetchUser).toHaveBeenCalled()
      expect(mockSessionStorage.getItem('auth_token')).toBeNull()
    })
  })

  test('returns null when user is not authenticated after auth check', async () => {
    mockLocalStorage.setItem('auth_token', 'test-token')
    
    const mockFetchUser = jest.fn().mockResolvedValue(undefined)
    
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: false,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.queryByText('运营管理')).not.toBeInTheDocument()
    })
  })

  test('redirects merchant user to home with error message', async () => {
    mockLocalStorage.setItem('auth_token', 'test-token')
    
    const mockFetchUser = jest.fn().mockResolvedValue(undefined)
    
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'merchant@example.com', name: 'Test Merchant', role: 'merchant', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(message.error).toHaveBeenCalledWith('无权限访问管理后台')
      expect(screen.queryByText('运营管理')).not.toBeInTheDocument()
    })
  })

  test('handles menu click navigation', async () => {
    mockLocalStorage.setItem('auth_token', 'test-token')
    
    const mockFetchUser = jest.fn().mockResolvedValue(undefined)
    
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'admin@example.com', name: 'Test Admin', role: 'admin', created_at: '2024-01-01T00:00:00Z', updated_at: '2024-01-01T00:00:00Z' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter initialEntries={['/admin']}>
          <AdminLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.getByText('运营管理')).toBeInTheDocument()
    })

    const userMenuItem = screen.getByText('用户管理')
    await act(async () => {
      fireEvent.click(userMenuItem)
    })
  })
})
