import api from './api';
import { Order, APIResponse, PaginatedResponse } from '@/types';

interface CreateOrderRequest {
  items: Array<{
    sku_id: number;
    quantity: number;
    flash_sale_id?: number;
  }>;
  /** 与 items 明细一致时写入订单，用于套餐销量统计 */
  entitlement_package_id?: number;
}

interface CancelOrderRequest {
  reason?: string;
}

export const orderService = {
  createOrder: (data: CreateOrderRequest) => api.post<APIResponse<Order>>('/orders', data),

  listOrders: (page?: number, per_page?: number) =>
    api.get<APIResponse<PaginatedResponse<Order>>>('/orders', {
      params: { page, per_page },
    }),

  getOrderByID: (id: number) => api.get<APIResponse<Order>>(`/orders/${id}`),

  cancelOrder: (id: number, reason?: string) =>
    api.put<APIResponse<Order>>(`/orders/${id}/cancel`, { reason } as CancelOrderRequest),

  requestRefund: (id: number, reason: string) =>
    api.post<APIResponse<Order>>(`/orders/${id}/refund`, { reason }),
};
