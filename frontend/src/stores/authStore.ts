import { create } from 'zustand'
import { User, APIResponse } from '@/types'
import { authService } from '@/services/auth'
import { userService } from '@/services/user'

interface LoginResponse {
  user: User
  token: string
}

interface AuthState {
  user: User | null
  token: string | null
  isLoading: boolean
  error: string | null
  isAuthenticated: boolean

  login: (email: string, password: string) => Promise<void>
  register: (email: string, name: string, password: string) => Promise<void>
  logout: () => Promise<void>
  fetchUser: () => Promise<void>
  clearError: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: localStorage.getItem('auth_token'),
  isLoading: false,
  error: null,
  isAuthenticated: !!localStorage.getItem('auth_token'),

  login: async (email, password) => {
    set({ isLoading: true, error: null })
    try {
      const response = await authService.login({ email, password })
      const apiResponse = response.data as APIResponse<LoginResponse>
      const data = apiResponse.data
      if (data) {
        const { user, token } = data
        localStorage.setItem('auth_token', token)
        set({ user, token, isAuthenticated: true, isLoading: false })
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : '登录失败'
      set({ error: message, isLoading: false })
      throw error
    }
  },

  register: async (email, name, password) => {
    set({ isLoading: true, error: null })
    try {
      const response = await authService.register({ email, name, password })
      const apiResponse = response.data as APIResponse<LoginResponse>
      const data = apiResponse.data
      if (data) {
        const { user, token } = data
        localStorage.setItem('auth_token', token)
        set({ user, token, isAuthenticated: true, isLoading: false })
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : '注册失败'
      set({ error: message, isLoading: false })
      throw error
    }
  },

  logout: async () => {
    set({ isLoading: true })
    try {
      await authService.logout()
    } finally {
      localStorage.removeItem('auth_token')
      set({ user: null, token: null, isAuthenticated: false, isLoading: false })
    }
  },

  fetchUser: async () => {
    if (!localStorage.getItem('auth_token')) return

    set({ isLoading: true, error: null })
    try {
      const response = await userService.getCurrentUser()
      const apiResponse = response.data as APIResponse<User>
      set({ user: apiResponse.data || null, isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取用户信息失败'
      set({ error: message, isLoading: false })
    }
  },

  clearError: () => set({ error: null }),
}))
