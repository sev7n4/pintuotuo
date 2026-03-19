import { renderHook } from '@testing-library/react'
import { useAuth } from '../useAuth'
import { useAuthStore } from '@/stores/authStore'

// 模拟 useAuthStore
jest.mock('@/stores/authStore')

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>

describe('useAuth hook', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('returns auth state when not authenticated', () => {
    // 模拟未认证状态
    const mockAuthState = {
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
    }

    mockUseAuthStore.mockReturnValue(mockAuthState)

    const { result } = renderHook(() => useAuth())

    expect(result.current.user).toBeNull()
    expect(result.current.token).toBeNull()
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
  })

  test('returns auth state when authenticated', () => {
    // 模拟认证状态
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      role: 'user',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    const mockAuthState = {
      user: mockUser,
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
    }

    mockUseAuthStore.mockReturnValue(mockAuthState)

    const { result } = renderHook(() => useAuth())

    expect(result.current.user).toEqual(mockUser)
    expect(result.current.token).toBe('test-token')
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
    expect(result.current.isAuthenticated).toBe(true)
  })

  test('calls fetchUser when authenticated but no user', () => {
    // 模拟已认证但无用户信息状态
    const mockAuthState = {
      user: null,
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
    }

    mockUseAuthStore.mockReturnValue(mockAuthState)

    renderHook(() => useAuth())

    expect(mockAuthState.fetchUser).toHaveBeenCalled()
  })

  test('does not call fetchUser when not authenticated', () => {
    // 模拟未认证状态
    const mockAuthState = {
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
    }

    mockUseAuthStore.mockReturnValue(mockAuthState)

    renderHook(() => useAuth())

    expect(mockAuthState.fetchUser).not.toHaveBeenCalled()
  })

  test('does not call fetchUser when authenticated and has user', () => {
    // 模拟已认证且有用户信息状态
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      role: 'user',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    const mockAuthState = {
      user: mockUser,
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
    }

    mockUseAuthStore.mockReturnValue(mockAuthState)

    renderHook(() => useAuth())

    expect(mockAuthState.fetchUser).not.toHaveBeenCalled()
  })
})