import api from './api';
import { Token, TokenTransaction, UserAPIKey, APIResponse } from '@/types';

export const tokenService = {
  getBalance: () => api.get<Token>('/tokens/balance'),

  getConsumption: () => api.get<TokenTransaction[]>('/tokens/consumption'),

  transfer: (recipientId: number, amount: number) =>
    api.post<{ message: string }>('/tokens/transfer', { recipient_id: recipientId, amount }),

  getAPIKeys: () => api.get<APIResponse<UserAPIKey[]>>('/tokens/keys'),

  createAPIKey: (name: string) => api.post<UserAPIKey>('/tokens/keys', { name }),

  updateAPIKey: (id: number, data: Partial<UserAPIKey>) =>
    api.put<UserAPIKey>(`/tokens/keys/${id}`, data),

  deleteAPIKey: (id: number) => api.delete(`/tokens/keys/${id}`),
};
