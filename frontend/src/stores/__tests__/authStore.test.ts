import { useAuthStore } from '../authStore'

const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value },
    removeItem: (key: string) => { delete store[key] },
    clear: () => { store = {} },
  }
})()

Object.defineProperty(window, 'localStorage', { value: localStorageMock })

describe('AuthStore', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  describe('clearError', () => {
    it('should clear error', () => {
      const store = useAuthStore.getState()
      store.clearError()
      
      const newState = useAuthStore.getState()
      expect(newState.error).toBeNull()
    })
  })

  describe('isAuthenticated', () => {
    it('should reflect authentication state', () => {
      const store = useAuthStore.getState()
      expect(typeof store.isAuthenticated).toBe('boolean')
    })
  })

  describe('initial state', () => {
    it('should have isLoading as false', () => {
      const state = useAuthStore.getState()
      expect(state.isLoading).toBe(false)
    })

    it('should have error as null', () => {
      const state = useAuthStore.getState()
      expect(state.error).toBeNull()
    })
  })

  describe('login function', () => {
    it('should exist', () => {
      const state = useAuthStore.getState()
      expect(typeof state.login).toBe('function')
    })
  })

  describe('register function', () => {
    it('should exist', () => {
      const state = useAuthStore.getState()
      expect(typeof state.register).toBe('function')
    })
  })

  describe('logout function', () => {
    it('should exist', () => {
      const state = useAuthStore.getState()
      expect(typeof state.logout).toBe('function')
    })
  })

  describe('fetchUser function', () => {
    it('should exist', () => {
      const state = useAuthStore.getState()
      expect(typeof state.fetchUser).toBe('function')
    })
  })
})
