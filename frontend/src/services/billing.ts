import api from './api';
import type { BillingRecord, BillingStats, BillingFilter } from '@/types';

export interface BillingListResponse {
  billings: BillingRecord[];
  total: number;
  page: number;
  page_size: number;
}

export interface BillingTrend {
  date: string;
  total_cost: number;
  total_tokens: number;
  total_requests: number;
  avg_latency: number;
}

export interface ProviderBreakdown {
  provider: string;
  count: number;
  cost: number;
  percentage: number;
}

export const billingService = {
  getBillings: (params?: BillingFilter & { page?: number; page_size?: number }) =>
    api.get<BillingListResponse>('/admin/billings', { params }),

  getBillingStats: (params?: BillingFilter) =>
    api.get<BillingStats>('/admin/billings/stats', { params }),

  getBillingTrends: (params?: BillingFilter & { granularity?: 'day' | 'week' | 'month' }) =>
    api.get<{ trends: BillingTrend[] }>('/admin/billings/trends', { params }),

  exportBillingsCSV: (params?: BillingFilter) =>
    api.get<Blob>('/admin/billings/export', { 
      params,
      responseType: 'blob'
    }),

  getUserBillings: (params?: BillingFilter & { page?: number; page_size?: number }) =>
    api.get<BillingListResponse>('/admin/user-billings', { params }),

  getUserBillingStats: (params?: BillingFilter) =>
    api.get<BillingStats>('/admin/user-billings/stats', { params }),

  exportUserBillingsCSV: (params?: BillingFilter) =>
    api.get<Blob>('/admin/user-billings/export', { 
      params,
      responseType: 'blob'
    }),
};
