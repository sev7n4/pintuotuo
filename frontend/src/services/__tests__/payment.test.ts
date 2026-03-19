import { paymentService } from '../payment'
import api from '../api'

// 模拟 api
jest.mock('../api')

const mockApi = api as jest.Mocked<typeof api>

describe('paymentService', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('initiatePayment calls api.post with correct parameters', async () => {
    const mockData = {
      order_id: 1,
      method: 'alipay' as const,
    }
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        order_id: 1,
        method: 'alipay',
        amount: 100,
        status: 'pending',
        transaction_id: 'test-transaction-id',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Payment initiated successfully',
    }

    mockApi.post.mockResolvedValue(mockResponse)

    const result = await paymentService.initiatePayment(mockData)

    expect(mockApi.post).toHaveBeenCalledWith('/payments', mockData)
    expect(result).toEqual(mockResponse)
  })

  test('initiatePayment calls api.post with wechat method', async () => {
    const mockData = {
      order_id: 1,
      method: 'wechat' as const,
    }
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        order_id: 1,
        method: 'wechat',
        amount: 100,
        status: 'pending',
        transaction_id: 'test-transaction-id',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Payment initiated successfully',
    }

    mockApi.post.mockResolvedValue(mockResponse)

    const result = await paymentService.initiatePayment(mockData)

    expect(mockApi.post).toHaveBeenCalledWith('/payments', mockData)
    expect(result).toEqual(mockResponse)
  })

  test('getPaymentByID calls api.get with correct parameters', async () => {
    const mockPaymentId = 1
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        order_id: 1,
        method: 'alipay',
        amount: 100,
        status: 'completed',
        transaction_id: 'test-transaction-id',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Payment retrieved successfully',
    }

    mockApi.get.mockResolvedValue(mockResponse)

    const result = await paymentService.getPaymentByID(mockPaymentId)

    expect(mockApi.get).toHaveBeenCalledWith(`/payments/${mockPaymentId}`)
    expect(result).toEqual(mockResponse)
  })

  test('refundPayment calls api.post with correct parameters', async () => {
    const mockPaymentId = 1
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        order_id: 1,
        method: 'alipay',
        amount: 100,
        status: 'refunded',
        transaction_id: 'test-transaction-id',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Payment refunded successfully',
    }

    mockApi.post.mockResolvedValue(mockResponse)

    const result = await paymentService.refundPayment(mockPaymentId)

    expect(mockApi.post).toHaveBeenCalledWith(`/payments/${mockPaymentId}/refund`, {})
    expect(result).toEqual(mockResponse)
  })

  test('handleAlipayCallback calls api.post with correct parameters', async () => {
    const mockData = {
      payment_id: 1,
      status: 'completed',
      amount: 100,
    }
    const mockResponse = {
      success: true,
      data: undefined,
      message: 'Callback handled successfully',
    }

    mockApi.post.mockResolvedValue(mockResponse)

    const result = await paymentService.handleAlipayCallback(mockData)

    expect(mockApi.post).toHaveBeenCalledWith('/payments/webhooks/alipay', mockData)
    expect(result).toEqual(mockResponse)
  })

  test('handleWechatCallback calls api.post with correct parameters', async () => {
    const mockData = {
      payment_id: 1,
      status: 'completed',
      amount: 100,
    }
    const mockResponse = {
      success: true,
      data: undefined,
      message: 'Callback handled successfully',
    }

    mockApi.post.mockResolvedValue(mockResponse)

    const result = await paymentService.handleWechatCallback(mockData)

    expect(mockApi.post).toHaveBeenCalledWith('/payments/webhooks/wechat', mockData)
    expect(result).toEqual(mockResponse)
  })
})
