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

export const adminMerchantService = {
  getPendingMerchants: (page = 1, perPage = 20) =>
    api.get<MerchantListResponse>('/admin/merchants/pending', {
      params: { page, per_page: perPage },
    }),

  approveMerchant: (merchantId: number) =>
    api.post<{ code: number; message: string }>(`/admin/merchants/${merchantId}/approve`, {}),

  rejectMerchant: (merchantId: number) =>
    api.post<{ code: number; message: string }>(`/admin/merchants/${merchantId}/reject`, {}),
};
