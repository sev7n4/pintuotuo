import { render, screen, waitFor, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import MerchantLayout from '../MerchantLayout'
import { useAuthStore } from '@/stores/authStore'

jest.mock('@/stores/authStore')
jest.mock('../MerchantLayout.module.css', () => ({}))

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

describe('MerchantLayout Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockLocalStorage.clear()
    mockSessionStorage.clear()
  })

  test('renders MerchantLayout with sidebar and content for merchant user', async () => {
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
          <MerchantLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.getByText('商家后台')).toBeInTheDocument()
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
          <MerchantLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.queryByText('商家后台')).not.toBeInTheDocument()
    })
  })

  test('redirects non-merchant user to home', async () => {
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
          <MerchantLayout />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.queryByText('商家后台')).not.toBeInTheDocument()
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
          <MerchantLayout />
        </MemoryRouter>
      )
    })

    expect(document.querySelector('.ant-spin-spinning')).toBeInTheDocument()
  })
})
