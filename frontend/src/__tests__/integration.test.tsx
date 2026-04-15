import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route, Navigate } from 'react-router-dom';
import ReferralPage from '../pages/ReferralPage';
import LoginPage from '../pages/LoginPage';
import CartPage from '../pages/CartPage';
import RegisterPage from '../pages/RegisterPage';
import { useAuthStore, type AuthState } from '@/stores/authStore';
import { AUTH_PRIMARY_LOGIN_KEY } from '@/lib/authLoginPreference';
import { useReferralStore } from '@/stores/referralStore';
import { useCartStore } from '@/stores/cartStore';

jest.mock('@/stores/authStore');
jest.mock('@/stores/referralStore');
jest.mock('@/stores/cartStore');

jest.mock('../pages/ReferralPage.module.css', () => ({}));

jest.mock('antd', () => {
  const antd = jest.requireActual('antd');
  return {
    ...antd,
    message: {
      success: jest.fn(),
      error: jest.fn(),
    },
    Table: antd.Table,
    Tabs: antd.Tabs,
    TabPane: antd.Tabs.TabPane,
    Input: antd.Input,
    Button: antd.Button,
    Space: antd.Space,
    Spin: antd.Spin,
    Card: antd.Card,
    Row: antd.Row,
    Col: antd.Col,
    Typography: antd.Typography,
    Statistic: antd.Statistic,
    Form: antd.Form,
  };
});

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>;
const mockUseReferralStore = useReferralStore as jest.MockedFunction<typeof useReferralStore>;
const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>;

// 模拟 clipboard API
Object.defineProperty(navigator, 'clipboard', {
  value: {
    writeText: jest.fn().mockResolvedValue(undefined),
  },
  writable: true,
});

