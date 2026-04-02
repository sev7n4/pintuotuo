import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { MemoryRouter, useNavigate } from 'react-router-dom';
import ProductListPage from '../ProductListPage';
import { useProductStore } from '@/stores/productStore';
import { useAuthStore } from '@/stores/authStore';

// 模拟 useProductStore
jest.mock('@/stores/productStore');

// 模拟 useAuthStore
jest.mock('@/stores/authStore');

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>;

// 模拟 useNavigate
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: jest.fn(),
}));

// 模拟 antd
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  Table: jest.fn(({ columns, dataSource, rowKey, locale }) => (
    <div data-testid="product-table">
      {dataSource.length > 0 ? (
        dataSource.map((item: any) => (
          <div key={item[rowKey]} data-testid={`product-${item.id}`}>
            {columns.map((col: any) => (
              <div key={col.key}>
                {col.render ? col.render(item[col.dataIndex], item) : item[col.dataIndex]}
              </div>
            ))}
          </div>
        ))
      ) : (
        <div data-testid="empty-text">{locale?.emptyText || '暂无数据'}</div>
      )}
    </div>
  )),
  Input: {
    Search: jest.fn(({ placeholder, onSearch }) => (
      <div data-testid="search-input">
        <input type="text" placeholder={placeholder} data-testid="search-input-field" />
        <button
          type="button"
          onClick={(e) => {
            const input = e.currentTarget.parentElement?.querySelector('input') as { value?: string } | null;
            onSearch(input?.value ?? '');
          }}
        >
          搜索
        </button>
      </div>
    )),
  },
  Button: jest.fn(({ type, children, onClick }) => (
    <button data-testid="button" onClick={onClick} className={type === 'primary' ? 'primary' : ''}>
      {children}
    </button>
  )),
  Row: jest.fn(({ children }) => <div data-testid="row">{children}</div>),
  Col: jest.fn(({ children }) => <div data-testid="col">{children}</div>),
  Tag: jest.fn(({ color, children }) => (
    <span data-testid="tag" style={{ color }}>
      {children}
    </span>
  )),
  Empty: jest.fn(({ description }) => <div data-testid="empty">{description}</div>),
  Spin: jest.fn(({ spinning, children }) => (
    <div data-testid="spin" data-spinning={spinning}>
      {children}
    </div>
  )),
  Pagination: jest.fn(({ current, pageSize, onChange }) => (
    <div data-testid="pagination">
      <button onClick={() => onChange(current - 1, pageSize)}>上一页</button>
      <span>第 {current} 页</span>
      <button onClick={() => onChange(current + 1, pageSize)}>下一页</button>
    </div>
  )),
  Space: jest.fn(({ children }) => <div data-testid="space">{children}</div>),
}));

const mockUseProductStore = useProductStore as jest.MockedFunction<typeof useProductStore>;
const mockUseNavigate = useNavigate as jest.MockedFunction<typeof useNavigate>;

