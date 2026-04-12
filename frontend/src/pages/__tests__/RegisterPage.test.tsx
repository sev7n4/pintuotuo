import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { act } from 'react';
import RegisterPage from '../RegisterPage';
import { useAuthStore } from '@/stores/authStore';

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

import { message } from 'antd';
const mockMessage = message as jest.Mocked<typeof message>;

describe('RegisterPage Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders RegisterPage with form and buyer/merchant tabs', () => {
    // 模拟 store 状态
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
        <RegisterPage />
      </MemoryRouter>
    );

    // 检查页面元素
    expect(screen.getByText('拼脱脱 - 注册')).toBeInTheDocument();
    expect(screen.getByText('买家注册')).toBeInTheDocument();
    expect(screen.getByText('商户入驻')).toBeInTheDocument();
    expect(screen.getByLabelText('邮箱')).toBeInTheDocument();
    expect(screen.getByLabelText('名字')).toBeInTheDocument();
    expect(screen.getByLabelText('密码')).toBeInTheDocument();
    expect(screen.getByLabelText('确认密码')).toBeInTheDocument();
    expect(screen.getByText('创建账户')).toBeInTheDocument();
    expect(screen.getByText('立即登录')).toBeInTheDocument();
  });

  test('submits form successfully as user', async () => {
    // 模拟注册成功
    const mockRegister = jest.fn().mockResolvedValue(undefined);

    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      login: jest.fn(),
      register: mockRegister,
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>
    );

    // 填写表单
    const emailInput = screen.getByPlaceholderText('example@email.com');
    const nameInput = screen.getByPlaceholderText('输入你的名字');
    const passwordInput = screen.getByPlaceholderText('设置密码');
    const confirmPasswordInput = screen.getByPlaceholderText('再次输入密码');
    const submitButton = screen.getByText('创建账户');

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(nameInput, { target: { value: 'Test User' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } });

    // 提交表单
    await act(async () => {
      fireEvent.click(submitButton);
    });

    // 验证注册函数被调用
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith(
        'test@example.com',
        'Test User',
        'password123',
        'user'
      );
    });

    // 验证成功消息
    expect(mockMessage.success).toHaveBeenCalledWith('注册成功');
  });

  test('submits form successfully as merchant', async () => {
    // 模拟注册成功
    const mockRegister = jest.fn().mockResolvedValue(undefined);

    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      login: jest.fn(),
      register: mockRegister,
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole('tab', { name: /商户入驻/i }));

    // 填写表单
    const emailInput = screen.getByPlaceholderText('example@email.com');
    const nameInput = screen.getByPlaceholderText('输入你的名字');
    const passwordInput = screen.getByPlaceholderText('设置密码');
    const confirmPasswordInput = screen.getByPlaceholderText('再次输入密码');
    const submitButton = screen.getByText('创建账户');

    fireEvent.change(emailInput, { target: { value: 'merchant@example.com' } });
    fireEvent.change(nameInput, { target: { value: 'Merchant User' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } });

    // 提交表单
    await act(async () => {
      fireEvent.click(submitButton);
    });

    // 验证注册函数被调用
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith(
        'merchant@example.com',
        'Merchant User',
        'password123',
        'merchant'
      );
    });

    // 验证成功消息
    expect(mockMessage.success).toHaveBeenCalledWith('注册成功');
  });

  test('handles password mismatch error', async () => {
    // 模拟 store 状态
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
        <RegisterPage />
      </MemoryRouter>
    );

    // 填写表单，密码不一致
    const emailInput = screen.getByPlaceholderText('example@email.com');
    const nameInput = screen.getByPlaceholderText('输入你的名字');
    const passwordInput = screen.getByPlaceholderText('设置密码');
    const confirmPasswordInput = screen.getByPlaceholderText('再次输入密码');
    const submitButton = screen.getByText('创建账户');

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(nameInput, { target: { value: 'Test User' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'differentpassword' } });

    // 提交表单
    await act(async () => {
      fireEvent.click(submitButton);
    });

    // 验证错误消息
    expect(mockMessage.error).toHaveBeenCalledWith('两次输入的密码不一致');
  });

  test('handles registration error', async () => {
    // 模拟注册失败
    const errorMessage = 'Email already exists';
    const mockRegister = jest.fn().mockRejectedValue(new Error(errorMessage));

    mockUseAuthStore.mockReturnValue({
      user: null,
      token: null,
      isLoading: false,
      error: errorMessage,
      isAuthenticated: false,
      login: jest.fn(),
      register: mockRegister,
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>
    );

    // 填写表单
    const emailInput = screen.getByPlaceholderText('example@email.com');
    const nameInput = screen.getByPlaceholderText('输入你的名字');
    const passwordInput = screen.getByPlaceholderText('设置密码');
    const confirmPasswordInput = screen.getByPlaceholderText('再次输入密码');
    const submitButton = screen.getByText('创建账户');

    fireEvent.change(emailInput, { target: { value: 'existing@example.com' } });
    fireEvent.change(nameInput, { target: { value: 'Test User' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } });

    // 提交表单
    await act(async () => {
      fireEvent.click(submitButton);
    });

    // 验证注册函数被调用
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith(
        'existing@example.com',
        'Test User',
        'password123',
        'user'
      );
    });

    // 验证错误消息
    expect(mockMessage.error).toHaveBeenCalledWith(errorMessage);
  });

  test('shows loading state during registration', () => {
    // 模拟加载状态
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
        <RegisterPage />
      </MemoryRouter>
    );

    // 检查创建账户按钮是否处于加载状态
    const submitButton = screen.getByText(/创建账户/);
    const buttonElement = submitButton.closest('button');
    expect(buttonElement).toHaveClass('ant-btn-loading');
  });

  test('navigates to login page when login button is clicked', () => {
    // 模拟 store 状态
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
        <RegisterPage />
      </MemoryRouter>
    );

    // 点击立即登录按钮
    const loginButton = screen.getByText('立即登录');
    fireEvent.click(loginButton);

    // 验证导航到登录页面
    expect(screen.getByText('拼脱脱 - 注册')).toBeInTheDocument();
  });

  test('switches to merchant tab and shows onboarding hint', () => {
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
        <RegisterPage />
      </MemoryRouter>
    );

    expect(screen.getByText('买家注册')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('tab', { name: /商户入驻/i }));
    expect(screen.getByText(/若仅购买 Token，请使用「买家注册」/)).toBeInTheDocument();
  });
});
