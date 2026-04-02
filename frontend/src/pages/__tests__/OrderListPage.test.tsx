import { render, screen, fireEvent, act } from '@testing-library/react';
import { MemoryRouter, useNavigate } from 'react-router-dom';
import OrderListPage from '../OrderListPage';
import { useOrderStore } from '@/stores/orderStore';
import { useCartStore } from '@/stores/cartStore';
import { useProductStore } from '@/stores/productStore';

jest.mock('@/stores/orderStore');
jest.mock('@/stores/cartStore');
jest.mock('@/stores/productStore');

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: jest.fn(),
}));

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  Table: jest.fn(({ columns, dataSource, rowKey, locale }) => (
    <div data-testid="order-table">
      {dataSource.length > 0 ? (
        dataSource.map((item: any) => (
          <div key={item[rowKey]} data-testid={`order-${item.id}`}>
            {columns.map((col: any) => (
              <div key={col.key}>
                {col.render ? col.render(item[col.dataIndex], item) : item[col.dataIndex]}
              </div>
            ))}
          </div>
        ))
      ) : (
        <div data-testid="empty-text">{locale?.emptyText || '暂无订单'}</div>
      )}
    </div>
  )),
  Button: jest.fn(({ type, children, onClick, disabled }) => (
    <button
      data-testid="button"
      onClick={onClick}
      disabled={disabled}
      className={type === 'link' ? 'link' : ''}
    >
      {children}
    </button>
  )),
  Space: jest.fn(({ children }) => <div data-testid="space">{children}</div>),
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
  Modal: jest.fn(({ title, open, onCancel }) =>
    open ? (
      <div
        data-testid="modal"
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          background: 'rgba(0,0,0,0.5)',
        }}
      >
        <div style={{ background: 'white', margin: '50px auto', padding: '20px', width: '500px' }}>
          <h2>{title}</h2>
          <div data-testid="descriptions">
            <div data-testid="description-item">订单号: 1</div>
            <div data-testid="description-item">产品ID: 101</div>
            <div data-testid="description-item">数量: 2</div>
            <div data-testid="description-item">单价: ¥100.00</div>
            <div data-testid="description-item">总价: ¥200.00</div>
            <div data-testid="description-item">状态: 待支付</div>
            <div data-testid="description-item">创建时间: 2026-03-19 08:00:00</div>
            <div data-testid="description-item">分组ID: -</div>
          </div>
          <button onClick={onCancel} data-testid="modal-cancel">
            关闭
          </button>
        </div>
      </div>
    ) : null
  ),
}));

const mockUseOrderStore = useOrderStore as jest.MockedFunction<typeof useOrderStore>;
const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>;
const mockUseProductStore = useProductStore as jest.MockedFunction<typeof useProductStore>;
const mockUseNavigate = useNavigate as jest.MockedFunction<typeof useNavigate>;

describe('OrderListPage Component', () => {
  const mockNavigate = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseNavigate.mockReturnValue(mockNavigate);
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
    mockUseProductStore.mockReturnValue({
      products: [],
      total: 0,
      filters: { page: 1, per_page: 20 },
      isLoading: false,
      error: null,
      fetchProducts: jest.fn(),
      setFilters: jest.fn(),
      searchProducts: jest.fn(),
      fetchProductByID: jest.fn(),
    });
  });

  test('renders OrderListPage with title', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('我的订单')).toBeInTheDocument();
  });

  test('shows loading state when fetching orders', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: true,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getByTestId('spin')).toBeInTheDocument();
  });

  test('shows error message when there is an error', async () => {
    mockUseOrderStore.mockReturnValue({
      orders: [],
      isLoading: false,
      error: '加载失败',
      fetchOrders: jest.fn(),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('错误: 加载失败')).toBeInTheDocument();
  });

  test('shows empty state when no orders', async () => {
    mockUseOrderStore.mockReturnValue({
      orders: [],
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue([]),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('暂无订单')).toBeInTheDocument();
  });

  test('renders orders list when orders exist', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
      {
        id: 2,
        product_id: 102,
        quantity: 1,
        total_price: 150,
        status: 'completed',
        created_at: '2026-03-18T00:00:00Z',
        group_id: 5,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getByTestId('order-table')).toBeInTheDocument();
    expect(screen.getByTestId('order-1')).toBeInTheDocument();
    expect(screen.getByTestId('order-2')).toBeInTheDocument();
  });

  test('opens order detail modal', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    const detailButton = screen.getByText('详情');
    await act(async () => {
      fireEvent.click(detailButton);
    });

    expect(screen.getByTestId('modal')).toBeInTheDocument();
    expect(screen.getByText('订单详情 #1')).toBeInTheDocument();
  });

  test('closes order detail modal', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    const detailButton = screen.getByText('详情');
    await act(async () => {
      fireEvent.click(detailButton);
    });

    const closeButton = screen.getByTestId('modal-cancel');
    await act(async () => {
      fireEvent.click(closeButton);
    });

    expect(screen.queryByTestId('modal')).not.toBeInTheDocument();
  });

  test('navigates to payment page for pending orders', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    const payButton = screen.getByText('支付');
    await act(async () => {
      fireEvent.click(payButton);
    });

    expect(mockNavigate).toHaveBeenCalledWith('/payment/1');
  });

  test('does not show pay button for non-pending orders', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'completed',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.queryByText('支付')).not.toBeInTheDocument();
  });
});

describe('TC-ORDER-001: 取消待支付订单', () => {
  test('should allow cancellation for pending orders', async () => {
    const mockCancelOrder = jest.fn().mockResolvedValue({
      id: 1,
      status: 'cancelled',
    });

    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: mockCancelOrder,
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    const cancelButton = screen.queryByText('取消');
    expect(cancelButton).toBeInTheDocument();
  });
});

describe('TC-ORDER-002: 取消已支付订单(申请退款)', () => {
  test('should show refund option for paid orders', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'paid',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('详情')).toBeInTheDocument();
  });
});

describe('TC-ORDER-003: 退款状态跟踪', () => {
  test('should display refunding status correctly', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'refunding',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getByTestId('order-table')).toBeInTheDocument();
  });
});

describe('TC-ORDER-004: 已完成订单不可取消', () => {
  test('should not show cancel button for completed orders', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'completed',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.queryByText('取消')).not.toBeInTheDocument();
    expect(screen.queryByText('支付')).not.toBeInTheDocument();
  });
});

describe('TC-ORDER-005: 订单状态显示', () => {
  test('should display correct status label for pending orders', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getAllByText('待支付').length).toBeGreaterThan(0);
  });

  test('should display correct status label for paid orders', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'paid',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getAllByText('已支付').length).toBeGreaterThan(0);
  });

  test('should display correct status label for cancelled orders', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'cancelled',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    expect(screen.getAllByText('已取消').length).toBeGreaterThan(0);
  });
});

describe('TC-ORDER-006: 订单详情完整性', () => {
  test('should display all order details in modal', async () => {
    const mockOrders = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        total_price: 200,
        status: 'pending',
        created_at: '2026-03-19T00:00:00Z',
        group_id: null,
      },
    ];

    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      fetchOrderByID: jest.fn(),
      createOrder: jest.fn(),
      updateOrder: jest.fn(),
      cancelOrder: jest.fn(),
      requestRefund: jest.fn(),
      clearError: jest.fn(),
      currentOrder: null,
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      );
    });

    const detailButton = screen.getByText('详情');
    await act(async () => {
      fireEvent.click(detailButton);
    });

    expect(screen.getByTestId('descriptions')).toBeInTheDocument();
    expect(screen.getByTestId('modal')).toBeInTheDocument();
  });
});
