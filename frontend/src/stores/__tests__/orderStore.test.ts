import { useOrderStore } from '../orderStore'
import { orderService } from '@/services/order'

// 模拟orderService
jest.mock('@/services/order', () => ({
  orderService: {
    listOrders: jest.fn(),
    getOrderByID: jest.fn(),
    createOrder: jest.fn(),
    cancelOrder: jest.fn(),
  },
}))

const mockOrderService = orderService as jest.Mocked<typeof orderService>

describe('orderStore', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // 重置store状态
    useOrderStore.setState({
      orders: [],
      currentOrder: null,
      isLoading: false,
      error: null,
    })
  })

  test('fetchOrders 成功获取订单列表', async () => {
    const mockOrders = [
      { id: 1, order_id: 'ORD123', amount: 100, status: 'completed', product_id: 1, quantity: 2 },
      { id: 2, order_id: 'ORD124', amount: 200, status: 'pending', product_id: 2, quantity: 1 },
    ]

    mockOrderService.listOrders.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: {
          data: mockOrders,
          pagination: { total: 2, page: 1, per_page: 20 },
        },
      },
    })

    const store = useOrderStore.getState()
    await store.fetchOrders()

    const newState = useOrderStore.getState()
    expect(newState.orders).toEqual(mockOrders)
    expect(newState.isLoading).toBe(false)
    expect(newState.error).toBe(null)
  })

  test('fetchOrders 获取订单列表失败', async () => {
    const errorMessage = '获取订单列表失败'
    mockOrderService.listOrders.mockRejectedValue(new Error(errorMessage))

    const store = useOrderStore.getState()
    await store.fetchOrders()

    const newState = useOrderStore.getState()
    expect(newState.orders).toEqual([])
    expect(newState.isLoading).toBe(false)
    expect(newState.error).toBe(errorMessage)
  })

  test('fetchOrderByID 成功获取订单详情', async () => {
    const mockOrder = {
      id: 1,
      order_id: 'ORD123',
      amount: 100,
      status: 'completed',
      product_id: 1,
      quantity: 2,
      created_at: new Date().toISOString(),
    }

    mockOrderService.getOrderByID.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: mockOrder,
      },
    })

    const store = useOrderStore.getState()
    await store.fetchOrderByID(1)

    const newState = useOrderStore.getState()
    expect(newState.currentOrder).toEqual(mockOrder)
    expect(newState.isLoading).toBe(false)
    expect(newState.error).toBe(null)
  })

  test('fetchOrderByID 获取订单详情失败', async () => {
    const errorMessage = '获取订单详情失败'
    mockOrderService.getOrderByID.mockRejectedValue(new Error(errorMessage))

    const store = useOrderStore.getState()
    await store.fetchOrderByID(1)

    const newState = useOrderStore.getState()
    expect(newState.currentOrder).toBe(null)
    expect(newState.isLoading).toBe(false)
    expect(newState.error).toBe(errorMessage)
  })

  test('createOrder 成功创建订单', async () => {
    const mockOrder = {
      id: 3,
      order_id: 'ORD125',
      amount: 150,
      status: 'pending',
      product_id: 1,
      quantity: 1,
      group_id: 1,
    }

    mockOrderService.createOrder.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: mockOrder,
      },
    })

    const store = useOrderStore.getState()
    await store.createOrder(1, 1, 1)

    const newState = useOrderStore.getState()
    expect(newState.orders).toContainEqual(mockOrder)
    expect(newState.orders[0]).toEqual(mockOrder)
    expect(newState.currentOrder).toEqual(mockOrder)
    expect(newState.isLoading).toBe(false)
    expect(newState.error).toBe(null)
  })

  test('createOrder 创建订单失败', async () => {
    const errorMessage = '创建订单失败'
    mockOrderService.createOrder.mockRejectedValue(new Error(errorMessage))

    const store = useOrderStore.getState()
    await expect(store.createOrder(1, 1)).rejects.toThrow(errorMessage)

    const newState = useOrderStore.getState()
    expect(newState.isLoading).toBe(false)
    expect(newState.error).toBe(errorMessage)
  })

  test('cancelOrder 成功取消订单', async () => {
    const initialOrders = [
      { id: 1, order_id: 'ORD123', amount: 100, status: 'pending', product_id: 1, quantity: 2 },
      { id: 2, order_id: 'ORD124', amount: 200, status: 'completed', product_id: 2, quantity: 1 },
    ]

    const cancelledOrder = {
      id: 1,
      order_id: 'ORD123',
      amount: 100,
      status: 'cancelled',
      product_id: 1,
      quantity: 2,
    }

    // 设置初始订单状态
    useOrderStore.setState({ orders: initialOrders })

    mockOrderService.cancelOrder.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: cancelledOrder,
      },
    })

    const store = useOrderStore.getState()
    await store.cancelOrder(1)

    const newState = useOrderStore.getState()
    expect(newState.orders).toHaveLength(2)
    expect(newState.orders[0]).toEqual(cancelledOrder)
    expect(newState.orders[1]).toEqual(initialOrders[1])
    expect(newState.isLoading).toBe(false)
    expect(newState.error).toBe(null)
  })

  test('cancelOrder 取消订单失败', async () => {
    const errorMessage = '取消订单失败'
    mockOrderService.cancelOrder.mockRejectedValue(new Error(errorMessage))

    const store = useOrderStore.getState()
    await expect(store.cancelOrder(1)).rejects.toThrow(errorMessage)

    const newState = useOrderStore.getState()
    expect(newState.isLoading).toBe(false)
    expect(newState.error).toBe(errorMessage)
  })

  test('clearError 清除错误信息', () => {
    // 先设置一个错误
    useOrderStore.setState({ error: '测试错误' })
    
    const store = useOrderStore.getState()
    store.clearError()
    
    const newState = useOrderStore.getState()
    expect(newState.error).toBe(null)
  })
})