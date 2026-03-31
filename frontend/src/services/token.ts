import api from './api';
import { Token, TokenTransaction, UserAPIKey, APIResponse, RechargeOrder } from '@/types';

export const tokenService = {
  getBalance: () => api.get<Token>('/tokens/balance'),

  getConsumption: () => api.get<TokenTransaction[]>('/tokens/consumption'),

  transfer: (
    amount: number,
    opts: { recipientId?: number; recipientEmail?: string }
  ) => {
    const body: Record<string, unknown> = { amount };
    if (opts.recipientEmail != null && opts.recipientEmail.trim() !== '') {
      body.recipient_email = opts.recipientEmail.trim();
    } else if (opts.recipientId != null && opts.recipientId > 0) {
      body.recipient_id = opts.recipientId;
    }
    return api.post<{ message: string }>('/tokens/transfer', body);
  },

  mockCompleteRechargeOrder: (orderId: number) =>
    api.post<APIResponse<RechargeOrder>>(`/tokens/recharge/orders/${orderId}/mock-pay`),

  getAPIKeys: () => api.get<UserAPIKey[] | APIResponse<UserAPIKey[]>>('/tokens/keys'),

  createAPIKey: (name: string) =>
    api.post<{ id: number; key: string; name: string; status: 'active' | 'inactive' }>('/tokens/keys', { name }),

  updateAPIKey: (id: number, data: Partial<UserAPIKey>) =>
    api.put<UserAPIKey>(`/tokens/keys/${id}`, data),

  deleteAPIKey: (id: number) => api.delete(`/tokens/keys/${id}`),

  createRechargeOrder: (amount: number, method: 'alipay' | 'wechat' | 'balance') =>
    api.post<APIResponse<RechargeOrder>>('/tokens/recharge', { amount, method }),

  getRechargeOrders: (page = 1, perPage = 10) =>
    api.get<
      APIResponse<{
        total: number;
        page: number;
        per_page: number;
        data: RechargeOrder[];
      }>
    >(`/tokens/recharge/orders?page=${page}&per_page=${perPage}`),
};
