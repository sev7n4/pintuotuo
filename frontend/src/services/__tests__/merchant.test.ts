import { merchantService } from '../merchant';
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

describe('merchantService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('registerMerchant calls api.post with correct parameters', async () => {
    const mockData = {
      company_name: 'Test Company',
      contact_name: 'Test Contact',
      contact_phone: '13800138000',
      contact_email: 'test@example.com',
    };
    const mockResponse = {
      id: 1,
      user_id: 1,
      company_name: 'Test Company',
      contact_name: 'Test Contact',
      contact_phone: '13800138000',
      contact_email: 'test@example.com',
      status: 'pending',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.registerMerchant(mockData);

    expect(mockApi.post).toHaveBeenCalledWith('/merchants/register', mockData);
    expect(result.data).toEqual(mockResponse);
  });

  test('getProfile calls api.get with correct parameters', async () => {
    const mockResponse = {
      id: 1,
      user_id: 1,
      company_name: 'Test Company',
      contact_name: 'Test Contact',
      contact_phone: '13800138000',
      contact_email: 'test@example.com',
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.getProfile();

    expect(mockApi.get).toHaveBeenCalledWith('/merchants/profile');
    expect(result.data).toEqual(mockResponse);
  });

  test('updateProfile calls api.put with correct parameters', async () => {
    const mockData = {
      contact_name: 'Updated Contact',
      contact_phone: '13900139000',
    };
    const mockResponse = {
      id: 1,
      user_id: 1,
      company_name: 'Test Company',
      contact_name: 'Updated Contact',
      contact_phone: '13900139000',
      contact_email: 'test@example.com',
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockApi.put.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.updateProfile(mockData);

    expect(mockApi.put).toHaveBeenCalledWith('/merchants/profile', mockData);
    expect(result.data).toEqual(mockResponse);
  });

  test('getStats calls api.get with correct parameters', async () => {
    const mockResponse = {
      total_orders: 100,
      total_sales: 10000,
      total_products: 50,
      pending_orders: 5,
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.getStats();

    expect(mockApi.get).toHaveBeenCalledWith('/merchants/stats');
    expect(result.data).toEqual(mockResponse);
  });

  test('getProducts calls api.get with correct parameters', async () => {
    const mockPage = 1;
    const mockPerPage = 10;
    const mockStatus = 'active';
    const mockResponse = {
      success: true,
      data: {
        items: [
          {
            id: 1,
            name: 'Test Product',
            price: 100,
            stock: 100,
            status: 'active',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        ],
        total: 1,
        page: 1,
        per_page: 10,
      },
      message: 'Products retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.getProducts(mockPage, mockPerPage, mockStatus);

    expect(mockApi.get).toHaveBeenCalledWith('/merchants/products', {
      params: { page: mockPage, per_page: mockPerPage, status: mockStatus },
    });
    expect(result.data).toEqual(mockResponse);
  });

  test('getOrders calls api.get with correct parameters', async () => {
    const mockPage = 1;
    const mockPerPage = 10;
    const mockStatus = 'pending';
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

    const result = await merchantService.getOrders(mockPage, mockPerPage, mockStatus);

    expect(mockApi.get).toHaveBeenCalledWith('/merchants/orders', {
      params: { page: mockPage, per_page: mockPerPage, status: mockStatus },
    });
    expect(result.data).toEqual(mockResponse);
  });

  test('getSettlements calls api.get with correct parameters', async () => {
    const mockResponse = {
      success: true,
      data: [
        {
          id: 1,
          merchant_id: 1,
          amount: 5000,
          status: 'pending',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ],
      message: 'Settlements retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.getSettlements();

    expect(mockApi.get).toHaveBeenCalledWith('/merchants/settlements');
    expect(result.data).toEqual(mockResponse);
  });

  test('requestSettlement calls api.post with correct parameters', async () => {
    const mockResponse = {
      id: 1,
      merchant_id: 1,
      amount: 5000,
      status: 'pending',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.requestSettlement();

    expect(mockApi.post).toHaveBeenCalledWith('/merchants/settlements');
    expect(result.data).toEqual(mockResponse);
  });

  test('getSettlementDetail calls api.get with correct parameters', async () => {
    const mockSettlementId = 1;
    const mockResponse = {
      id: 1,
      merchant_id: 1,
      amount: 5000,
      status: 'pending',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.getSettlementDetail(mockSettlementId);

    expect(mockApi.get).toHaveBeenCalledWith(`/merchants/settlements/${mockSettlementId}`);
    expect(result.data).toEqual(mockResponse);
  });

  test('getAPIKeys calls api.get with correct parameters', async () => {
    const mockResponse = {
      success: true,
      data: [
        {
          id: 1,
          merchant_id: 1,
          name: 'Test API Key',
          provider: 'alipay',
          api_key: 'test-api-key',
          quota_limit: 1000,
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ],
      message: 'API keys retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.getAPIKeys();

    expect(mockApi.get).toHaveBeenCalledWith('/merchants/api-keys');
    expect(result.data).toEqual(mockResponse);
  });

  test('createAPIKey calls api.post with correct parameters', async () => {
    const mockData = {
      name: 'Test API Key',
      provider: 'alipay',
      api_key: 'test-api-key',
      api_secret: 'test-api-secret',
      quota_limit: 1000,
    };
    const mockResponse = {
      id: 1,
      merchant_id: 1,
      name: 'Test API Key',
      provider: 'alipay',
      api_key: 'test-api-key',
      quota_limit: 1000,
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.createAPIKey(mockData);

    expect(mockApi.post).toHaveBeenCalledWith('/merchants/api-keys', mockData);
    expect(result.data).toEqual(mockResponse);
  });

  test('updateAPIKey calls api.put with correct parameters', async () => {
    const mockApiKeyId = 1;
    const mockData = {
      name: 'Updated API Key',
      quota_limit: 2000,
    };
    const mockResponse = {
      id: 1,
      merchant_id: 1,
      name: 'Updated API Key',
      provider: 'alipay',
      api_key: 'test-api-key',
      quota_limit: 2000,
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockApi.put.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.updateAPIKey(mockApiKeyId, mockData);

    expect(mockApi.put).toHaveBeenCalledWith(`/merchants/api-keys/${mockApiKeyId}`, mockData);
    expect(result.data).toEqual(mockResponse);
  });

  test('deleteAPIKey calls api.delete with correct parameters', async () => {
    const mockApiKeyId = 1;
    const mockResponse = {
      success: true,
      message: 'API key deleted successfully',
    };

    mockApi.delete.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.deleteAPIKey(mockApiKeyId);

    expect(mockApi.delete).toHaveBeenCalledWith(`/merchants/api-keys/${mockApiKeyId}`);
    expect(result.data).toEqual(mockResponse);
  });

  test('getAPIKeyUsage calls api.get with correct parameters', async () => {
    const mockResponse = {
      success: true,
      data: [
        {
          api_key_id: 1,
          api_key_name: 'Test API Key',
          usage_count: 500,
          quota_limit: 1000,
          percentage: 50,
        },
      ],
      message: 'API key usage retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await merchantService.getAPIKeyUsage();

    expect(mockApi.get).toHaveBeenCalledWith('/merchants/api-keys/usage');
    expect(result.data).toEqual(mockResponse);
  });
});
