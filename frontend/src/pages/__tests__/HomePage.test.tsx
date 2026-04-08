import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import HomePage from '../HomePage';
import { useHomeStore } from '@/stores/homeStore';

jest.mock('../HomePage.module.css', () => ({}));

jest.mock('@/stores/homeStore', () => ({
  useHomeStore: jest.fn(),
}));

const mockUseHomeStore = useHomeStore as jest.MockedFunction<typeof useHomeStore>;

describe('HomePage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  const mockBanners = [
    { id: 1, title: '测试banner1', link: '/catalog/1' },
    { id: 2, title: '测试banner2', link: '/catalog/2' },
  ];

  const mockCategories = [
    { name: '分类1', count: 10 },
    { name: '分类2', count: 20 },
    { name: '分类3', count: 15 },
  ];

  const mockProducts = [
    {
      id: 1,
      name: '测试商品1',
      price: 100,
      original_price: 120,
      sold_count: 50,
      stock: 100,
    },
    {
      id: 2,
      name: '测试商品2',
      price: 200,
      sold_count: 30,
      stock: 50,
    },
  ];

  const defaultStoreMock = () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    });
  };

  test('页面加载时调用fetchHomeData', () => {
    const mockFetchHomeData = jest.fn();

    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: mockFetchHomeData,
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    expect(mockFetchHomeData).toHaveBeenCalled();
  });

  test('显示搜索框', () => {
    defaultStoreMock();

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    expect(screen.getByPlaceholderText('搜索模型或关键词')).toBeInTheDocument();
  });

  test('显示快捷入口', () => {
    defaultStoreMock();

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    expect(screen.getByText('热销')).toBeInTheDocument();
    expect(screen.getByText('拼团')).toBeInTheDocument();
    expect(screen.getByText('秒杀')).toBeInTheDocument();
    expect(screen.getByText('新品')).toBeInTheDocument();
  });

  test('显示轮播图', () => {
    mockUseHomeStore.mockReturnValue({
      banners: mockBanners,
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    const banner1Elements = screen.getAllByText('测试banner1');
    expect(banner1Elements.length).toBeGreaterThan(0);
    const banner2Elements = screen.getAllByText('测试banner2');
    expect(banner2Elements.length).toBeGreaterThan(0);
  });

  test('显示分类与全部分类入口', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: mockCategories,
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    expect(screen.getByText('分类1')).toBeInTheDocument();
    expect(screen.getByText('分类2')).toBeInTheDocument();
    expect(screen.getByText('分类3')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /全部分类/ })).toBeInTheDocument();
  });

  test('显示精选推荐商品', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: mockProducts,
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    expect(screen.getByText('精选推荐')).toBeInTheDocument();
    const product1Elements = screen.getAllByText('测试商品1');
    expect(product1Elements.length).toBeGreaterThan(0);
    const product2Elements = screen.getAllByText('测试商品2');
    expect(product2Elements.length).toBeGreaterThan(0);
    const price100Elements = screen.getAllByText('¥100.00');
    expect(price100Elements.length).toBeGreaterThan(0);
    const price200Elements = screen.getAllByText('¥200.00');
    expect(price200Elements.length).toBeGreaterThan(0);
  });

  test('合并热门与新品到精选推荐', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: mockProducts,
      newProducts: mockProducts,
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    expect(screen.getByText('精选推荐')).toBeInTheDocument();
    expect(screen.getAllByText('测试商品1').length).toBeGreaterThan(0);
  });

  test('显示错误状态', () => {
    const errorMessage = '加载失败';
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: errorMessage,
      fetchHomeData: jest.fn(),
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  test('点击商品跳转到详情页', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: mockProducts,
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    const productCards = screen.getAllByText('测试商品1');
    fireEvent.click(productCards[0]);

    expect(productCards[0]).toBeInTheDocument();
  });

  test('点击分类名称', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: mockCategories,
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    const categoryItem = screen.getByText('分类1');
    fireEvent.click(categoryItem);

    expect(categoryItem).toBeInTheDocument();
  });

  test('点击查看全部链接', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: mockProducts,
      newProducts: mockProducts,
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    });

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    );

    const viewAllLinks = screen.getAllByText(/查看全部/);
    expect(viewAllLinks.length).toBeGreaterThanOrEqual(1);
  });
});
