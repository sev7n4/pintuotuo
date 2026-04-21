import api from './api';
import {
  Merchant,
  MerchantStats,
  MerchantSettlement,
  MerchantOrder,
  MerchantAPIKey,
  APIKeyUsage,
  PaginatedResponse,
  APIResponse,
} from '@/types';
import type { ModelProvider } from '@/types/sku';

export const merchantService = {
  registerMerchant: (data: {
    company_name: string;
    business_license?: string;
    contact_name?: string;
    contact_phone?: string;
    contact_email?: string;
    address?: string;
    description?: string;
  }) => api.post<Merchant>('/merchants/register', data),

  getProfile: () => api.get<Merchant>('/merchants/profile'),

  updateProfile: (data: Partial<Merchant>) => api.put<Merchant>('/merchants/profile', data),

  getStats: () => api.get<MerchantStats>('/merchants/stats'),

  getOrders: (page?: number, perPage?: number, status?: string) =>
    api.get<APIResponse<PaginatedResponse<MerchantOrder>>>('/merchants/orders', {
      params: { page, per_page: perPage, status },
    }),

  getSettlements: () => api.get<APIResponse<MerchantSettlement[]>>('/merchants/settlements'),

  requestSettlement: () => api.post<MerchantSettlement>('/merchants/settlements'),

  getSettlementDetail: (id: number) => api.get<MerchantSettlement>(`/merchants/settlements/${id}`),

  getAPIKeys: () => api.get<APIResponse<MerchantAPIKey[]>>('/merchants/api-keys'),

  createAPIKey: (data: {
    name: string;
    provider: string;
    api_key: string;
    api_secret?: string;
    quota_limit?: number | null;
    health_check_level?: MerchantAPIKey['health_check_level'];
    endpoint_url?: string;
  }) => api.post<MerchantAPIKey>('/merchants/api-keys', data),

  updateAPIKey: (id: number, data: Partial<MerchantAPIKey>) =>
    api.put<MerchantAPIKey>(`/merchants/api-keys/${id}`, data),

  deleteAPIKey: (id: number) => api.delete(`/merchants/api-keys/${id}`),

  getAPIKeyUsage: () => api.get<APIResponse<APIKeyUsage[]>>('/merchants/api-keys/usage'),

  submitDocuments: (data: {
    business_license_url: string;
    id_card_front_url?: string;
    id_card_back_url?: string;
    attachments?: string;
    contact_name?: string;
    contact_phone?: string;
    contact_email?: string;
    address?: string;
  }) => api.post<Merchant>('/merchants/documents', data),

  getMerchantStatus: () =>
    api.get<{
      status: string;
      can_submit: boolean;
      rejection_reason?: string;
    }>('/merchants/status'),

  /** Active model_providers only; same data as admin dropdown, merchant role required */
  getMerchantModelProviders: () => api.get<{ data: ModelProvider[] }>('/merchants/model-providers'),
};
