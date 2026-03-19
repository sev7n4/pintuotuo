import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { MemoryRouter, useNavigate } from 'react-router-dom'
import OrderListPage from '../OrderListPage'
import { useOrderStore } from '@/stores/orderStore'

// 模拟 useOrderStore
jest.mock('@/stores/orderStore')

// 模拟 useNavigate
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: jest.fn(),
}))

// 模拟 antd
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  Table: jest.fn(({ columns, dataSource, rowKey, pagination, locale }) => (
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
  Space: jest.fn(({ children }) => (
    <div data-testid="space">
      {children}
    </div>
  )),
  Tag: jest.fn(({ color, children }) => (
    <span data-testid="tag" style={{ color }}>
      {children}
    </span>
  )),
  Empty: jest.fn(({ description }) => (
    <div data-testid="empty">
      {description}
    </div>
  )),
  Spin: jest.fn(({ spinning, children }) => (
    <div data-testid="spin" data-spinning={spinning}>
      {children}
    </div>
  )),
  Modal: jest.fn(({ title, open, onCancel, footer, children }) => (
    open ? (
      <div data-testid="modal" style={{ position: 'fixed', top: 0, left: 0, width: '100%', height: '100%', background: 'rgba(0,0,0,0.5)' }}>
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
          <button onClick={onCancel} data-testid="modal-cancel">关闭</button>
        </div>
      </div>
    ) : null
  )),
}))

const mockUseOrderStore = useOrderStore as jest.MockedFunction<typeof useOrderStore>
const mockUseNavigate = useNavigate as jest.MockedFunction<typeof useNavigate>

describe('OrderListPage Component', () => {
  const mockNavigate = jest.fn()
  
  beforeEach(() => {
    jest.clearAllMocks()
    mockUseNavigate.mockReturnValue(mockNavigate)
  })

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
    ]

    // 模拟 store 状态
    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 检查页面标题
    expect(screen.getByText('订单列表')).toBeInTheDocument()
  })

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
    ]

    // 模拟加载状态
    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: true,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 检查加载状态
    expect(screen.getByTestId('spin')).toBeInTheDocument()
  })

  test('shows error message when there is an error', async () => {
    // 模拟错误状态
    mockUseOrderStore.mockReturnValue({
      orders: [],
      isLoading: false,
      error: '加载失败',
      fetchOrders: jest.fn(),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 检查错误信息
    expect(screen.getByText('错误: 加载失败')).toBeInTheDocument()
  })

  test('shows empty state when no orders', async () => {
    // 模拟无订单状态
    mockUseOrderStore.mockReturnValue({
      orders: [],
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue([]),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 检查空状态
    expect(screen.getByText('暂无订单')).toBeInTheDocument()
  })

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
    ]

    // 模拟有订单状态
    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 检查订单列表
    expect(screen.getByTestId('order-table')).toBeInTheDocument()
    expect(screen.getByTestId('order-1')).toBeInTheDocument()
    expect(screen.getByTestId('order-2')).toBeInTheDocument()
  })

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
    ]

    // 模拟 store 状态
    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 点击详情按钮
    const detailButton = screen.getByText('详情')
    await act(async () => {
      fireEvent.click(detailButton)
    })

    // 检查模态框是否打开
    expect(screen.getByTestId('modal')).toBeInTheDocument()
    expect(screen.getByText('订单详情 #1')).toBeInTheDocument()
  })

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
    ]

    // 模拟 store 状态
    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 点击详情按钮打开模态框
    const detailButton = screen.getByText('详情')
    await act(async () => {
      fireEvent.click(detailButton)
    })

    // 点击关闭按钮
    const closeButton = screen.getByTestId('modal-cancel')
    await act(async () => {
      fireEvent.click(closeButton)
    })

    // 检查模态框是否关闭
    expect(screen.queryByTestId('modal')).not.toBeInTheDocument()
  })

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
    ]

    // 模拟 store 状态
    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 点击支付按钮
    const payButton = screen.getByText('支付')
    await act(async () => {
      fireEvent.click(payButton)
    })

    // 验证导航函数被调用
    expect(mockNavigate).toHaveBeenCalledWith('/payment/1')
  })

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
    ]

    // 模拟 store 状态
    mockUseOrderStore.mockReturnValue({
      orders: mockOrders,
      isLoading: false,
      error: null,
      fetchOrders: jest.fn().mockResolvedValue(mockOrders),
      createOrder: jest.fn(),
      cancelOrder: jest.fn(),
      clearError: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <OrderListPage />
        </MemoryRouter>
      )
    })

    // 检查支付按钮是否不存在
    expect(screen.queryByText('支付')).not.toBeInTheDocument()
  })
})
