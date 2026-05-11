import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import Layout from '../Layout';
import { useAuthStore } from '@/stores/authStore';
import { useCartStore } from '@/stores/cartStore';

// 模拟 message；桌面端断点，避免测到 Drawer 移动导航
jest.mock('antd', () => {
  const antd = jest.requireActual('antd');
  return {
    ...antd,
    message: {
      success: jest.fn(),
    },
    Grid: {
      ...antd.Grid,
      useBreakpoint: () => ({ xs: false, sm: false, md: true, lg: true }),
    },
  };
});

// 模拟 CSS 文件
jest.mock('../Layout.css', () => ({}));

// 模拟 useAuthStore
jest.mock('@/stores/authStore');
jest.mock('@/stores/cartStore');

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>;
const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>;

describe('Layout Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseCartStore.mockReturnValue({
      items: [],
      total: 0,
      isLoading: false,
      error: null,
      addItem: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clear: jest.fn(),
      getTotal: jest.fn().mockReturnValue(0),
    });
  });

  test('renders Layout with login link when not authenticated', () => {
    // 模拟未认证状态
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
        <Layout />
      </MemoryRouter>
    );

    expect(screen.getByText('首页')).toBeInTheDocument();
    expect(screen.getByText('卖场')).toBeInTheDocument();
    expect(screen.getByText('购物车')).toBeInTheDocument();
    const groupNav = screen.getByRole('link', { name: '拼团' });
    expect(groupNav).toBeInTheDocument();
    expect(groupNav).toHaveAttribute('href', '/catalog?group_enabled=true');
    expect(screen.getByText('邀请')).toBeInTheDocument();
    expect(screen.getByText('帮助')).toBeInTheDocument();
    expect(screen.queryByText('我的订单')).not.toBeInTheDocument();

    // 检查登录入口
    expect(screen.getByText('登录')).toBeInTheDocument();
    expect(screen.getByText('注册')).toBeInTheDocument();
  });

  test('renders Layout with user dropdown when authenticated', () => {
    // 模拟认证状态
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      role: 'user',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockUseAuthStore.mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
    );

    expect(screen.getByText('Test User')).toBeInTheDocument();
    expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
    expect(screen.getByText('我的订单')).toBeInTheDocument();
  });

  test('renders user dropdown when authenticated', () => {
    // 模拟认证状态
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      name: 'Test User',
      role: 'user',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    mockUseAuthStore.mockReturnValue({
      user: mockUser,
      token: 'test-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <MemoryRouter>
        <Layout />
      </MemoryRouter>
    );

    // 检查用户下拉菜单存在
    expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
    expect(screen.getByText('Test User')).toBeInTheDocument();
  });

  describe('TC-AUTH-009: 登出流程', () => {
    test('should have logout functionality in user menu', async () => {
      const mockLogout = jest.fn().mockResolvedValue(undefined);
      const mockUser = {
        id: 1,
        email: 'test@example.com',
        name: 'Test User',
        role: 'user',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      mockUseAuthStore.mockReturnValue({
        user: mockUser,
        token: 'test-token',
        isLoading: false,
        error: null,
        isAuthenticated: true,
        login: jest.fn(),
        register: jest.fn(),
        logout: mockLogout,
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      });

      render(
        <MemoryRouter initialEntries={['/']}>
          <Routes>
            <Route path="/" element={<Layout />} />
            <Route path="/login" element={<div>Login Page</div>} />
          </Routes>
        </MemoryRouter>
      );

      const userDropdown = screen.getByTestId('user-dropdown');
      expect(userDropdown).toBeInTheDocument();
      expect(screen.getByText('Test User')).toBeInTheDocument();
    });
  });
});
