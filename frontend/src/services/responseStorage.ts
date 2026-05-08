import api from './api';

export interface ResponseStorageItem {
  id: number;
  response_id: string;
  user_id: number;
  merchant_id: number;
  model: string;
  status: string;
  error_message?: string;
  background_job_id?: string;
  created_at: string;
  expires_at: string;
}

export const responseStorageService = {
  getList: (params?: { page?: number; per_page?: number; status?: string; user_id?: string }) =>
    api.get<{ total: number; page: number; per_page: number; data: ResponseStorageItem[] }>('/admin/response-storage', { params }),

  delete: (id: number) =>
    api.delete(`/admin/response-storage/${id}`),

  cleanExpired: () =>
    api.post<{ message: string; deleted_count: number }>('/admin/response-storage/clean-expired'),
};
