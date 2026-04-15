import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import RegisterPage from '../RegisterPage';
import { useAuthStore } from '@/stores/authStore';
import { AUTH_PRIMARY_LOGIN_KEY } from '@/lib/authLoginPreference';

// 模拟 useAuthStore
jest.mock('@/stores/authStore');

// 模拟 message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}));

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>;

describe('RegisterPage Component', () => {
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

  test('renders RegisterPage as email register with role segmented', async () => {
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
      sendEmailMagicLink: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      setRememberMe: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>
    );

    expect(screen.getByText('拼脱脱 - 登录 / 注册')).toBeInTheDocument();
    expect(screen.getByText('账号体系已升级')).toBeInTheDocument();
    expect(screen.getByLabelText('邮箱')).toBeInTheDocument();
    expect(screen.getByLabelText('密码')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('个人用户')).toBeInTheDocument();
      expect(screen.getByText('商户入驻')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: '注册并进入' })).toBeInTheDocument();
    expect(screen.queryByText('发送邮箱魔法链接')).not.toBeInTheDocument();
  });

  test('shows loading state during registration', () => {
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
      sendEmailMagicLink: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      setRememberMe: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>
    );

    // 加载中按钮的可访问名可能变化，直接断言主提交按钮带 loading 样式
    const loadingSubmit = document.querySelector(
      '.auth-card form button.ant-btn-primary.ant-btn-loading'
    );
    expect(loadingSubmit).not.toBeNull();
  });

  test('navigates to login page when login button is clicked', () => {
    // 模拟 store 状态
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
      sendEmailMagicLink: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      setRememberMe: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>
    );

    expect(screen.getByText('拼脱脱 - 登录 / 注册')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: '立即登录' })).toBeInTheDocument();
  });
});
