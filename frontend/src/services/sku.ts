import api from './api';
import type {
  SPU,
  SKU,
  SKUWithSPU,
  ModelProvider,
  ComputePointTransaction,
  UserSubscriptionWithSKU,
  SPUCreateRequest,
  SKUCreateRequest,
  SKUUpdateRequest,
} from '@/types/sku';

export const skuService = {
  getSPUs: (params?: {
    page?: number;
    per_page?: number;
    provider?: string;
    tier?: string;
    status?: string;
  }) =>
    api.get<{ total: number; page: number; per_page: number; data: SPU[] }>('/admin/spus', {
      params,
    }),

  getSPU: (id: number) => api.get<{ data: SPU }>(`/admin/spus/${id}`),

  createSPU: (data: SPUCreateRequest) => api.post<{ data: SPU }>('/admin/spus', data),

  updateSPU: (id: number, data: Partial<SPUCreateRequest>) =>
    api.put<{ data: SPU }>(`/admin/spus/${id}`, data),

  deleteSPU: (id: number) => api.delete(`/admin/spus/${id}`),

  getSPUScenarios: (id: number) =>
    api.get<{
      scenarios: Array<{
        id: number;
        code: string;
        name: string;
        description?: string;
        is_linked: boolean;
        is_primary: boolean;
      }>;
    }>(`/admin/spus/${id}/scenarios`),

  updateSPUScenarios: (id: number, data: { scenario_ids: number[]; primary_id?: number }) =>
    api.put<{ message: string }>(`/admin/spus/${id}/scenarios`, data),

  getSKUs: (params?: {
    page?: number;
    per_page?: number;
    spu_id?: number;
    type?: string;
    status?: string;
  }) =>
    api.get<{ total: number; page: number; per_page: number; data: SKUWithSPU[] }>('/admin/skus', {
      params,
    }),

  getSKU: (id: number) => api.get<{ data: SKUWithSPU }>(`/admin/skus/${id}`),

  createSKU: (data: SKUCreateRequest) => api.post<{ data: SKU }>('/admin/skus', data),

  updateSKU: (id: number, data: SKUUpdateRequest) =>
    api.put<{ data: SKU }>(`/admin/skus/${id}`, data),

  deleteSKU: (id: number) => api.delete(`/admin/skus/${id}`),

  getModelProviders: () => api.get<{ data: ModelProvider[] }>('/admin/model-providers'),

  /** 全部厂商（含停用），供管理端维护页使用 */
  getAllModelProviders: () => api.get<{ data: ModelProvider[] }>('/admin/model-providers/all'),

  createModelProvider: (data: {
    code: string;
    name: string;
    api_base_url?: string;
    api_format?: string;
    billing_type?: string;
    status?: string;
    sort_order?: number;
  }) => api.post<{ data: ModelProvider }>('/admin/model-providers', data),

  patchModelProvider: (
    id: number,
    data: Partial<{
      name: string;
      api_base_url: string;
      api_format: string;
      billing_type: string;
      status: string;
      sort_order: number;
    }>
  ) => api.patch<{ data: ModelProvider }>(`/admin/model-providers/${id}`, data),

  getComputePointBalance: () =>
    api.get<{
      data: {
        balance: number;
        total_earned: number;
        total_used: number;
        total_expired: number;
      };
    }>('/compute-points/balance'),

  getComputePointTransactions: (params?: { page?: number; per_page?: number }) =>
    api.get<{ total: number; page: number; per_page: number; data: ComputePointTransaction[] }>(
      '/compute-points/transactions',
      { params }
    ),

  getUserSubscriptions: () => api.get<{ data: UserSubscriptionWithSKU[] }>('/subscriptions'),

  getPublicSKUs: (params?: {
    page?: number;
    per_page?: number;
    spu_id?: number;
    type?: string;
    q?: string;
    search?: string;
    provider?: string;
    tier?: string;
    model_name?: string;
    category?: string;
    group_enabled?: boolean | string;
    price_min?: number;
    price_max?: number;
    valid_days_min?: number;
    valid_days_max?: number;
    sort?: 'hot' | 'new' | 'price_asc' | 'price_desc';
    scenario?: string;
  }) =>
    api.get<{ total: number; page: number; per_page: number; data: SKUWithSPU[] }>('/skus', {
      params,
    }),

  getPublicSKU: (id: number) => api.get<{ data: SKUWithSPU }>(`/skus/${id}`),
};

export default skuService;
