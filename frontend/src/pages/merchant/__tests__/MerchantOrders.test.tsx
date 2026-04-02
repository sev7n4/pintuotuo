import { render, screen, act, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

jest.mock('@/stores/merchantStore', () => ({
  useMerchantStore: jest.fn(() => ({
    orders: [
      {
        id: 1,
        product_name: 'Test Product',
        user_id: 100,
        quantity: 2,
        total_price: 199.99,
        status: 'pending',
        created_at: '2024-01-15T10:30:00Z',
        updated_at: '2024-01-15T10:30:00Z',
      },
      {
        id: 2,
        product_name: 'Another Product',
        user_id: 101,
        quantity: 1,
        total_price: 99.0,
        status: 'completed',
        created_at: '2024-01-16T14:20:00Z',
        updated_at: '2024-01-16T14:20:00Z',
      },
    ],
    fetchOrders: jest.fn(),
    isLoading: false,
  })),
}));

jest.mock('@/stores/authStore', () => ({
  useAuthStore: jest.fn(() => ({
    user: { id: 1, name: 'Test Merchant', role: 'merchant' },
  })),
}));

jest.mock('@/services/api', () => ({
  default: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

describe('MerchantOrders', () => {
  let MerchantOrders: React.FC;

  beforeEach(async () => {
    jest.clearAllMocks();
    MerchantOrders = (await import('../MerchantOrders')).default;
  });

  it('renders orders page with title', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantOrders />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('订单管理')).toBeInTheDocument();
  });

  it('displays status filter dropdown', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantOrders />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('全部状态')).toBeInTheDocument();
  });

  it('displays export button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantOrders />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('导出数据')).toBeInTheDocument();
  });

  it('displays order table with data', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantOrders />
        </MemoryRouter>
      );
    });
    const table = screen.getByRole('table');
    expect(within(table).getAllByText('订单ID').length).toBeGreaterThan(0);
    expect(within(table).getAllByText('商品名称').length).toBeGreaterThan(0);
    expect(screen.getByText('Test Product')).toBeInTheDocument();
    expect(screen.getByText('Another Product')).toBeInTheDocument();
  });

  it('displays order status tags', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantOrders />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('待支付')).toBeInTheDocument();
    expect(screen.getByText('已完成')).toBeInTheDocument();
  });
});
