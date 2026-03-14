import { create } from 'zustand'

interface UIState {
  theme: 'light' | 'dark'
  sidebarCollapsed: boolean
  notifications: Array<{
    id: string
    type: 'success' | 'error' | 'warning' | 'info'
    message: string
  }>

  // Actions
  setTheme: (theme: 'light' | 'dark') => void
  toggleSidebar: () => void
  addNotification: (
    type: 'success' | 'error' | 'warning' | 'info',
    message: string
  ) => void
  removeNotification: (id: string) => void
  clearNotifications: () => void
}

export const useUIStore = create<UIState>((set) => ({
  theme: (localStorage.getItem('theme') as 'light' | 'dark') || 'light',
  sidebarCollapsed: false,
  notifications: [],

  setTheme: (theme) => {
    localStorage.setItem('theme', theme)
    set({ theme })
  },

  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),

  addNotification: (type, message) => {
    const id = `${type}-${Date.now()}`
    set((state) => ({
      notifications: [...state.notifications, { id, type, message }],
    }))
    // Auto remove after 3 seconds
    setTimeout(() => {
      set((state) => ({
        notifications: state.notifications.filter((n) => n.id !== id),
      }))
    }, 3000)
  },

  removeNotification: (id) => {
    set((state) => ({
      notifications: state.notifications.filter((n) => n.id !== id),
    }))
  },

  clearNotifications: () => set({ notifications: [] }),
}))
