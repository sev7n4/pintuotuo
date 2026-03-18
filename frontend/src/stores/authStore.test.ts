import { useAuthStore } from './authStore'
import { authService } from '../services/auth'
import { userService } from '../services/user'
import { act } from 'react-dom/test-utils'

// Mock services
jest.mock('../services/auth')
jest.mock('../services/user')

describe('authStore', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
    // Reset store state
    const { logout } = useAuthStore.getState()
    act(() => {
      logout()
    })
  })

  it('should have initial state', () => {
    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.token).toBeNull()
    expect(state.isAuthenticated).toBe(false)
    expect(state.isLoading).toBe(false)
    expect(state.error).toBeNull()
  })

  it('should login successfully', async () => {
    const mockUser = { id: '1', email: 'test@example.com', name: 'Test User' }
    const mockToken = 'test-token'
    const mockResponse = { data: { user: mockUser, token: mockToken } }

    ;(authService.login as jest.Mock).mockResolvedValue(mockResponse)

    await act(async () => {
      await useAuthStore.getState().login('test@example.com', 'password')
    })

    const state = useAuthStore.getState()
    expect(state.user).toEqual(mockUser)
    expect(state.token).toBe(mockToken)
    expect(state.isAuthenticated).toBe(true)
    expect(state.isLoading).toBe(false)
    expect(localStorage.getItem('auth_token')).toBe(mockToken)
  })

  it('should handle login error', async () => {
    const errorMessage = 'Invalid credentials'
    ;(authService.login as jest.Mock).mockRejectedValue(new Error(errorMessage))

    await act(async () => {
      try {
        await useAuthStore.getState().login('test@example.com', 'wrong-password')
      } catch (e) {
        // Expected error
      }
    })

    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.error).toBe(errorMessage)
    expect(state.isLoading).toBe(false)
  })

  it('should logout successfully', async () => {
    // Set initial state
    localStorage.setItem('auth_token', 'test-token')
    
    ;(authService.logout as jest.Mock).mockResolvedValue({ data: {} })

    await act(async () => {
      await useAuthStore.getState().logout()
    })

    const state = useAuthStore.getState()
    expect(state.user).toBeNull()
    expect(state.token).toBeNull()
    expect(state.isAuthenticated).toBe(false)
    expect(localStorage.getItem('auth_token')).toBeNull()
  })
})
