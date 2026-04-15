import { create } from 'zustand';
import { Order, APIResponse, PaginatedResponse } from '@/types';
import { orderService } from '@/services/order';
import { getApiErrorMessage } from '@/utils/apiError';

interface OrderState {
  orders: Order[];
  currentOrder: Order | null;
  isLoading: boolean;
  error: string | null;

  fetchOrders: (page?: number, per_page?: number) => Promise<void>;
  fetchOrderByID: (id: number) => Promise<void>;
  createOrder: (items: Array<{ sku_id: number; quantity: number }>) => Promise<number | null>;
  cancelOrder: (id: number, reason?: string) => Promise<void>;
  requestRefund: (id: number, reason: string) => Promise<void>;
  clearError: () => void;
}

export const useOrderStore = create<OrderState>((set) => ({
  orders: [],
  currentOrder: null,
  isLoading: false,
  error: null,

  fetchOrders: async (page = 1, per_page = 20) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orderService.listOrders(page, per_page);
      const apiResponse = response.data as APIResponse<PaginatedResponse<Order>>;
      set({ orders: apiResponse.data?.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取订单列表失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchOrderByID: async (id) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orderService.getOrderByID(id);
      const apiResponse = response.data as APIResponse<Order>;
      set({ currentOrder: apiResponse.data || null, isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取订单详情失败';
      set({ error: message, isLoading: false });
    }
  },

  createOrder: async (items) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orderService.createOrder({ items });
      const apiResponse = response.data as APIResponse<Order>;
      const newOrder = apiResponse.data;
      if (newOrder) {
        set((state) => ({
          orders: [newOrder, ...state.orders],
          currentOrder: newOrder,
          isLoading: false,
        }));
        return newOrder.id;
      }
      set({ isLoading: false });
      return null;
    } catch (error) {
      const msg = getApiErrorMessage(error);
      set({ error: msg, isLoading: false });
      throw new Error(msg);
    }
  },

  cancelOrder: async (id, reason) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orderService.cancelOrder(id, reason);
      const apiResponse = response.data as APIResponse<Order>;
      const cancelledOrder = apiResponse.data;
      if (cancelledOrder) {
        set((state) => ({
          orders: state.orders.map((order) => (order.id === id ? cancelledOrder : order)),
          isLoading: false,
        }));
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : '取消订单失败';
      set({ error: message, isLoading: false });
      throw error;
    }
  },

  requestRefund: async (id, reason) => {
    set({ isLoading: true, error: null });
    try {
      const response = await orderService.requestRefund(id, reason);
      const apiResponse = response.data as APIResponse<Order>;
      const refundedOrder = apiResponse.data;
      if (refundedOrder) {
        set((state) => ({
          orders: state.orders.map((order) => (order.id === id ? refundedOrder : order)),
          isLoading: false,
        }));
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : '退款申请失败';
      set({ error: message, isLoading: false });
      throw error;
    }
  },

  clearError: () => set({ error: null }),
}));
