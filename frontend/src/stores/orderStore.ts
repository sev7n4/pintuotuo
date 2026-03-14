import { create } from 'zustand'
import { Order } from '@types/index'
import { orderService } from '@services/order'

interface OrderState {
  orders: Order[]
  currentOrder: Order | null
  isLoading: boolean
  error: string | null

  // Actions
  fetchOrders: (page?: number, per_page?: number) => Promise<void>
  fetchOrderByID: (id: number) => Promise<void>
  createOrder: (productId: number, quantity: number, groupId?: number) => Promise<void>
  cancelOrder: (id: number) => Promise<void>
  clearError: () => void
}

export const useOrderStore = create<OrderState>((set) => ({
  orders: [],
  currentOrder: null,
  isLoading: false,
  error: null,

  fetchOrders: async (page = 1, per_page = 20) => {
    set({ isLoading: true, error: null })
    try {
      const response = await orderService.listOrders(page, per_page)
      set({ orders: response.data.data, isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取订单列表失败'
      set({ error: message, isLoading: false })
    }
  },

  fetchOrderByID: async (id) => {
    set({ isLoading: true, error: null })
    try {
      const response = await orderService.getOrderByID(id)
      set({ currentOrder: response.data, isLoading: false })
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取订单详情失败'
      set({ error: message, isLoading: false })
    }
  },

  createOrder: async (productId, quantity, groupId) => {
    set({ isLoading: true, error: null })
    try {
      const response = await orderService.createOrder({
        product_id: productId,
        quantity,
        group_id: groupId,
      })
      set((state) => ({
        orders: [response.data, ...state.orders],
        currentOrder: response.data,
        isLoading: false,
      }))
    } catch (error) {
      const message = error instanceof Error ? error.message : '创建订单失败'
      set({ error: message, isLoading: false })
      throw error
    }
  },

  cancelOrder: async (id) => {
    set({ isLoading: true, error: null })
    try {
      const response = await orderService.cancelOrder(id)
      set((state) => ({
        orders: state.orders.map((order) =>
          order.id === id ? response.data : order
        ),
        isLoading: false,
      }))
    } catch (error) {
      const message = error instanceof Error ? error.message : '取消订单失败'
      set({ error: message, isLoading: false })
      throw error
    }
  },

  clearError: () => set({ error: null }),
}))
