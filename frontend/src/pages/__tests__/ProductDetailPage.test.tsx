import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter, useParams } from 'react-router-dom';
import ProductDetailPage from '../ProductDetailPage';
import { useProductStore } from '@/stores/productStore';
import { useCartStore } from '@/stores/cartStore';
import { useGroupStore } from '@/stores/groupStore';

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useParams: jest.fn(),
  useNavigate: () => jest.fn(),
}));

jest.mock('@/stores/productStore', () => ({
  useProductStore: jest.fn(),
}));

jest.mock('@/stores/cartStore', () => ({
  useCartStore: jest.fn(),
}));

jest.mock('@/stores/groupStore', () => ({
  useGroupStore: jest.fn(),
}));

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}));

const mockUseParams = useParams as jest.MockedFunction<typeof useParams>;
const mockUseProductStore = useProductStore as jest.MockedFunction<typeof useProductStore>;
const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>;
const mockUseGroupStore = useGroupStore as jest.MockedFunction<typeof useGroupStore>;

describe('ProductDetailPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseParams.mockReturnValue({ id: '1' });
  });

  const mockProduct = {
    id: 1,
    merchant_id: 1,
    name: '测试商品',
    description: '这是一个测试商品',
    price: 100,
    stock: 50,
    sold_count: 10,
    category: '测试分类',
    status: 'active' as const,
    created_at: '2026-01-01',
    updated_at: '2026-01-01',
    token_count: 1000000,
    models: ['GLM-5', 'K2.5'],
    validity_period: '1年',
    context_length: '128K',
    rating: 4.8,
    review_count: 1000,
    group_prices: [
      { min_members: 2, price_per_person: 60, discount_percent: 40 },
      { min_members: 5, price_per_person: 50, discount_percent: 50 },
    ],
  };

  const defaultStoreMocks = () => {
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn().mockResolvedValue(mockProduct),
      isLoading: false,
      error: null,
      products: [],
      total: 0,
      fetchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
      items: [],
      total: 0,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    });

    mockUseGroupStore.mockReturnValue({
      groups: [],
      currentGroup: null,
      total: 0,
      isLoading: false,
      error: null,
      fetchGroups: jest.fn(),
      fetchGroupByID: jest.fn(),
      createGroup: jest
        .fn()
        .mockResolvedValue({
          id: 1,
          product_id: 1,
          creator_id: 1,
          target_count: 2,
          current_count: 1,
          status: 'active',
          deadline: new Date().toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }),
      joinGroup: jest.fn(),
      cancelGroup: jest.fn(),
      getGroupProgress: jest.fn(),
      clearError: jest.fn(),
    });
  };

  test('显示加载状态', () => {
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn(),
      isLoading: true,
      error: null,
      products: [],
      total: 0,
      fetchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
      items: [],
      total: 0,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    });

    mockUseGroupStore.mockReturnValue({
      groups: [],
      currentGroup: null,
      total: 0,
      isLoading: false,
      error: null,
      fetchGroups: jest.fn(),
      fetchGroupByID: jest.fn(),
      createGroup: jest.fn(),
      joinGroup: jest.fn(),
      cancelGroup: jest.fn(),
      getGroupProgress: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    const spinner = document.querySelector('.ant-spin');
    expect(spinner).toBeInTheDocument();
  });

  test('显示错误状态', () => {
    const errorMessage = '加载失败';
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn(),
      isLoading: false,
      error: errorMessage,
      products: [],
      total: 0,
      fetchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
      items: [],
      total: 0,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    });

    mockUseGroupStore.mockReturnValue({
      groups: [],
      currentGroup: null,
      total: 0,
      isLoading: false,
      error: null,
      fetchGroups: jest.fn(),
      fetchGroupByID: jest.fn(),
      createGroup: jest.fn(),
      joinGroup: jest.fn(),
      cancelGroup: jest.fn(),
      getGroupProgress: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    expect(screen.getByText(`错误: ${errorMessage}`)).toBeInTheDocument();
  });

  test('显示产品详情和定价信息', async () => {
    defaultStoreMocks();

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument();
    });

    expect(screen.getByText('这是一个测试商品')).toBeInTheDocument();
    expect(screen.getByText('定价信息')).toBeInTheDocument();
    expect(screen.getByText('单独购买')).toBeInTheDocument();
    expect(screen.getByText('拼团购买')).toBeInTheDocument();
  });

  test('显示拼团规则选项', async () => {
    defaultStoreMocks();

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument();
    });

    expect(screen.getByText('选择拼团规则')).toBeInTheDocument();
    expect(screen.getByText('2人团')).toBeInTheDocument();
    expect(screen.getByText('5人团')).toBeInTheDocument();
  });

  test('显示商品详情Tab', async () => {
    defaultStoreMocks();

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument();
    });

    expect(screen.getByText('商品详情')).toBeInTheDocument();
    expect(screen.getByText(/用户评价/)).toBeInTheDocument();
  });

  test('显示用户评价', async () => {
    defaultStoreMocks();

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument();
    });

    const reviewsTab = screen.getByText(/用户评价/);
    fireEvent.click(reviewsTab);

    await waitFor(() => {
      expect(screen.getByText('综合评分')).toBeInTheDocument();
    });
  });

  test('点击拼团按钮创建拼团', async () => {
    const mockCreateGroup = jest
      .fn()
      .mockResolvedValue({
        id: 1,
        product_id: 1,
        creator_id: 1,
        target_count: 2,
        current_count: 1,
        status: 'active',
        deadline: new Date().toISOString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      });

    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn().mockResolvedValue(mockProduct),
      isLoading: false,
      error: null,
      products: [],
      total: 0,
      fetchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
      items: [],
      total: 0,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    });

    mockUseGroupStore.mockReturnValue({
      groups: [],
      currentGroup: null,
      total: 0,
      isLoading: false,
      error: null,
      fetchGroups: jest.fn(),
      fetchGroupByID: jest.fn(),
      createGroup: mockCreateGroup,
      joinGroup: jest.fn(),
      cancelGroup: jest.fn(),
      getGroupProgress: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument();
    });

    const groupButton = screen.getByRole('button', { name: /立即拼团/ });
    fireEvent.click(groupButton);

    await waitFor(() => {
      expect(mockCreateGroup).toHaveBeenCalled();
    });
  });

  test('产品无库存时显示暂无库存', async () => {
    const outOfStockProduct = {
      ...mockProduct,
      stock: 0,
    };

    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn().mockResolvedValue(outOfStockProduct),
      isLoading: false,
      error: null,
      products: [],
      total: 0,
      fetchProducts: jest.fn(),
      createProduct: jest.fn(),
      updateProduct: jest.fn(),
      deleteProduct: jest.fn(),
      clearError: jest.fn(),
    });

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
      items: [],
      total: 0,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    });

    mockUseGroupStore.mockReturnValue({
      groups: [],
      currentGroup: null,
      total: 0,
      isLoading: false,
      error: null,
      fetchGroups: jest.fn(),
      fetchGroupByID: jest.fn(),
      createGroup: jest.fn(),
      joinGroup: jest.fn(),
      cancelGroup: jest.fn(),
      getGroupProgress: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('暂无库存')).toBeInTheDocument();
    });
  });

  test('返回按钮存在', async () => {
    defaultStoreMocks();

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument();
    });

    const backButton = screen.getByText('返回列表');
    expect(backButton).toBeInTheDocument();
  });
});
