import { create } from 'zustand'
import { Group, PaginatedResponse } from '@types/index'
import { groupService } from '@services/group'

interface GroupState {
  groups: Group[]
  currentGroup: Group | null
  total: number
  isLoading: boolean
  error: string | null

  // Actions
  fetchGroups: (page?: number, perPage?: number) => Promise<Group[] | null>
  fetchGroupByID: (id: number) => Promise<void>
  createGroup: (productId: number, targetCount: number, deadline: string) => Promise<void>
  joinGroup: (id: number) => Promise<void>
  cancelGroup: (id: number) => Promise<void>
  getGroupProgress: (id: number) => Promise<void>
  clearError: () => void
}

export const useGroupStore = create<GroupState>((set, get) => ({
  groups: [],
  currentGroup: null,
  total: 0,
  isLoading: false,
  error: null,

  fetchGroups: async (page = 1, perPage = 20) => {
    set({ isLoading: true, error: null })
    try {
      const response = await groupService.listGroups(page, perPage)
      const data = response.data as PaginatedResponse<Group>
      set({
        groups: data.data,
        total: data.total,
        isLoading: false,
      })
      return data.data
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取分组列表失败'
      set({ error: message, isLoading: false })
      return null
    }
  },

  fetchGroupByID: async (id) => {
    set({ isLoading: true, error: null })
    try {
      const response = await groupService.getGroupByID(id)
      set({ currentGroup: response.data, isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取分组详情失败'
      set({ error: message, isLoading: false })
    }
  },

  createGroup: async (productId, targetCount, deadline) => {
    set({ isLoading: true, error: null })
    try {
      const response = await groupService.createGroup({
        product_id: productId,
        target_count: targetCount,
        deadline: new Date(deadline).toISOString(),
      })
      set((state) => ({
        groups: [response.data, ...state.groups],
        currentGroup: response.data,
        isLoading: false,
      }))
    } catch (error) {
      const message = error instanceof Error ? error.message : '创建分组失败'
      set({ error: message, isLoading: false })
      throw error
    }
  },

  joinGroup: async (id) => {
    set({ isLoading: true, error: null })
    try {
      const response = await groupService.joinGroup(id)
      set((state) => ({
        groups: state.groups.map((g) =>
          g.id === id ? response.data : g
        ),
        currentGroup: response.data,
        isLoading: false,
      }))
    } catch (error) {
      const message = error instanceof Error ? error.message : '加入分组失败'
      set({ error: message, isLoading: false })
      throw error
    }
  },

  cancelGroup: async (id) => {
    set({ isLoading: true, error: null })
    try {
      await groupService.cancelGroup(id)
      set((state) => ({
        groups: state.groups.filter((g) => g.id !== id),
        isLoading: false,
      }))
    } catch (error) {
      const message = error instanceof Error ? error.message : '取消分组失败'
      set({ error: message, isLoading: false })
      throw error
    }
  },

  getGroupProgress: async (id) => {
    set({ isLoading: true, error: null })
    try {
      const response = await groupService.getGroupProgress(id)
      set({ currentGroup: response.data, isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取分组进度失败'
      set({ error: message, isLoading: false })
    }
  },

  clearError: () => set({ error: null }),
}))
