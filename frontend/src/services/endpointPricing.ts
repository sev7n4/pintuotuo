import api from './api';
import type { EndpointPricing, EndpointPricingCreateRequest, EndpointPricingUpdateRequest } from '@/types/sku';

export const endpointPricingService = {
  getList: (params?: { page?: number; per_page?: number; endpoint_type?: string; provider_code?: string }) =>
    api.get<{ total: number; page: number; per_page: number; data: EndpointPricing[] }>('/admin/endpoint-pricing', { params }),

  create: (data: EndpointPricingCreateRequest) =>
    api.post<{ data: EndpointPricing }>('/admin/endpoint-pricing', data),

  update: (id: number, data: EndpointPricingUpdateRequest) =>
    api.put<{ data: EndpointPricing }>(`/admin/endpoint-pricing/${id}`, data),

  delete: (id: number) =>
    api.delete(`/admin/endpoint-pricing/${id}`),
};
