import { useTokenStore } from '../tokenStore';
import { tokenService } from '@/services/token';

// 模拟tokenService
jest.mock('@/services/token', () => ({
  tokenService: {
    getBalance: jest.fn(),
    getConsumption: jest.fn(),
    getAPIKeys: jest.fn(),
    createAPIKey: jest.fn(),
    deleteAPIKey: jest.fn(),
    transfer: jest.fn(),
  },
}));

const mockTokenService = tokenService as jest.Mocked<typeof tokenService>;

describe('tokenStore', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // 重置store状态
    useTokenStore.setState({
      balance: null,
      transactions: [],
      apiKeys: [],
      isLoading: false,
      error: null,
    });
  });

  test('fetchBalance 成功获取余额', async () => {
    const mockBalance = {
      id: 1,
      user_id: 1,
      amount: 1000,
      frozen_amount: 100,
      balance: 900,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    mockTokenService.getBalance.mockResolvedValue({ data: mockBalance } as any);

    const store = useTokenStore.getState();
    await store.fetchBalance();

    const newState = useTokenStore.getState();
    expect(newState.balance).toEqual(mockBalance);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchBalance 获取余额失败', async () => {
    const errorMessage = '获取余额失败';
    mockTokenService.getBalance.mockRejectedValue(new Error(errorMessage));

    const store = useTokenStore.getState();
    await store.fetchBalance();

    const newState = useTokenStore.getState();
    expect(newState.balance).toBe(null);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchTransactions 成功获取交易记录', async () => {
    const mockTransactions = [
      {
        id: 1,
        user_id: 1,
        amount: -100,
        type: 'usage' as const,
        status: 'completed',
        created_at: new Date().toISOString(),
      },
      {
        id: 2,
        user_id: 1,
        amount: 50,
        type: 'reward' as const,
        status: 'completed',
        created_at: new Date().toISOString(),
      },
    ];

    mockTokenService.getConsumption.mockResolvedValue({ data: mockTransactions } as any);

    const store = useTokenStore.getState();
    await store.fetchTransactions();

    const newState = useTokenStore.getState();
    expect(newState.transactions).toEqual(mockTransactions);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchTransactions 获取交易记录失败', async () => {
    const errorMessage = '获取交易记录失败';
    mockTokenService.getConsumption.mockRejectedValue(new Error(errorMessage));

    const store = useTokenStore.getState();
    await store.fetchTransactions();

    const newState = useTokenStore.getState();
    expect(newState.transactions).toEqual([]);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchAPIKeys 成功获取API密钥', async () => {
    const mockAPIKeys = [
      {
        id: 1,
        user_id: 1,
        name: '测试密钥1',
        api_key: 'test_key1',
        status: 'active',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      {
        id: 2,
        user_id: 1,
        name: '测试密钥2',
        api_key: 'test_key2',
        status: 'inactive',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ];

    mockTokenService.getAPIKeys.mockResolvedValue({
      data: {
        data: mockAPIKeys,
        pagination: { total: 2, page: 1, per_page: 20 },
      },
    } as any);

    const store = useTokenStore.getState();
    await store.fetchAPIKeys();

    const newState = useTokenStore.getState();
    expect(newState.apiKeys).toEqual(mockAPIKeys);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchAPIKeys 获取API密钥失败', async () => {
    const errorMessage = '获取API密钥失败';
    mockTokenService.getAPIKeys.mockRejectedValue(new Error(errorMessage));

    const store = useTokenStore.getState();
    await store.fetchAPIKeys();

    const newState = useTokenStore.getState();
    expect(newState.apiKeys).toEqual([]);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('createAPIKey 成功创建API密钥', async () => {
    const mockName = '新API密钥';

    mockTokenService.createAPIKey.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: {
          id: 3,
          user_id: 1,
          name: mockName,
          api_key: 'new_api_key',
          status: 'active',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      },
    } as any);

    const store = useTokenStore.getState();
    const success = await store.createAPIKey(mockName);

    const newState = useTokenStore.getState();
    expect(success).toBe(true);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('createAPIKey 创建API密钥失败', async () => {
    const errorMessage = '创建API密钥失败';
    mockTokenService.createAPIKey.mockRejectedValue(new Error(errorMessage));

    const store = useTokenStore.getState();
    const success = await store.createAPIKey('新API密钥');

    const newState = useTokenStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('deleteAPIKey 成功删除API密钥', async () => {
    const mockId = 1;

    mockTokenService.deleteAPIKey.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: { message: '删除成功' },
      },
    } as any);

    const store = useTokenStore.getState();
    const success = await store.deleteAPIKey(mockId);

    const newState = useTokenStore.getState();
    expect(success).toBe(true);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('deleteAPIKey 删除API密钥失败', async () => {
    const errorMessage = '删除API密钥失败';
    mockTokenService.deleteAPIKey.mockRejectedValue(new Error(errorMessage));

    const store = useTokenStore.getState();
    const success = await store.deleteAPIKey(1);

    const newState = useTokenStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('transfer 成功转账', async () => {
    const mockRecipientId = 2;
    const mockAmount = 100;

    mockTokenService.transfer.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: { message: '转账成功' },
      },
    } as any);

    const store = useTokenStore.getState();
    const success = await store.transfer(mockRecipientId, mockAmount);

    const newState = useTokenStore.getState();
    expect(success).toBe(true);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('transfer 转账失败', async () => {
    const errorMessage = '转账失败';
    mockTokenService.transfer.mockRejectedValue(new Error(errorMessage));

    const store = useTokenStore.getState();
    const success = await store.transfer(2, 100);

    const newState = useTokenStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('clearError 清除错误信息', () => {
    // 先设置一个错误
    useTokenStore.setState({ error: '测试错误' });

    const store = useTokenStore.getState();
    store.clearError();

    const newState = useTokenStore.getState();
    expect(newState.error).toBe(null);
  });
});
