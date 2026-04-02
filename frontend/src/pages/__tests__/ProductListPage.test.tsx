import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import ProductListPage from '../ProductListPage';
import { useAuthStore } from '@/stores/authStore';
import { useCartStore } from '@/stores/cartStore';
import { skuService } from '@/services/sku';

jest.mock('@/stores/authStore');
jest.mock('@/stores/cartStore');
jest.mock('@/services/sku', () => ({
  skuService: {
    getPublicSKUs: jest.fn(),
  },
}));

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: jest.fn(),
}));

const mockUseNavigate = jest.requireMock('react-router-dom').useNavigate as jest.Mock;
const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>;
const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>;
const mockGetPublicSKUs = skuService.getPublicSKUs as jest.Mock;

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  Table: jest.fn(({ columns, dataSource, rowKey, locale }) => (
    <div data-testid="sku-table">
      {dataSource.length > 0 ? (
        dataSource.map((item: Record<string, unknown>) => (
          <div key={String(item[rowKey as string])} data-testid={`sku-${item.id}`}>
            {columns.map((col: { key?: string; render?: unknown; dataIndex?: string }) => (
              <div key={col.key || col.dataIndex}>
                {typeof col.render === 'function'
                  ? (col.render as (...args: unknown[]) => React.ReactNode)(
                      item[col.dataIndex as string],
                      item
                    )
                  : item[col.dataIndex as string]}
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
            const input = e.currentTarget.parentElement?.querySelector('input') as {
              value?: string;
            } | null;
            onSearch(input?.value ?? '');
          }}
        >
          搜索
        </button>
      </div>
    )),
  },
  Select: jest.fn(() => null),
  Button: jest.fn(({ type, children, onClick }) => (
    <button data-testid="button" onClick={onClick} className={type === 'primary' ? 'primary' : ''}>
      {children}
    </button>
  )),
  Row: jest.fn(({ children }) => <div data-testid="row">{children}</div>),
  Col: jest.fn(({ children }) => <div data-testid="col">{children}</div>),
  Tag: jest.fn(({ children }) => <span data-testid="tag">{children}</span>),
  Empty: jest.fn(({ description }) => <div data-testid="empty">{description}</div>),
  Spin: jest.fn(({ spinning, children }) => (
    <div data-testid="spin" data-spinning={spinning}>
      {children}
    </div>
  )),
  Pagination: jest.fn(({ current, pageSize, onChange }) => (
    <div data-testid="pagination">
      <button type="button" onClick={() => onChange(current + 1, pageSize)}>
        下一页
      </button>
    </div>
  )),
  Space: jest.fn(({ children }) => <div data-testid="space">{children}</div>),
  Badge: jest.fn(({ children }) => <div data-testid="badge">{children}</div>),
  FloatButton: jest.fn(() => null),
}));

describe('ProductListPage (SKU catalog)', () => {
  const mockNavigate = jest.fn();

  const baseSku = {
    id: 1,
    spu_id: 10,
    sku_code: 'T-1',
    sku_type: 'token_pack' as const,
    is_unlimited: false,
    valid_days: 365,
    retail_price: 100,
    stock: 50,
    group_enabled: false,
    min_group_size: 2,
    max_group_size: 5,
    is_trial: false,
    status: 'active',
    is_promoted: false,
    sales_count: 12,
    created_at: '',
    updated_at: '',
    spu_name: '测试套餐A',
    model_provider: 'zhipu',
    model_name: 'GLM',
    model_tier: 'lite',
    token_amount: 100000,
  };

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseNavigate.mockReturnValue(mockNavigate);
    mockGetPublicSKUs.mockResolvedValue({
      data: { data: [baseSku], total: 1, page: 1, per_page: 20 },
    });
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

  test('shows search and loads SKUs', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getByPlaceholderText(/搜索 SKU/)).toBeInTheDocument();
    await waitFor(() => {
      expect(mockGetPublicSKUs).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(screen.getByTestId('sku-1')).toBeInTheDocument();
    });
  });

  test('search navigates with q param', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    const searchInput = screen.getByTestId('search-input-field');
    fireEvent.change(searchInput, { target: { value: 'glm' } });
    await act(async () => {
      fireEvent.click(screen.getByText('搜索'));
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
      const path = mockNavigate.mock.calls[0][0] as string;
      expect(path).toContain('q=glm');
    });
  });

  test('pagination requests next page', async () => {
    mockGetPublicSKUs.mockResolvedValue({
      data: { data: [baseSku], total: 50, page: 1, per_page: 20 },
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    await waitFor(() => expect(mockGetPublicSKUs).toHaveBeenCalled());

    await act(async () => {
      fireEvent.click(screen.getByText('下一页'));
    });

    await waitFor(() => {
      expect(mockGetPublicSKUs).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 2,
        })
      );
    });
  });

  test('clicks SKU name navigates to detail', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <ProductListPage />
        </MemoryRouter>
      );
    });

    await waitFor(() => {
      expect(screen.getByText('测试套餐A')).toBeInTheDocument();
    });

    await act(async () => {
      fireEvent.click(screen.getByText('测试套餐A'));
    });

    expect(mockNavigate).toHaveBeenCalledWith('/catalog/1');
  });
});
