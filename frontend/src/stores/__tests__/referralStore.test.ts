import { useReferralStore } from '../referralStore';
import { referralService } from '@/services/referral';

// 模拟referralService
jest.mock('@/services/referral', () => ({
  referralService: {
    getMyReferralCode: jest.fn(),
    getReferralStats: jest.fn(),
    getReferralList: jest.fn(),
    getReferralRewards: jest.fn(),
    bindReferralCode: jest.fn(),
  },
}));

const mockReferralService = referralService as jest.Mocked<typeof referralService>;

describe('referralStore', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // 重置store状态
    useReferralStore.setState({
      referralCode: '',
      stats: null,
      referrals: [],
      rewards: [],
      isLoading: false,
      error: null,
    });
  });

  test('fetchReferralCode 成功获取邀请码', async () => {
    const mockCode = 'TESTCODE123';

    mockReferralService.getMyReferralCode.mockResolvedValue({ data: { code: mockCode } } as any);

    const store = useReferralStore.getState();
    await store.fetchReferralCode();

    const newState = useReferralStore.getState();
    expect(newState.referralCode).toBe(mockCode);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchReferralCode 获取邀请码失败', async () => {
    const errorMessage = '获取邀请码失败';
    mockReferralService.getMyReferralCode.mockRejectedValue(new Error(errorMessage));

    const store = useReferralStore.getState();
    await store.fetchReferralCode();

    const newState = useReferralStore.getState();
    expect(newState.referralCode).toBe('');
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchStats 成功获取统计数据', async () => {
    const mockStats = {
      total_referrals: 10,
      total_rewards: 1000,
      pending_rewards: 200,
      completed_rewards: 800,
      paid_rewards: 600,
    };

    mockReferralService.getReferralStats.mockResolvedValue({ data: mockStats } as any);

    const store = useReferralStore.getState();
    await store.fetchStats();

    const newState = useReferralStore.getState();
    expect(newState.stats).toEqual(mockStats);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchStats 获取统计数据失败', async () => {
    const errorMessage = '获取统计数据失败';
    mockReferralService.getReferralStats.mockRejectedValue(new Error(errorMessage));

    const store = useReferralStore.getState();
    await store.fetchStats();

    const newState = useReferralStore.getState();
    expect(newState.stats).toBe(null);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchReferrals 成功获取邀请列表', async () => {
    const mockReferrals = [
      {
        id: 1,
        referrer_id: 1,
        referee_id: 2,
        code: 'TESTCODE123',
        status: 'completed',
        created_at: new Date().toISOString(),
      },
      {
        id: 2,
        referrer_id: 1,
        referee_id: 3,
        code: 'TESTCODE123',
        status: 'pending',
        created_at: new Date().toISOString(),
      },
    ];

    mockReferralService.getReferralList.mockResolvedValue({
      data: {
        data: mockReferrals,
        pagination: { total: 2, page: 1, per_page: 20 },
      },
    } as any);

    const store = useReferralStore.getState();
    await store.fetchReferrals();

    const newState = useReferralStore.getState();
    expect(newState.referrals).toEqual(mockReferrals);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchReferrals 获取邀请列表失败', async () => {
    const errorMessage = '获取邀请列表失败';
    mockReferralService.getReferralList.mockRejectedValue(new Error(errorMessage));

    const store = useReferralStore.getState();
    await store.fetchReferrals();

    const newState = useReferralStore.getState();
    expect(newState.referrals).toEqual([]);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('fetchRewards 成功获取返利记录', async () => {
    const mockRewards = [
      {
        id: 1,
        referral_id: 1,
        amount: 100,
        status: 'completed',
        created_at: new Date().toISOString(),
      },
      {
        id: 2,
        referral_id: 2,
        amount: 50,
        status: 'pending',
        created_at: new Date().toISOString(),
      },
    ];

    mockReferralService.getReferralRewards.mockResolvedValue({
      data: {
        data: mockRewards,
        pagination: { total: 2, page: 1, per_page: 20 },
      },
    } as any);

    const store = useReferralStore.getState();
    await store.fetchRewards();

    const newState = useReferralStore.getState();
    expect(newState.rewards).toEqual(mockRewards);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
  });

  test('fetchRewards 获取返利记录失败', async () => {
    const errorMessage = '获取返利记录失败';
    mockReferralService.getReferralRewards.mockRejectedValue(new Error(errorMessage));

    const store = useReferralStore.getState();
    await store.fetchRewards();

    const newState = useReferralStore.getState();
    expect(newState.rewards).toEqual([]);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('bindReferralCode 成功绑定邀请码', async () => {
    const mockCode = 'TESTCODE123';
    const mockStats = {
      total_referrals: 1,
      total_rewards: 100,
      pending_rewards: 100,
      completed_rewards: 0,
      paid_rewards: 0,
    };

    mockReferralService.bindReferralCode.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: { message: '绑定成功' },
      },
    } as any);
    mockReferralService.getReferralStats.mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: mockStats,
      },
    } as any);

    const store = useReferralStore.getState();
    const success = await store.bindReferralCode(mockCode);

    const newState = useReferralStore.getState();
    expect(success).toBe(true);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(null);
    expect(mockReferralService.getReferralStats).toHaveBeenCalled();
  });

  test('bindReferralCode 绑定邀请码失败', async () => {
    const errorMessage = '绑定邀请码失败';
    mockReferralService.bindReferralCode.mockRejectedValue(new Error(errorMessage));

    const store = useReferralStore.getState();
    const success = await store.bindReferralCode('TESTCODE123');

    const newState = useReferralStore.getState();
    expect(success).toBe(false);
    expect(newState.isLoading).toBe(false);
    expect(newState.error).toBe(errorMessage);
  });

  test('clearError 清除错误信息', () => {
    // 先设置一个错误
    useReferralStore.setState({ error: '测试错误' });

    const store = useReferralStore.getState();
    store.clearError();

    const newState = useReferralStore.getState();
    expect(newState.error).toBe(null);
  });
});