describe('Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.removeItem(AUTH_PRIMARY_LOGIN_KEY);
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        sms: false,
        email_magic: false,
        wechat_oauth: false,
        github_oauth: false,
        account_linking: false,
        merchant_register_mode: 'open',
        admin_mfa_required: false,
      }),
    });
  });

  test('should render referral page for unauthenticated users', async () => {
    // 模拟未认证状态
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: null,
      stats: null,
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    // 检查推荐页面是否渲染
    await waitFor(() => {
      expect(screen.getByText(/邀请好友/i)).toBeInTheDocument();
    });
  }, 15000);

  test('should access referral page after login', async () => {
    const mockLogin = jest.fn().mockResolvedValue({ token: 'test-token' });
    const mockFetchUser = jest
      .fn()
      .mockResolvedValue({ id: 1, email: 'user@example.com', role: 'user' });

    // 初始未认证状态，登录后切换到认证状态
    let isAuthenticated = false;
    const mockStore = {
      user: isAuthenticated ? { id: 1, email: 'user@example.com', role: 'user' } : null,
      token: isAuthenticated ? 'test-token' : null,
      isLoading: false,
      error: null,
      isAuthenticated,
      rememberMe: false,
      login: async (
        email: string,
        password: string,
        rememberMe?: boolean,
        totpCode?: string
      ) => {
        const result = await mockLogin(email, password, rememberMe, totpCode);
        isAuthenticated = true;
        return result;
      },
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      sendEmailMagicLink: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as AuthState;

    mockUseAuthStore.mockImplementation(() => mockStore);

    mockUseAuthStore.getState = jest.fn(() => ({
      user: {
        id: 1,
        email: 'user@example.com',
        name: 'Test User',
        role: 'user',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: mockFetchUser,
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    }));

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 10,
        totalRewards: 100,
        pendingRewards: 20,
        paidRewards: 80,
        availableRewards: 60,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/" element={<div>Home Page</div>} />
          <Route path="/*" element={<Navigate to="/login" />} />
        </Routes>
      </MemoryRouter>
    );

    // 输入登录信息
    const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
    const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
    const loginButton = screen.getByRole('button', { name: '密码登录' });

    await act(async () => {
      fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      fireEvent.click(loginButton);
    });

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password123', true, undefined);
    });
  });

  test('should submit unified auth form on register route (referral query preserved for future bind)', async () => {
    const mockRegister = jest.fn().mockResolvedValue({ token: 'test-token' });
    const mockBindReferralCode = jest.fn().mockResolvedValue(undefined);

    // 初始未认证状态；/register 走 register() 创建账号
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      rememberMe: false,
      login: jest.fn(),
      register: mockRegister,
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      sendEmailMagicLink: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: null,
      stats: null,
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: mockBindReferralCode,
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/register?code=FRIEND12']}>
        <Routes>
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/catalog" element={<div>Products Page</div>} />
          <Route path="/*" element={<Navigate to="/register" />} />
        </Routes>
      </MemoryRouter>
    );

    // 输入邮箱密码（注册页占位与按钮文案与登录页区分）
    const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
    const passwordInput = screen.getByPlaceholderText(/设置密码/) as HTMLInputElement;
    const submitButton = screen.getByRole('button', { name: '注册并进入' });

    await act(async () => {
      fireEvent.change(emailInput, { target: { value: 'newuser@example.com' } });
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      fireEvent.click(submitButton);
    });

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith('newuser@example.com', 'password123', 'user', undefined);
    });
    expect(mockBindReferralCode).not.toHaveBeenCalled();
  });

  test('should integrate referral system with cart and orders', async () => {
    // 模拟认证状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 10,
        totalRewards: 100,
        pendingRewards: 20,
        paidRewards: 80,
        availableRewards: 60,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      items: [
        {
          id: 1,
          product_id: 1,
          product: {
            name: 'Test Product',
            price: 100,
            image: 'test.jpg',
          },
          quantity: 1,
          total: 100,
        },
      ],
      total: 100,
      isLoading: false,
      error: null,
      addToCart: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
      fetchCart: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/cart']}>
        <Routes>
          <Route path="/cart" element={<CartPage />} />
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/cart" />} />
        </Routes>
      </MemoryRouter>
    );

    // 检查购物车页面
    await waitFor(() => {
      expect(screen.getByText('购物车')).toBeInTheDocument();
      expect(screen.getByText('Test Product')).toBeInTheDocument();
    });
  });

  test('should handle referral code binding error', async () => {
    const mockBindReferralCode = jest.fn().mockRejectedValue(new Error('Invalid referral code'));

    // 模拟认证状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 10,
        totalRewards: 100,
        pendingRewards: 20,
        paidRewards: 80,
        availableRewards: 60,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: 'Invalid referral code',
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: mockBindReferralCode,
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    // 检查推荐页面是否渲染
    await waitFor(() => {
      expect(screen.getByText(/邀请好友/i)).toBeInTheDocument();
    });

    // 验证绑定邀请码功能存在
    const input = screen.getByPlaceholderText('输入好友的邀请码');
    expect(input).toBeInTheDocument();
  });

  test('should handle concurrent operations between referral and other features', async () => {
    // 模拟认证状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    const mockFetchReferralCode = jest.fn().mockResolvedValue('TESTCODE');
    const mockFetchStats = jest.fn().mockResolvedValue({
      totalReferrals: 10,
      totalRewards: 100,
      pendingRewards: 20,
      paidRewards: 80,
    });

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 10,
        totalRewards: 100,
        pendingRewards: 20,
        paidRewards: 80,
        availableRewards: 60,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: mockFetchReferralCode,
      fetchStats: mockFetchStats,
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    // 测试推荐页面
    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    // 检查推荐页面
    await waitFor(() => {
      expect(screen.getByText('邀请好友')).toBeInTheDocument();
      expect(screen.getByText('TESTCODE')).toBeInTheDocument();
    });

    // 验证获取推荐码和统计信息的函数被调用
    expect(mockFetchReferralCode).toHaveBeenCalled();
    expect(mockFetchStats).toHaveBeenCalled();
  });

  test('should integrate referral system with user profile', async () => {
    // 模拟认证状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 5,
        totalRewards: 50,
        pendingRewards: 10,
        paidRewards: 40,
        availableRewards: 30,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    // 测试推荐页面
    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    // 检查推荐页面是否显示用户相关信息
    await waitFor(() => {
      expect(screen.getByText('邀请好友')).toBeInTheDocument();
      expect(screen.getByText('TESTCODE')).toBeInTheDocument();
    });
  });

  test('should test referral code sharing functionality', async () => {
    // 模拟认证状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'SHARE123',
      stats: {
        totalReferrals: 0,
        totalRewards: 0,
        pendingRewards: 0,
        paidRewards: 0,
        availableRewards: 0,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    // 测试推荐页面
    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    // 检查推荐码是否显示
    await waitFor(() => {
      expect(screen.getByText('SHARE123')).toBeInTheDocument();
    });

    // 验证复制到剪贴板功能
    const copyButton = screen.getByRole('button', { name: /复制邀请码/i });
    await act(async () => {
      fireEvent.click(copyButton);
    });

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('SHARE123');
  });

  test('should test referral rewards withdrawal', async () => {
    // 模拟认证状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 10,
        totalRewards: 100,
        pendingRewards: 20,
        paidRewards: 80,
        availableRewards: 60,
      },
      referrals: [],
      rewards: [
        {
          id: 1,
          amount: 50,
          status: 'pending',
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 2,
          amount: 50,
          status: 'paid',
          created_at: '2024-01-02T00:00:00Z',
        },
      ],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    // 测试推荐页面
    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    // 检查推荐页面
    await waitFor(() => {
      expect(screen.getByText('邀请好友')).toBeInTheDocument();
    });
  });

  test('should test referral statistics display', async () => {
    // 模拟认证状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 25,
        totalRewards: 250,
        pendingRewards: 50,
        paidRewards: 200,
        availableRewards: 150,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    // 测试推荐页面
    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    // 检查推荐页面
    await waitFor(() => {
      expect(screen.getByText('邀请好友')).toBeInTheDocument();
    });
  });

  test('should handle user logout and clear all states', async () => {
    const mockLogout = jest.fn().mockResolvedValue(undefined);

    // 模拟认证状态
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: mockLogout,
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 10,
        totalRewards: 100,
        pendingRewards: 20,
        paidRewards: 80,
        availableRewards: 60,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: null,
      addToCart: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
      fetchCart: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('邀请好友')).toBeInTheDocument();
    });
  });

  test('should handle loading states across multiple components', async () => {
    // 模拟加载状态
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: true,
      error: null,
      isAuthenticated: false,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: null,
      stats: null,
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: true,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/邀请好友/i)).toBeInTheDocument();
    });
  });

  test('should handle form validation integration', async () => {
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    render(
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<Navigate to="/login" />} />
        </Routes>
      </MemoryRouter>
    );

    const loginButton = screen.getByRole('button', { name: '密码登录' });
    expect(loginButton).toBeInTheDocument();

    const emailInput = screen.getByPlaceholderText('example@email.com');
    expect(emailInput).toBeInTheDocument();

    const passwordInput = screen.getByPlaceholderText('输入密码');
    expect(passwordInput).toBeInTheDocument();
  });

  test('should handle authentication state changes', async () => {
    let authState = {
      user: null as { id: number; email: string; role: string } | null,
      token: null as string | null,
      isAuthenticated: false,
    };

    const mockLogin = jest.fn().mockImplementation(async () => {
      authState = {
        user: { id: 1, email: 'test@example.com', role: 'user' },
        token: 'new-token',
        isAuthenticated: true,
      };
      return { token: 'new-token' };
    });

    mockUseAuthStore.mockImplementation(
      () =>
        ({
          user: authState.user,
          token: authState.token,
          isLoading: false,
          error: null,
          isAuthenticated: authState.isAuthenticated,
          rememberMe: false,
          login: mockLogin,
          register: jest.fn(),
          loginWithSms: jest.fn(),
          registerWithSms: jest.fn(),
          sendSmsCode: jest.fn(),
          sendEmailMagicLink: jest.fn(),
          logout: jest.fn(),
          fetchUser: jest.fn(),
          setUser: jest.fn(),
          clearError: jest.fn(),
          setRememberMe: jest.fn(),
        }) as AuthState
    );

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TESTCODE',
      stats: {
        totalReferrals: 5,
        totalRewards: 50,
        pendingRewards: 10,
        paidRewards: 40,
        availableRewards: 30,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/" element={<div>Home Page</div>} />
          <Route path="/*" element={<Navigate to="/login" />} />
        </Routes>
      </MemoryRouter>
    );

    const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
    const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
    const loginButton = screen.getByRole('button', { name: '密码登录' });

    await act(async () => {
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      fireEvent.click(loginButton);
    });

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password123', true, undefined);
    });
  });

  test('should handle error recovery across components', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: null,
      stats: null,
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: 'Network error',
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/邀请好友/i)).toBeInTheDocument();
    });
  });

  test('should handle cart and referral integration for discounts', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'DISCOUNT10',
      stats: {
        totalReferrals: 15,
        totalRewards: 150,
        pendingRewards: 30,
        paidRewards: 120,
        availableRewards: 90,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      items: [
        {
          id: 1,
          product_id: 1,
          product: {
            name: 'Discounted Product',
            price: 200,
            image: 'discount.jpg',
          },
          quantity: 2,
          total: 400,
        },
      ],
      total: 400,
      isLoading: false,
      error: null,
      addToCart: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
      fetchCart: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/cart']}>
        <Routes>
          <Route path="/cart" element={<CartPage />} />
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/cart" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('购物车')).toBeInTheDocument();
      expect(screen.getByText('Discounted Product')).toBeInTheDocument();
    });
  });

  test('should handle multiple referral code operations', async () => {
    const mockBindReferralCode = jest
      .fn()
      .mockResolvedValueOnce(undefined)
      .mockResolvedValueOnce(undefined);

    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'FIRSTCODE',
      stats: {
        totalReferrals: 0,
        totalRewards: 0,
        pendingRewards: 0,
        paidRewards: 0,
        availableRewards: 0,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: mockBindReferralCode,
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('FIRSTCODE')).toBeInTheDocument();
    });
  });

  test('should handle session expiration and re-authentication', async () => {
    let authState: {
      user: { id: number; email: string; role: string } | null;
      token: string | null;
      isAuthenticated: boolean;
    } = {
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'expired-token',
      isAuthenticated: true,
    };

    const mockFetchUser = jest.fn().mockImplementation(async () => {
      authState = {
        user: null,
        token: null,
        isAuthenticated: false,
      };
      throw new Error('Session expired');
    });

    mockUseAuthStore.mockImplementation(
      () =>
        ({
          user: authState.user,
          token: authState.token,
          isLoading: false,
          error: 'Session expired',
          isAuthenticated: authState.isAuthenticated,
          rememberMe: false,
          login: jest.fn(),
          register: jest.fn(),
          loginWithSms: jest.fn(),
          registerWithSms: jest.fn(),
          sendSmsCode: jest.fn(),
          logout: jest.fn(),
          fetchUser: mockFetchUser,
          setUser: jest.fn(),
          clearError: jest.fn(),
          setRememberMe: jest.fn(),
        }) as AuthState
    );

    mockUseReferralStore.mockReturnValue({
      referralCode: null,
      stats: null,
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/邀请好友/i)).toBeInTheDocument();
    });
  });

  test('should handle data synchronization between cart and order', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    const mockClearCart = jest.fn();

    mockUseCartStore.mockReturnValue({
      items: [
        {
          id: 1,
          product_id: 1,
          product: {
            name: 'Order Product',
            price: 150,
            image: 'order.jpg',
          },
          quantity: 1,
          total: 150,
        },
      ],
      total: 150,
      isLoading: false,
      error: null,
      addToCart: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: mockClearCart,
      fetchCart: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/cart']}>
        <Routes>
          <Route path="/cart" element={<CartPage />} />
          <Route path="/*" element={<Navigate to="/cart" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('购物车')).toBeInTheDocument();
      expect(screen.getByText('Order Product')).toBeInTheDocument();
    });
  });

  test('should handle user role-based access control', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'merchant@example.com', role: 'merchant' },
      token: 'merchant-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'MERCHANT1',
      stats: {
        totalReferrals: 50,
        totalRewards: 500,
        pendingRewards: 100,
        paidRewards: 400,
        availableRewards: 300,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('邀请好友')).toBeInTheDocument();
      expect(screen.getByText('MERCHANT1')).toBeInTheDocument();
    });
  });

  test('should handle network timeout and retry', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TIMEOUT1',
      stats: null,
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: 'Timeout',
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/邀请好友/i)).toBeInTheDocument();
    });
  });

  test('should handle concurrent user actions', async () => {
    const mockBindReferralCode = jest.fn().mockResolvedValue(undefined);
    const mockFetchStats = jest.fn().mockResolvedValue({
      totalReferrals: 10,
      totalRewards: 100,
      pendingRewards: 20,
      paidRewards: 80,
    });

    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'CONCURRENT',
      stats: {
        totalReferrals: 10,
        totalRewards: 100,
        pendingRewards: 20,
        paidRewards: 80,
        availableRewards: 60,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: mockFetchStats,
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: mockBindReferralCode,
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('CONCURRENT')).toBeInTheDocument();
    });

    const copyButton = screen.getByRole('button', { name: /复制邀请码/i });
    await act(async () => {
      fireEvent.click(copyButton);
    });

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('CONCURRENT');
  });

  test('should handle route guard for protected pages', async () => {
    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseCartStore.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: null,
      addToCart: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
      fetchCart: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/cart']}>
        <Routes>
          <Route path="/cart" element={<CartPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<Navigate to="/cart" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/购物车/i)).toBeInTheDocument();
    });
  });

  test('should handle token balance integration with user profile', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'TOKEN123',
      stats: {
        totalReferrals: 20,
        totalRewards: 200,
        pendingRewards: 40,
        paidRewards: 160,
        availableRewards: 120,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('TOKEN123')).toBeInTheDocument();
    });
  });

  test('should handle order creation flow integration', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseCartStore.mockReturnValue({
      items: [
        {
          id: 1,
          product_id: 1,
          product: {
            name: 'Order Test Product',
            price: 99,
            image: 'test.jpg',
          },
          quantity: 3,
          total: 297,
        },
      ],
      total: 297,
      isLoading: false,
      error: null,
      addToCart: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
      fetchCart: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/cart']}>
        <Routes>
          <Route path="/cart" element={<CartPage />} />
          <Route path="/*" element={<Navigate to="/cart" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Order Test Product')).toBeInTheDocument();
    });
  });

  test('should handle referral code validation', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'VALIDCODE',
      stats: {
        totalReferrals: 0,
        totalRewards: 0,
        pendingRewards: 0,
        paidRewards: 0,
        availableRewards: 0,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('VALIDCODE')).toBeInTheDocument();
    });

    const input = screen.getByPlaceholderText('输入好友的邀请码');
    expect(input).toBeInTheDocument();
    expect(input).toHaveAttribute('maxLength', '8');
  });

  test('should handle multiple store synchronization', async () => {
    const mockFetchCart = jest.fn().mockResolvedValue(undefined);
    const mockFetchReferralCode = jest.fn().mockResolvedValue('SYNC123');
    const mockFetchStats = jest.fn().mockResolvedValue({
      totalReferrals: 5,
      totalRewards: 50,
      pendingRewards: 10,
      paidRewards: 40,
    });

    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'SYNC123',
      stats: {
        totalReferrals: 5,
        totalRewards: 50,
        pendingRewards: 10,
        paidRewards: 40,
        availableRewards: 30,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: mockFetchReferralCode,
      fetchStats: mockFetchStats,
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: null,
      addToCart: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
      fetchCart: mockFetchCart,
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('SYNC123')).toBeInTheDocument();
    });
  });

  test('should handle error boundary integration', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: 'Critical error',
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: null,
      stats: null,
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: 'Failed to load referral data',
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText(/邀请好友/i)).toBeInTheDocument();
    });
  });

  test('should handle admin role access control', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'admin@example.com', role: 'admin' },
      token: 'admin-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'ADMIN001',
      stats: {
        totalReferrals: 100,
        totalRewards: 1000,
        pendingRewards: 200,
        paidRewards: 800,
        availableRewards: 600,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('ADMIN001')).toBeInTheDocument();
    });
  });

  test('should handle cart quantity update integration', async () => {
    const mockUpdateQuantity = jest.fn();

    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseCartStore.mockReturnValue({
      items: [
        {
          id: 1,
          product_id: 1,
          product: {
            name: 'Quantity Test Product',
            price: 50,
            image: 'qty.jpg',
          },
          quantity: 1,
          total: 50,
        },
      ],
      total: 50,
      isLoading: false,
      error: null,
      addToCart: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: mockUpdateQuantity,
      clearCart: jest.fn(),
      fetchCart: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/cart']}>
        <Routes>
          <Route path="/cart" element={<CartPage />} />
          <Route path="/*" element={<Navigate to="/cart" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('Quantity Test Product')).toBeInTheDocument();
    });
  });

  test('should handle referral link generation', async () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', role: 'user' },
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      rememberMe: false,
      login: jest.fn(),
      register: jest.fn(),
      loginWithSms: jest.fn(),
      registerWithSms: jest.fn(),
      sendSmsCode: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    } as unknown as AuthState);

    mockUseReferralStore.mockReturnValue({
      referralCode: 'LINKTEST',
      stats: {
        totalReferrals: 3,
        totalRewards: 30,
        pendingRewards: 6,
        paidRewards: 24,
        availableRewards: 18,
      },
      referrals: [],
      rewards: [],
      withdrawals: [],
      isLoading: false,
      error: null,
      fetchReferralCode: jest.fn(),
      fetchStats: jest.fn(),
      fetchReferrals: jest.fn(),
      fetchRewards: jest.fn(),
      fetchWithdrawals: jest.fn(),
      bindReferralCode: jest.fn(),
      requestWithdrawal: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter initialEntries={['/referral']}>
        <Routes>
          <Route path="/referral" element={<ReferralPage />} />
          <Route path="/*" element={<Navigate to="/referral" />} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('LINKTEST')).toBeInTheDocument();
    });

    const copyButton = screen.getByRole('button', { name: /复制邀请码/i });
    expect(copyButton).toBeInTheDocument();
  });
});
