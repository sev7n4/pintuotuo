import api from './api'
import { Order, APIResponse, PaginatedResponse } from '@/types'

interface CreateOrderRequest {
  product_id: number
  group_id?: number
  quantity: number
}

interface CancelOrderRequest {
  reason?: string
}

export const orderService = {
  createOrder: (data: CreateOrderRequest) =>
    api.post<APIResponse<Order>>('/orders', data),

  listOrders: (page?: number, per_page?: number) =>
    api.get<APIResponse<PaginatedResponse<Order>>>('/orders', {
      params: { page, per_page },
    }),

  getOrderByID: (id: number) =>
    api.get<APIResponse<Order>>(`/orders/${id}`),

  cancelOrder: (id: number, reason?: string) =>
    api.put<APIResponse<Order>>(`/orders/${id}/cancel`, { reason } as CancelOrderRequest),

  requestRefund: (id: number, reason: string) =>
    api.post<APIResponse<Order>>(`/orders/${id}/refund`, { reason }),
}
