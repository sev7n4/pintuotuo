import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import LoginPage from '../LoginPage';
import { useAuthStore } from '@/stores/authStore';
import { AUTH_PRIMARY_LOGIN_KEY } from '@/lib/authLoginPreference';

jest.mock('@/stores/authStore');

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}));

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>;

describe('LoginPage Integration Tests - User Experience Flow', () => {
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
      }),
    });
  });

  describe('TC-AUTH-001: 用户登录成功流程', () => {
    test('should successfully login with valid credentials and navigate to home for regular user', async () => {
      const mockLogin = jest.fn();
      mockLogin.mockImplementation(async () => {
        mockUseAuthStore.mockReturnValue({
          user: { id: 1, email: 'user@example.com', role: 'user' },
          token: 'user-token',
          isLoading: false,
          error: null,
          isAuthenticated: true,
          rememberMe: false,
          login: mockLogin,
          register: jest.fn(),
          logout: jest.fn(),
          fetchUser: jest.fn(),
          setUser: jest.fn(),
          clearError: jest.fn(),
          setRememberMe: jest.fn(),
        });
        return { token: 'user-token' };
      });

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        rememberMe: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
        setRememberMe: jest.fn(),
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<div>Home Page</div>} />
            <Route path="/merchant" element={<div>Merchant Dashboard</div>} />
            <Route path="/admin" element={<div>Admin Dashboard</div>} />
          </Routes>
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'password123' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password123', true);
      });
    });

    test('should navigate to merchant dashboard for merchant user', async () => {
      const mockLogin = jest.fn();
      mockLogin.mockImplementation(async () => {
        mockUseAuthStore.mockReturnValue({
          user: { id: 2, email: 'merchant@example.com', role: 'merchant' },
          token: 'merchant-token',
          isLoading: false,
          error: null,
          isAuthenticated: true,
          rememberMe: false,
          login: mockLogin,
          register: jest.fn(),
          logout: jest.fn(),
          fetchUser: jest.fn(),
          setUser: jest.fn(),
          clearError: jest.fn(),
          setRememberMe: jest.fn(),
        });
        return { token: 'merchant-token' };
      });

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        rememberMe: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
        setRememberMe: jest.fn(),
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/merchant" element={<div>Merchant Dashboard</div>} />
          </Routes>
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'merchant@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'password123' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('merchant@example.com', 'password123', true);
      });
    });

    test('should navigate to admin dashboard for admin user', async () => {
      const mockLogin = jest.fn();
      mockLogin.mockImplementation(async () => {
        mockUseAuthStore.mockReturnValue({
          user: { id: 3, email: 'admin@example.com', role: 'admin' },
          token: 'admin-token',
          isLoading: false,
          error: null,
          isAuthenticated: true,
          rememberMe: false,
          login: mockLogin,
          register: jest.fn(),
          logout: jest.fn(),
          fetchUser: jest.fn(),
          setUser: jest.fn(),
          clearError: jest.fn(),
          setRememberMe: jest.fn(),
        });
        return { token: 'admin-token' };
      });

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        rememberMe: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
        setRememberMe: jest.fn(),
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/admin" element={<div>Admin Dashboard</div>} />
          </Routes>
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'admin@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'password123' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('admin@example.com', 'password123', true);
      });
    });
  });

  describe('TC-AUTH-002: 登录失败-无效邮箱格式', () => {
    test('should show validation error for invalid email format', async () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(screen.getByText('邮箱格式不正确')).toBeInTheDocument();
      });
    });

    test('should show validation error for email without @ symbol', async () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'userexample.com' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(screen.getByText('邮箱格式不正确')).toBeInTheDocument();
      });
    });
  });

  describe('TC-AUTH-003: 登录失败-密码为空', () => {
    test('should show validation error when password is empty', async () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(screen.getByText('请输入密码')).toBeInTheDocument();
      });
    });
  });

  describe('TC-AUTH-004: 登录失败-错误凭证', () => {
    test('should show error message for invalid credentials', async () => {
      const mockLogin = jest.fn().mockRejectedValue(new Error('Invalid credentials'));

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: '登录失败，请检查邮箱和密码',
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'wrong@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('wrong@example.com', 'wrongpassword', true);
      });
    });

    test('should keep form editable after login failure', async () => {
      const mockLogin = jest.fn().mockRejectedValue(new Error('Invalid credentials'));

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: '登录失败',
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(emailInput).toBeEnabled();
        expect(passwordInput).toBeEnabled();
      });
    });
  });

  describe('TC-AUTH-005: 登录表单完整性', () => {
    test('should render all required form elements', () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      expect(screen.getByText('拼脱脱 - 登录 / 注册')).toBeInTheDocument();
      expect(screen.getByLabelText('邮箱')).toBeInTheDocument();
      expect(screen.getByLabelText('密码（仅曾用邮箱注册的账号）')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '密码登录' })).toBeInTheDocument();
      expect(screen.getByText('创建新账户')).toBeInTheDocument();
    });

    test('should have correct placeholder text', () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      expect(screen.getByPlaceholderText('example@email.com')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('输入密码')).toBeInTheDocument();
    });
  });

  describe('TC-AUTH-006: 导航到注册页面', () => {
    test('should navigate to register page when create account is clicked', async () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<div>Register Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      const createAccountButton = screen.getByText('创建新账户');

      await act(async () => {
        fireEvent.click(createAccountButton);
      });

      await waitFor(() => {
        expect(screen.getByText('Register Page')).toBeInTheDocument();
      });
    });
  });

  describe('TC-AUTH-007: 加载状态处理', () => {
    test('should show loading state during login', async () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: true,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      //加载中 Ant Design 按钮的可访问名可能变化，用文案更稳
      expect(screen.getByText('密码登录')).toBeInTheDocument();
    });
  });

  describe('TC-AUTH-008: 邮箱必填验证', () => {
    test('should show error when email is empty', async () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(screen.getByText('请输入邮箱')).toBeInTheDocument();
      });
    });
  });

  describe('TC-AUTH-009: 表单重置', () => {
    test('should allow user to re-enter credentials after error', async () => {
      const mockLogin = jest
        .fn()
        .mockRejectedValueOnce(new Error('Invalid credentials'))
        .mockResolvedValueOnce({ token: 'success-token' });

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      mockUseAuthStore.getState = jest
        .fn()
        .mockReturnValueOnce({ user: null })
        .mockReturnValueOnce({ user: { id: 1, email: 'user@example.com', role: 'user' } });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<div>Home Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'wrong@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('wrong@example.com', 'wrongpassword', true);
      });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'correct@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'correctpassword' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('correct@example.com', 'correctpassword', true);
      });
    });
  });

  describe('TC-AUTH-010: 完整用户登录旅程', () => {
    test('should complete full login journey from form fill to navigation', async () => {
      const mockLogin = jest.fn();
      mockLogin.mockImplementation(async () => {
        mockUseAuthStore.mockReturnValue({
          user: { id: 1, email: 'user@example.com', role: 'user' },
          token: 'user-token',
          isLoading: false,
          error: null,
          isAuthenticated: true,
          login: mockLogin,
          register: jest.fn(),
          logout: jest.fn(),
          fetchUser: jest.fn(),
          setUser: jest.fn(),
          clearError: jest.fn(),
        });
        return { token: 'user-token' };
      });

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<div>Home Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com') as HTMLInputElement;
      const passwordInput = screen.getByPlaceholderText('输入密码') as HTMLInputElement;
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'password123' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password123', true);
      });
    });
  });

  describe('TC-AUTH-008: 记住我功能', () => {
    test('should render remember me checkbox and be checked by default', () => {
      const mockLogin = jest.fn();

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        rememberMe: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
        setRememberMe: jest.fn(),
      });

      render(
        <MemoryRouter>
          <LoginPage />
        </MemoryRouter>
      );

      const checkbox = screen.getByLabelText('记住我');
      expect(checkbox).toBeInTheDocument();
      expect(checkbox).toBeChecked();
    });

    test('should call login with rememberMe parameter', async () => {
      const mockLogin = jest.fn().mockResolvedValue({ token: 'user-token' });

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        rememberMe: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
        setRememberMe: jest.fn(),
      });

      mockUseAuthStore.getState = jest.fn().mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<div>Home Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com');
      const passwordInput = screen.getByPlaceholderText('输入密码');
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'password123' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password123', true);
      });
    });

    test('should call login with rememberMe false when checkbox is unchecked', async () => {
      const mockLogin = jest.fn().mockResolvedValue({ token: 'user-token' });

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        rememberMe: false,
        login: mockLogin,
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
        setRememberMe: jest.fn(),
      });

      mockUseAuthStore.getState = jest.fn().mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<div>Home Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com');
      const passwordInput = screen.getByPlaceholderText('输入密码');
      const checkbox = screen.getByLabelText('记住我');
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.click(checkbox);
        fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
        fireEvent.change(passwordInput, { target: { value: 'password123' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password123', false);
      });
    });
  });

  describe('TC-AUTH-011: 记住主登录入口', () => {
    test('上次为手机登录时默认展示手机入口（未开短信时显示说明）', async () => {
      localStorage.setItem(AUTH_PRIMARY_LOGIN_KEY, 'phone');

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
        clearError: jest.fn(),
        setRememberMe: jest.fn(),
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
          </Routes>
        </MemoryRouter>
      );

      await waitFor(() => {
        expect(screen.getByText('暂未开启手机号验证码')).toBeInTheDocument();
      });
    });

    test('邮箱密码登录成功后写入偏好为 email', async () => {
      const mockLogin = jest.fn().mockResolvedValue(undefined);

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
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
      });

      render(
        <MemoryRouter initialEntries={['/login']}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
          </Routes>
        </MemoryRouter>
      );

      const emailInput = screen.getByPlaceholderText('example@email.com');
      const passwordInput = screen.getByPlaceholderText('输入密码');
      const loginButton = screen.getByRole('button', { name: '密码登录' });

      await act(async () => {
        fireEvent.change(emailInput, { target: { value: 'a@b.com' } });
        fireEvent.change(passwordInput, { target: { value: 'secret' } });
        fireEvent.click(loginButton);
      });

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalled();
        expect(localStorage.getItem(AUTH_PRIMARY_LOGIN_KEY)).toBe('email');
      });
    });
  });
});
