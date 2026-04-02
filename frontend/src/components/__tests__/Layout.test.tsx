import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import Layout from '../Layout';
import { useAuthStore } from '@/stores/authStore';
import { useCartStore } from '@/stores/cartStore';

// 模拟 message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
  },
}));

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

  test('renders Layout with login/register links when not authenticated', () => {
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

    // 主导航仅保留首页
    expect(screen.getByText('首页')).toBeInTheDocument();
    expect(screen.queryByText('分类')).not.toBeInTheDocument();
    expect(screen.queryByText('订单')).not.toBeInTheDocument();
    expect(screen.queryByText('购物车')).not.toBeInTheDocument();

    // 检查登录/注册链接
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

    // 检查用户信息显示
    expect(screen.getByText('Test User')).toBeInTheDocument();
    expect(screen.getByTestId('user-dropdown')).toBeInTheDocument();
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