describe('ProductListPage Component', () => {
  const mockNavigate = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseNavigate.mockReturnValue(mockNavigate);
    // 默认模拟普通用户
    mockUseAuthStore.mockReturnValue({
      user: { id: 1, email: 'user@example.com', name: 'Test User', role: 'user' },
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
      setRememberMe: jest.fn(),
    });
  });

  test('renders ProductListPage with search input for regular user', async () => {
    const mockProducts = [
      {
        id: 1,
        name: '测试产品1',
        description: '测试描述1',
        price: 100,
        stock: 50,
        status: 'active',
      },
    ];

    // 模拟 store 状态
    mockUseProductStore.mockReturnValue({
      products: mockProducts,
      total: 1,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: null,
      fetchProducts: jest.fn().mockResolvedValue(mockProducts),
      setFilters: jest.fn(),
      searchProducts: jest.fn().mockResolvedValue(mockProducts),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 检查页面元素 - 普通用户不应该看到发布产品按钮
    expect(screen.getByPlaceholderText('搜索产品...')).toBeInTheDocument();
    expect(screen.queryByText('发布产品')).not.toBeInTheDocument();
  });

  test('renders ProductListPage with publish button for merchant user', async () => {
    const mockProducts = [
      {
        id: 1,
        name: '测试产品1',
        description: '测试描述1',
        price: 100,
        stock: 50,
        status: 'active',
      },
    ];

    // 模拟商家用户
    mockUseAuthStore.mockReturnValue({
      user: { id: 2, email: 'merchant@example.com', name: 'Merchant User', role: 'merchant' },
      token: 'merchant-token',
      isLoading: false,
      error: null,
      isAuthenticated: true,
      login: jest.fn(),
      register: jest.fn(),
      logout: jest.fn(),
      fetchUser: jest.fn(),
      setUser: jest.fn(),
      clearError: jest.fn(),
      setRememberMe: jest.fn(),
    });

    // 模拟 store 状态
    mockUseProductStore.mockReturnValue({
      products: mockProducts,
      total: 1,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: null,
      fetchProducts: jest.fn().mockResolvedValue(mockProducts),
      setFilters: jest.fn(),
      searchProducts: jest.fn().mockResolvedValue(mockProducts),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 检查页面元素 - 商家用户应该看到发布产品按钮
    expect(screen.getByPlaceholderText('搜索产品...')).toBeInTheDocument();
    expect(screen.getByText('发布产品')).toBeInTheDocument();
  });

  test('shows loading state when fetching products', async () => {
    const mockProducts = [
      {
        id: 1,
        name: '测试产品1',
        description: '测试描述1',
        price: 100,
        stock: 50,
        status: 'active',
      },
    ];

    // 模拟加载状态
    mockUseProductStore.mockReturnValue({
      products: mockProducts,
      total: 1,
      filters: { page: 1, per_page: 10 },
      isLoading: true,
      error: null,
      fetchProducts: jest.fn().mockResolvedValue(mockProducts),
      setFilters: jest.fn(),
      searchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 检查加载状态
    expect(screen.getByTestId('spin')).toBeInTheDocument();
  });

  test('shows error message when there is an error', async () => {
    // 模拟错误状态
    mockUseProductStore.mockReturnValue({
      products: [],
      total: 0,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: '加载失败',
      fetchProducts: jest.fn(),
      setFilters: jest.fn(),
      searchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 检查错误信息
    expect(screen.getByText('错误: 加载失败')).toBeInTheDocument();
  });

  test('shows empty state when no products', async () => {
    // 模拟无产品状态
    mockUseProductStore.mockReturnValue({
      products: [],
      total: 0,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: null,
      fetchProducts: jest.fn().mockResolvedValue([]),
      setFilters: jest.fn(),
      searchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 检查空状态
    expect(screen.getByText('暂无数据')).toBeInTheDocument();
  });

  test('renders products list when products exist', async () => {
    const mockProducts = [
      {
        id: 1,
        name: '测试产品1',
        description: '测试描述1',
        price: 100,
        stock: 50,
        status: 'active',
      },
      {
        id: 2,
        name: '测试产品2',
        description: '测试描述2',
        price: 200,
        stock: 30,
        status: 'inactive',
      },
    ];

    // 模拟有产品状态
    mockUseProductStore.mockReturnValue({
      products: mockProducts,
      total: 2,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: null,
      fetchProducts: jest.fn().mockResolvedValue(mockProducts),
      setFilters: jest.fn(),
      searchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 检查产品列表
    expect(screen.getByTestId('product-table')).toBeInTheDocument();
    expect(screen.getByTestId('product-1')).toBeInTheDocument();
    expect(screen.getByTestId('product-2')).toBeInTheDocument();
  });

  test('handles search functionality', async () => {
    const mockProducts = [
      {
        id: 1,
        name: '测试产品1',
        description: '测试描述1',
        price: 100,
        stock: 50,
        status: 'active',
      },
    ];

    const mockSearchProducts = jest.fn().mockResolvedValue(mockProducts);
    const mockFetchProducts = jest.fn().mockResolvedValue(mockProducts);

    // 模拟 store 状态
    mockUseProductStore.mockReturnValue({
      products: mockProducts,
      total: 1,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: null,
      fetchProducts: mockFetchProducts,
      setFilters: jest.fn(),
      searchProducts: mockSearchProducts,
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 输入搜索关键词并点击搜索
    const searchInput = screen.getByTestId('search-input-field') as HTMLInputElement;
    const searchButton = screen.getByText('搜索');

    fireEvent.change(searchInput, { target: { value: '测试' } });
    await act(async () => {
      fireEvent.click(searchButton);
    });

    // 卖场搜索通过 URL ?search= 触发，由 useEffect 再调 searchProducts
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
      const path = mockNavigate.mock.calls[0][0] as string;
      expect(path).toContain('search=');
      expect(decodeURIComponent(path.split('search=')[1] || '')).toBe('测试');
    });
  });

  test('handles pagination', async () => {
    const mockProducts = [
      {
        id: 1,
        name: '测试产品1',
        description: '测试描述1',
        price: 100,
        stock: 50,
        status: 'active',
      },
    ];

    const mockSetFilters = jest.fn();
    const mockFetchProducts = jest.fn().mockResolvedValue(mockProducts);

    // 模拟 store 状态
    mockUseProductStore.mockReturnValue({
      products: mockProducts,
      total: 10,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: null,
      fetchProducts: mockFetchProducts,
      setFilters: mockSetFilters,
      searchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 点击下一页
    const nextPageButton = screen.getByText('下一页');
    await act(async () => {
      fireEvent.click(nextPageButton);
    });

    // 验证分页函数被调用
    await waitFor(() => {
      expect(mockSetFilters).toHaveBeenCalledWith({ page: 2, per_page: 10 });
      expect(mockFetchProducts).toHaveBeenCalledWith({ page: 2, per_page: 10 });
    });
  });

  test('navigates to product detail page', async () => {
    const mockProducts = [
      {
        id: 1,
        name: '测试产品1',
        description: '测试描述1',
        price: 100,
        stock: 50,
        status: 'active',
      },
    ];

    // 模拟 store 状态
    mockUseProductStore.mockReturnValue({
      products: mockProducts,
      total: 1,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: null,
      fetchProducts: jest.fn().mockResolvedValue(mockProducts),
      setFilters: jest.fn(),
      searchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 点击产品名称
    const productName = screen.getByText('测试产品1');
    await act(async () => {
      fireEvent.click(productName);
    });

    // 验证导航函数被调用
    expect(mockNavigate).toHaveBeenCalledWith('/catalog/1');
  });

  test('navigates to add to cart page', async () => {
    const mockProducts = [
      {
        id: 1,
        name: '测试产品1',
        description: '测试描述1',
        price: 100,
        stock: 50,
        status: 'active',
      },
    ];

    // 模拟 store 状态
    mockUseProductStore.mockReturnValue({
      products: mockProducts,
      total: 1,
      filters: { page: 1, per_page: 10 },
      isLoading: false,
      error: null,
      fetchProducts: jest.fn().mockResolvedValue(mockProducts),
      setFilters: jest.fn(),
      searchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    // 点击加购按钮
    const addToCartButton = screen.getByText('加购');
    await act(async () => {
      fireEvent.click(addToCartButton);
    });

    // 验证导航函数被调用
    expect(mockNavigate).toHaveBeenCalledWith('/cart');
  });
});
