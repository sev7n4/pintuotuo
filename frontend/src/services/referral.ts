import api from './api';
import {
  ReferralStats,
  Referral,
  ReferralReward,
  ReferralWithdrawal,
  ReferralWithdrawalRequest,
  APIResponse,
  PaginatedResponse,
} from '@/types';

export const referralService = {
  getMyReferralCode: () => api.get<APIResponse<{ code: string }>>('/referrals/code'),

  validateReferralCode: (code: string) =>
    api.get<APIResponse<{ valid: boolean; referrer_id?: number; referrer_name?: string }>>(
      `/referrals/validate/${code}`
    ),

  bindReferralCode: (code: string) =>
    api.post<APIResponse<{ message: string }>>('/referrals/bind', { code }),

  getReferralStats: () => api.get<ReferralStats>('/referrals/stats'),

  getReferralList: (page?: number, perPage?: number) =>
    api.get<APIResponse<PaginatedResponse<Referral>>>('/referrals/list', {
      params: { page, per_page: perPage },
    }),

  getReferralRewards: (page?: number, perPage?: number, status?: string) =>
    api.get<APIResponse<PaginatedResponse<ReferralReward>>>('/referrals/rewards', {
      params: { page, per_page: perPage, status },
    }),

  payReferralRewards: (rewardIds: number[]) =>
    api.post<APIResponse<{ message: string }>>('/referrals/rewards/pay', { reward_ids: rewardIds }),

  getWithdrawalHistory: (page?: number, perPage?: number) =>
    api.get<APIResponse<PaginatedResponse<ReferralWithdrawal>>>('/referrals/withdrawals', {
      params: { page, per_page: perPage },
    }),

  requestWithdrawal: (request: ReferralWithdrawalRequest) =>
    api.post<APIResponse<{ message: string; withdrawal?: ReferralWithdrawal }>>(
      '/referrals/withdrawals',
      request
    ),
};
