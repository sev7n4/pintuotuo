import { orderService } from '../order';
import api from '../api';
import type { AxiosResponse } from 'axios';

jest.mock('../api');

const mockApi = api as jest.Mocked<typeof api>;

const createMockResponse = <T>(data: T): AxiosResponse<T> => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
});

describe('orderService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('createOrder calls api.post with correct parameters', async () => {
    const mockData = {
      items: [{ sku_id: 1, quantity: 2 }],
    };
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        product_id: 1,
        user_id: 1,
        quantity: 2,
        total_amount: 100,
        status: 'pending',
        payment_method: 'alipay',
        payment_status: 'unpaid',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Order created successfully',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await orderService.createOrder(mockData);

    expect(mockApi.post).toHaveBeenCalledWith('/orders', mockData);
    expect(result.data).toEqual(mockResponse);
  });

  test('createOrder calls api.post with group_id when provided', async () => {
    const mockData = {
      items: [{ sku_id: 1, quantity: 2 }],
    };
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        product_id: 1,
        group_id: 1,
        user_id: 1,
        quantity: 2,
        total_amount: 100,
        status: 'pending',
        payment_method: 'alipay',
        payment_status: 'unpaid',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Order created successfully',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await orderService.createOrder(mockData);

    expect(mockApi.post).toHaveBeenCalledWith('/orders', mockData);
    expect(result.data).toEqual(mockResponse);
  });

  test('listOrders calls api.get with correct parameters', async () => {
    const mockPage = 1;
    const mockPerPage = 10;
    const mockResponse = {
      success: true,
      data: {
        items: [
          {
            id: 1,
            product_id: 1,
            user_id: 1,
            quantity: 2,
            total_amount: 100,
            status: 'pending',
            payment_method: 'alipay',
            payment_status: 'unpaid',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        ],
        total: 1,
        page: 1,
        per_page: 10,
      },
      message: 'Orders retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await orderService.listOrders(mockPage, mockPerPage);

    expect(mockApi.get).toHaveBeenCalledWith('/orders', {
      params: { page: mockPage, per_page: mockPerPage },
    });
    expect(result.data).toEqual(mockResponse);
  });

  test('listOrders calls api.get without parameters when page and per_page are not provided', async () => {
    const mockResponse = {
      success: true,
      data: {
        items: [],
        total: 0,
        page: 1,
        per_page: 10,
      },
      message: 'Orders retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await orderService.listOrders();

    expect(mockApi.get).toHaveBeenCalledWith('/orders', {
      params: { page: undefined, per_page: undefined },
    });
    expect(result.data).toEqual(mockResponse);
  });

  test('getOrderByID calls api.get with correct parameters', async () => {
    const mockOrderId = 1;
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        product_id: 1,
        user_id: 1,
        quantity: 2,
        total_amount: 100,
        status: 'pending',
        payment_method: 'alipay',
        payment_status: 'unpaid',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Order retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await orderService.getOrderByID(mockOrderId);

    expect(mockApi.get).toHaveBeenCalledWith(`/orders/${mockOrderId}`);
    expect(result.data).toEqual(mockResponse);
  });

  test('cancelOrder calls api.put with correct parameters', async () => {
    const mockOrderId = 1;
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        product_id: 1,
        user_id: 1,
        quantity: 2,
        total_amount: 100,
        status: 'cancelled',
        payment_method: 'alipay',
        payment_status: 'unpaid',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Order cancelled successfully',
    };

    mockApi.put.mockResolvedValue(createMockResponse(mockResponse));

    const result = await orderService.cancelOrder(mockOrderId);

    expect(mockApi.put).toHaveBeenCalledWith(`/orders/${mockOrderId}/cancel`, {});
    expect(result.data).toEqual(mockResponse);
  });
});
