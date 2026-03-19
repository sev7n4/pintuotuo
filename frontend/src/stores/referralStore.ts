import { create } from 'zustand'
import { ReferralStats, Referral, ReferralReward, PaginatedResponse } from '@/types'
import { referralService } from '@/services/referral'

interface ReferralState {
  referralCode: string
  stats: ReferralStats | null
  referrals: Referral[]
  rewards: ReferralReward[]
  isLoading: boolean
  error: string | null

  fetchReferralCode: () => Promise<void>
  fetchStats: () => Promise<void>
  fetchReferrals: (page?: number, perPage?: number) => Promise<void>
  fetchRewards: (page?: number, perPage?: number, status?: string) => Promise<void>
  bindReferralCode: (code: string) => Promise<boolean>
  clearError: () => void
}

export const useReferralStore = create<ReferralState>((set, get) => ({
  referralCode: '',
  stats: null,
  referrals: [],
  rewards: [],
  isLoading: false,
  error: null,

  fetchReferralCode: async () => {
    set({ isLoading: true, error: null })
    try {
      const response = await referralService.getMyReferralCode()
      const code = response.data?.code
      set({ referralCode: typeof code === 'string' ? code : '', isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取邀请码失败'
      set({ error: message, isLoading: false })
    }
  },

  fetchStats: async () => {
    set({ isLoading: true, error: null })
    try {
      const response = await referralService.getReferralStats()
      set({ stats: response.data, isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取统计数据失败'
      set({ error: message, isLoading: false })
    }
  },

  fetchReferrals: async (page = 1, perPage = 20) => {
    set({ isLoading: true, error: null })
    try {
      const response = await referralService.getReferralList(page, perPage)
      const data = response.data as unknown as PaginatedResponse<Referral>
      set({ referrals: data?.data || [], isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取邀请列表失败'
      set({ error: message, isLoading: false })
    }
  },

  fetchRewards: async (page = 1, perPage = 20, status?: string) => {
    set({ isLoading: true, error: null })
    try {
      const response = await referralService.getReferralRewards(page, perPage, status)
      const data = response.data as unknown as PaginatedResponse<ReferralReward>
      set({ rewards: data?.data || [], isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取返利记录失败'
      set({ error: message, isLoading: false })
    }
  },

  bindReferralCode: async (code: string) => {
    set({ isLoading: true, error: null })
    try {
      await referralService.bindReferralCode(code)
      set({ isLoading: false })
      get().fetchStats()
      return true
    } catch (error) {
      const message = error instanceof Error ? error.message : '绑定邀请码失败'
      set({ error: message, isLoading: false })
      return false
    }
  },

  clearError: () => set({ error: null }),
}))
