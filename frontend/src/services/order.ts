import api from './api'
import { Order, APIResponse, PaginatedResponse } from '@types/index'

interface CreateOrderRequest {
  product_id: number
  group_id?: number
  quantity: number
}

export const orderService = {
  // Create order
  createOrder: (data: CreateOrderRequest) =>
    api.post<APIResponse<Order>>('/orders', data),

  // List user orders
  listOrders: (page?: number, per_page?: number) =>
    api.get<APIResponse<PaginatedResponse<Order>>>('/orders', {
      params: { page, per_page },
    }),

  // Get order by ID
  getOrderByID: (id: number) =>
    api.get<APIResponse<Order>>(`/orders/${id}`),

  // Cancel order
  cancelOrder: (id: number) =>
    api.put<APIResponse<Order>>(`/orders/${id}/cancel`, {}),
}
