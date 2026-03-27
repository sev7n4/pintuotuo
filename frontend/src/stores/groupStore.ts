import { create } from 'zustand'
import { Group, APIResponse, PaginatedResponse } from '@/types'
import { groupService, JoinGroupResponse } from '@/services/group'

interface GroupState {
  groups: Group[]
  currentGroup: Group | null
  total: number
  isLoading: boolean
  error: string | null

  fetchGroups: (page?: number, perPage?: number) => Promise<Group[] | null>
  fetchGroupByID: (id: number) => Promise<void>
  createGroup: (productId: number, targetCount: number, deadline: string) => Promise<Group | null>
  joinGroup: (id: number) => Promise<void>
  cancelGroup: (id: number) => Promise<void>
  getGroupProgress: (id: number) => Promise<void>
  clearError: () => void
}

export const useGroupStore = create<GroupState>((set) => ({
  groups: [],
  currentGroup: null,
  total: 0,
  isLoading: false,
  error: null,

  fetchGroups: async (page = 1, perPage = 20) => {
    set({ isLoading: true, error: null })
    try {
      const response = await groupService.listGroups(page, perPage)
      const apiResponse = response.data as APIResponse<PaginatedResponse<Group>>
      const data = apiResponse.data
      set({
        groups: data?.data || [],
        total: data?.total || 0,
        isLoading: false,
      })
      return data?.data || []
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
      const apiResponse = response.data as APIResponse<Group>
      set({ currentGroup: apiResponse.data || null, isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取分组详情失败'
      set({ error: message, isLoading: false })
    }
  },

  createGroup: async (productId, targetCount, deadline): Promise<Group | null> => {
    set({ isLoading: true, error: null })
    try {
      const response = await groupService.createGroup({
        product_id: productId,
        target_count: targetCount,
        deadline: new Date(deadline).toISOString(),
      })
      const apiResponse = response.data as APIResponse<Group>
      const newGroup = apiResponse.data
      if (newGroup) {
        set((state) => ({
          groups: [newGroup, ...state.groups],
          currentGroup: newGroup,
          isLoading: false,
        }))
        return newGroup
      }
      set({ isLoading: false })
      return null
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
      const apiResponse = response.data as APIResponse<JoinGroupResponse>
      const updatedGroup = apiResponse.data?.group
      if (updatedGroup) {
        set((state) => ({
          groups: state.groups.map((g) =>
            g.id === id ? updatedGroup : g
          ),
          currentGroup: updatedGroup,
          isLoading: false,
        }))
      }
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
      const apiResponse = response.data as APIResponse<Group>
      set({ currentGroup: apiResponse.data || null, isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取分组进度失败'
      set({ error: message, isLoading: false })
    }
  },

  clearError: () => set({ error: null }),
}))
