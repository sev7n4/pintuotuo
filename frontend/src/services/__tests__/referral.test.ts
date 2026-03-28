import { referralService } from '../referral';
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

describe('referralService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('getMyReferralCode calls api.get with correct parameters', async () => {
    const mockResponse = {
      success: true,
      data: {
        code: 'TEST123',
      },
      message: 'Referral code retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await referralService.getMyReferralCode();

    expect(mockApi.get).toHaveBeenCalledWith('/referrals/code');
    expect(result.data).toEqual(mockResponse);
  });

  test('validateReferralCode calls api.get with correct parameters', async () => {
    const mockCode = 'TEST123';
    const mockResponse = {
      success: true,
      data: {
        valid: true,
        referrer_id: 1,
        referrer_name: 'Test User',
      },
      message: 'Referral code validated successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await referralService.validateReferralCode(mockCode);

    expect(mockApi.get).toHaveBeenCalledWith(`/referrals/validate/${mockCode}`);
    expect(result.data).toEqual(mockResponse);
  });

  test('bindReferralCode calls api.post with correct parameters', async () => {
    const mockCode = 'TEST123';
    const mockResponse = {
      success: true,
      data: {
        message: 'Referral code bound successfully',
      },
      message: 'Referral code bound successfully',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await referralService.bindReferralCode(mockCode);

    expect(mockApi.post).toHaveBeenCalledWith('/referrals/bind', { code: mockCode });
    expect(result.data).toEqual(mockResponse);
  });

  test('getReferralStats calls api.get with correct parameters', async () => {
    const mockResponse = {
      total_referrals: 10,
      total_rewards: 500,
      pending_rewards: 100,
      paid_rewards: 400,
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await referralService.getReferralStats();

    expect(mockApi.get).toHaveBeenCalledWith('/referrals/stats');
    expect(result.data).toEqual(mockResponse);
  });

  test('getReferralList calls api.get with correct parameters', async () => {
    const mockPage = 1;
    const mockPerPage = 10;
    const mockResponse = {
      success: true,
      data: {
        items: [
          {
            id: 1,
            referrer_id: 1,
            referee_id: 2,
            code: 'TEST123',
            status: 'active',
            created_at: '2024-01-01T00:00:00Z',
          },
        ],
        total: 1,
        page: 1,
        per_page: 10,
      },
      message: 'Referrals retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await referralService.getReferralList(mockPage, mockPerPage);

    expect(mockApi.get).toHaveBeenCalledWith('/referrals/list', {
      params: { page: mockPage, per_page: mockPerPage },
    });
    expect(result.data).toEqual(mockResponse);
  });

  test('getReferralRewards calls api.get with correct parameters', async () => {
    const mockPage = 1;
    const mockPerPage = 10;
    const mockStatus = 'pending';
    const mockResponse = {
      success: true,
      data: {
        items: [
          {
            id: 1,
            referral_id: 1,
            amount: 50,
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        ],
        total: 1,
        page: 1,
        per_page: 10,
      },
      message: 'Referral rewards retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await referralService.getReferralRewards(mockPage, mockPerPage, mockStatus);

    expect(mockApi.get).toHaveBeenCalledWith('/referrals/rewards', {
      params: { page: mockPage, per_page: mockPerPage, status: mockStatus },
    });
    expect(result.data).toEqual(mockResponse);
  });

  test('payReferralRewards calls api.post with correct parameters', async () => {
    const mockRewardIds = [1, 2, 3];
    const mockResponse = {
      success: true,
      data: {
        message: 'Referral rewards paid successfully',
      },
      message: 'Referral rewards paid successfully',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await referralService.payReferralRewards(mockRewardIds);

    expect(mockApi.post).toHaveBeenCalledWith('/referrals/rewards/pay', {
      reward_ids: mockRewardIds,
    });
    expect(result.data).toEqual(mockResponse);
  });
});
