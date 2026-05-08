import api from './api';
import { APIResponse } from '@/types';

export interface AdminStats {
  total_users: number;
  total_merchants: number;
  total_orders: number;
  total_revenue: number;
  pending_orders: number;
  paid_orders: number;
  canceled_orders: number;
  multi_item_order_ratio: number;
  order_conversion_rate: number;
  payment_success_rate: number;
  cancellation_rate: number;
  by_endpoint_type?: { endpoint_type: string; count: number; tokens: number }[];
}

export interface ProbeEndpointResponse {
  success: boolean;
  status_code: number;
  latency_ms: number;
  error_msg?: string;
  error_code?: string;
}

export const adminService = {
  getStats: () => api.get<APIResponse<AdminStats>>('/admin/stats'),

  probeEndpoint: (providerCode: string, url: string, apiKey?: string, timeoutMs?: number) =>
    api.post<APIResponse<ProbeEndpointResponse>>(
      `/admin/route-configs/providers/${providerCode}/probe-endpoint`,
      {
        url,
        api_key: apiKey,
        timeout_ms: timeoutMs,
      }
    ),
};
