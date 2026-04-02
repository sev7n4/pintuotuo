import { useMerchantStore } from '../merchantStore';
import { merchantService } from '@/services/merchant';

// 模拟merchantService
jest.mock('@/services/merchant', () => ({
  merchantService: {
    getProfile: jest.fn(),
    updateProfile: jest.fn(),
    getStats: jest.fn(),
    getOrders: jest.fn(),
    getSettlements: jest.fn(),
    requestSettlement: jest.fn(),
    getAPIKeys: jest.fn(),
    createAPIKey: jest.fn(),
    updateAPIKey: jest.fn(),
    deleteAPIKey: jest.fn(),
    getAPIKeyUsage: jest.fn(),
  },
}));

const mockMerchantService = merchantService as jest.Mocked<typeof merchantService>;

describe('merchantStore', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // 重置store状态
    useMerchantStore.setState({
      profile: null,
      stats: null,
      orders: [],
      settlements: [],
      apiKeys: [],
      apiKeyUsage: [],
      isLoading: false,
      error: null,
    });
  });

  test('fetchProfile 成功获取商家信息', async () => {
    const mockProfile = {
      id: 1,
      user_id: 1,
      name: '测试商家',
      company_name: '测试公司',
      email: 'test@merchant.com',
      phone: '13800138000',
      address: '测试地址',
      business_license: '1234567890',
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockMerchantService.getProfile.mockResolvedValue({ data: mockProfile } as any);

    const store = useMerchantStore.getState();
    await store.fetchProfile();

    const newState = useMerchantStore.getState();
    expect(newState.profile).toEqual(mockProfile);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchProfile 获取商家信息失败', async () => {
    const errorMessage = '获取商家信息失败';
    mockMerchantService.getProfile.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    await store.fetchProfile();

    const newState = useMerchantStore.getState();
    expect(newState.profile).toBe(null);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('updateProfile 成功更新商家信息', async () => {
    const mockProfile = {
      id: 1,
      user_id: 1,
      name: '测试商家',
      company_name: '测试公司',
      email: 'test@merchant.com',
      phone: '13800138000',
      address: '测试地址',
      business_license: '1234567890',
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    const updateData = { company_name: '更新后的商家' };
    const updatedProfile = { ...mockProfile, ...updateData };

    mockMerchantService.updateProfile.mockResolvedValue({ data: updatedProfile } as any);

    const store = useMerchantStore.getState();
    const success = await store.updateProfile(updateData);

    const newState = useMerchantStore.getState();
    expect(success).toBe(true);
    expect(newState.profile).toEqual(updatedProfile);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('updateProfile 更新商家信息失败', async () => {
    const errorMessage = '更新商家信息失败';
    mockMerchantService.updateProfile.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    const success = await store.updateProfile({ company_name: '更新后的商家' } as any);

    const newState = useMerchantStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchStats 成功获取统计数据', async () => {
    const mockStats = {
      total_orders: 100,
      total_sales: 10000,
      total_products: 50,
      pending_orders: 5,
      active_products: 45,
      month_sales: 5000,
      month_orders: 50,
    };

    mockMerchantService.getStats.mockResolvedValue({ data: mockStats } as any);

    const store = useMerchantStore.getState();
    await store.fetchStats();

    const newState = useMerchantStore.getState();
    expect(newState.stats).toEqual(mockStats);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchStats 获取统计数据失败', async () => {
    const errorMessage = '获取统计数据失败';
    mockMerchantService.getStats.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    await store.fetchStats();

    const newState = useMerchantStore.getState();
    expect(newState.stats).toBe(null);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchOrders 成功获取订单列表', async () => {
    const mockOrders = [
      { id: 1, order_id: 'ORD123', amount: 100, status: 'completed' },
      { id: 2, order_id: 'ORD124', amount: 200, status: 'pending' },
    ];

    mockMerchantService.getOrders.mockResolvedValue({
      data: {
        total: 2,
        page: 1,
        per_page: 20,
        data: mockOrders,
      },
    } as any);

    const store = useMerchantStore.getState();
    await store.fetchOrders();

    const newState = useMerchantStore.getState();
    expect(newState.orders).toEqual(mockOrders);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchOrders 获取订单列表失败', async () => {
    const errorMessage = '获取订单列表失败';
    mockMerchantService.getOrders.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    await store.fetchOrders();

    const newState = useMerchantStore.getState();
    expect(newState.orders).toEqual([]);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchSettlements 成功获取结算记录', async () => {
    const mockSettlements = [
      {
        id: 1,
        merchant_id: 1,
        amount: 1000,
        status: 'completed',
        period_start: new Date().toISOString(),
        period_end: new Date().toISOString(),
        total_sales: 5000,
        fee: 100,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      {
        id: 2,
        merchant_id: 1,
        amount: 2000,
        status: 'pending',
        period_start: new Date().toISOString(),
        period_end: new Date().toISOString(),
        total_sales: 10000,
        fee: 200,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ];

    mockMerchantService.getSettlements.mockResolvedValue({
      data: {
        total: 2,
        page: 1,
        per_page: 20,
        data: mockSettlements,
      },
    } as any);

    const store = useMerchantStore.getState();
    await store.fetchSettlements();

    const newState = useMerchantStore.getState();
    expect(newState.settlements).toEqual(mockSettlements);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchSettlements 获取结算记录失败', async () => {
    const errorMessage = '获取结算记录失败';
    mockMerchantService.getSettlements.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    await store.fetchSettlements();

    const newState = useMerchantStore.getState();
    expect(newState.settlements).toEqual([]);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('requestSettlement 成功申请结算', async () => {
    mockMerchantService.requestSettlement.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: { message: '申请结算成功' },
      },
    } as any);

    const store = useMerchantStore.getState();
    const success = await store.requestSettlement();

    const newState = useMerchantStore.getState();
    expect(success).toBe(true);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('requestSettlement 申请结算失败', async () => {
    const errorMessage = '申请结算失败';
    mockMerchantService.requestSettlement.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    const success = await store.requestSettlement();

    const newState = useMerchantStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchAPIKeys 成功获取API密钥', async () => {
    const mockAPIKeys = [
      {
        id: 1,
        merchant_id: 1,
        name: '测试密钥',
        api_key: 'test_key',
        provider: 'openai',
        status: 'active' as const,
        quota_limit: 1000,
        quota_used: 0,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 2,
        merchant_id: 1,
        name: '测试密钥2',
        api_key: 'test_key2',
        provider: 'claude',
        status: 'inactive' as const,
        quota_limit: 2000,
        quota_used: 500,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ];

    mockMerchantService.getAPIKeys.mockResolvedValue({
      data: {
        total: 2,
        page: 1,
        per_page: 20,
        data: mockAPIKeys,
      },
    } as any);

    const store = useMerchantStore.getState();
    await store.fetchAPIKeys();

    const newState = useMerchantStore.getState();
    expect(newState.apiKeys).toEqual(mockAPIKeys);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchAPIKeys 获取API密钥失败', async () => {
    const errorMessage = '获取API密钥失败';
    mockMerchantService.getAPIKeys.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    await store.fetchAPIKeys();

    const newState = useMerchantStore.getState();
    expect(newState.apiKeys).toEqual([]);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('createAPIKey 成功创建API密钥', async () => {
    const apiKeyData = {
      name: '新密钥',
      provider: 'openai',
      api_key: 'new_api_key',
      quota_limit: 1000,
    };

    mockMerchantService.createAPIKey.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: {
          id: 3,
          merchant_id: 1,
          name: '新密钥',
          api_key: 'new_api_key',
          provider: 'openai',
          status: 'active' as const,
          quota_limit: 1000,
          quota_used: 0,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      },
    } as any);

    const store = useMerchantStore.getState();
    const success = await store.createAPIKey(apiKeyData);

    const newState = useMerchantStore.getState();
    expect(success).toBe(true);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('createAPIKey 创建API密钥失败', async () => {
    const errorMessage = '创建API密钥失败';
    mockMerchantService.createAPIKey.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    const success = await store.createAPIKey({
      name: '新密钥',
      provider: 'openai',
      api_key: 'new_api_key',
    });

    const newState = useMerchantStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('updateAPIKey 成功更新API密钥', async () => {
    const updateData = { name: '更新的密钥', status: 'active' as const };

    mockMerchantService.updateAPIKey.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: {
          id: 1,
          merchant_id: 1,
          name: '更新的密钥',
          api_key: 'test_key',
          provider: 'openai',
          status: 'active' as const,
          quota_limit: 1000,
          quota_used: 0,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      },
    } as any);

    const store = useMerchantStore.getState();
    const success = await store.updateAPIKey(1, updateData);

    const newState = useMerchantStore.getState();
    expect(success).toBe(true);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('updateAPIKey 更新API密钥失败', async () => {
    const errorMessage = '更新API密钥失败';
    mockMerchantService.updateAPIKey.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    const success = await store.updateAPIKey(1, { name: '更新的密钥' });

    const newState = useMerchantStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('deleteAPIKey 成功删除API密钥', async () => {
    mockMerchantService.deleteAPIKey.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: { message: '删除成功' },
      },
    } as any);

    const store = useMerchantStore.getState();
    const success = await store.deleteAPIKey(1);

    const newState = useMerchantStore.getState();
    expect(success).toBe(true);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('deleteAPIKey 删除API密钥失败', async () => {
    const errorMessage = '删除API密钥失败';
    mockMerchantService.deleteAPIKey.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    const success = await store.deleteAPIKey(1);

    const newState = useMerchantStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchAPIKeyUsage 成功获取API密钥使用情况', async () => {
    const mockUsage = [
      {
        id: 1,
        api_key_id: 1,
        name: '测试密钥',
        provider: 'openai',
        requests: 100,
        tokens: 5000,
        quota_limit: 1000,
        quota_used: 100,
        usage_percentage: 10,
        date: '2024-01-01',
      },
      {
        id: 2,
        api_key_id: 1,
        name: '测试密钥',
        provider: 'openai',
        requests: 200,
        tokens: 10000,
        quota_limit: 1000,
        quota_used: 300,
        usage_percentage: 30,
        date: '2024-01-02',
      },
    ];

    mockMerchantService.getAPIKeyUsage.mockResolvedValue({
      data: {
        total: 2,
        page: 1,
        per_page: 20,
        data: mockUsage,
      },
    } as any);

    const store = useMerchantStore.getState();
    await store.fetchAPIKeyUsage();

    const newState = useMerchantStore.getState();
    expect(newState.apiKeyUsage).toEqual(mockUsage);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchAPIKeyUsage 获取API密钥使用情况失败', async () => {
    const errorMessage = '获取使用情况失败';
    mockMerchantService.getAPIKeyUsage.mockRejectedValue(new Error(errorMessage));

    const store = useMerchantStore.getState();
    await store.fetchAPIKeyUsage();

    const newState = useMerchantStore.getState();
    expect(newState.apiKeyUsage).toEqual([]);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('clearError 清除错误信息', () => {
    const store = useMerchantStore.getState();
    store.clearError();
    expect(store.error).toBe(null);
  });
});
