import api from './api';

export interface PendingMerchant {
  id: number;
  user_id: number;
  company_name: string;
  business_license?: string;
  business_license_url?: string;
  id_card_front_url?: string;
  id_card_back_url?: string;
  contact_name: string;
  contact_phone: string;
  contact_email: string;
  address: string;
  description: string;
  business_category?: string;
  admin_notes?: string;
  reviewed_at?: string;
  status: string;
  rejection_reason?: string;
  created_at: string;
  updated_at: string;
  user_email?: string;
  user_name?: string;
}

export interface MerchantListResponse {
  data: PendingMerchant[];
  total: number;
  page: number;
  per_page: number;
}

export interface MerchantAuditLog {
  id: number;
  merchant_id: number;
  admin_user_id?: number;
  admin_email?: string;
  action: string;
  company_name_snapshot?: string;
  reason?: string;
  created_at: string;
}

export interface AuditLogListResponse {
  data: MerchantAuditLog[];
  total: number;
  page: number;
  per_page: number;
}

/** 经营类目（与后端存储一致，可扩展） */
export const MERCHANT_BUSINESS_CATEGORY_OPTIONS = [
  { value: 'retail', label: '零售电商' },
  { value: 'food', label: '餐饮/食品' },
  { value: 'service', label: '本地服务' },
  { value: 'digital', label: '数字/软件' },
  { value: 'other', label: '其他' },
];

export function labelForBusinessCategory(value?: string): string {
  if (!value) return '-';
  const opt = MERCHANT_BUSINESS_CATEGORY_OPTIONS.find((o) => o.value === value);
  return opt?.label ?? value;
}

export const adminMerchantService = {
  getPendingMerchants: (page = 1, perPage = 20) =>
    api.get<MerchantListResponse>('/admin/merchants/pending', {
      params: { page, per_page: perPage },
    }),

  getAdminMerchants: (params: {
    page?: number;
    per_page?: number;
    status?: string;
    business_category?: string;
    keyword?: string;
  }) => api.get<MerchantListResponse>('/admin/merchants', { params }),

  getMerchantAuditLogs: (params: {
    page?: number;
    per_page?: number;
    merchant_id?: number;
    action?: string;
  }) => api.get<AuditLogListResponse>('/admin/merchants/audit-logs', { params }),

  approveMerchant: (merchantId: number) =>
    api.post<{ code: number; message: string }>(`/admin/merchants/${merchantId}/approve`, {}),

  rejectMerchant: (merchantId: number, reason?: string) =>
    api.post<{ code: number; message: string }>(`/admin/merchants/${merchantId}/reject`, {
      reason: reason ?? '',
    }),

  patchMerchant: (merchantId: number, body: { business_category?: string; admin_notes?: string }) =>
    api.patch<{ code: number; message: string; data: PendingMerchant }>(
      `/admin/merchants/${merchantId}`,
      body
    ),
};
