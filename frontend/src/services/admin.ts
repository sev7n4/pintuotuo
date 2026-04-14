import api from './api';
import { APIResponse } from '@/types';

export interface AdminStats {
  total_users: number;
  total_merchants: number;
  total_orders: number;
  total_revenue: number;
  pending_orders: number;
  paid_orders: number;
  cancelled_orders: number;
  multi_item_order_ratio: number;
  order_conversion_rate: number;
  payment_success_rate: number;
  cancellation_rate: number;
}

export const adminService = {
  getStats: () => api.get<APIResponse<AdminStats>>('/admin/stats'),
};
