import { render, screen, fireEvent, act } from '@testing-library/react'
import { MemoryRouter, useNavigate } from 'react-router-dom'
import CartPage from '../CartPage'
import { useCartStore } from '@/stores/cartStore'

jest.mock('@/stores/cartStore')

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: jest.fn(),
}))

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  Table: jest.fn(({ columns, dataSource, rowKey, style }) => {
    const getValue = (obj: any, path: any) => {
      if (Array.isArray(path)) {
        return path.reduce((acc, key) => acc && acc[key], obj);
      }
      return obj[path];
    };

    return (
      <div data-testid="cart-table" style={style}>
        {dataSource.map((item: any) => (
          <div key={item[rowKey]} data-testid={`cart-item-${item.id}`}>
            {columns.map((col: any) => (
              <div key={col.key}>
                {col.render ? col.render(getValue(item, col.dataIndex), item) : getValue(item, col.dataIndex)}
              </div>
            ))}
          </div>
        ))}
      </div>
    );
  }),
  Button: jest.fn(({ type, danger, children, onClick, size }) => (
    <button 
      data-testid="button" 
      onClick={onClick}
      className={`${type === 'primary' ? 'primary' : ''} ${danger ? 'danger' : ''} ${size === 'large' ? 'large' : ''}`}
    >
      {children}
    </button>
  )),
  Space: jest.fn(({ children }) => (
    <div data-testid="space">
      {children}
    </div>
  )),
  Empty: jest.fn(({ description, children }) => (
    <div data-testid="empty">
      {description}
      {children}
    </div>
  )),
  InputNumber: jest.fn(({ min, value, onChange }) => (
    <input
      type="number"
      min={min}
      value={value}
      onChange={(e) => onChange(Number(e.target.value))}
      data-testid="input-number"
    />
  )),
  Row: jest.fn(({ children, justify }) => (
    <div data-testid="row" style={{ display: 'flex', justifyContent: justify }}>
      {children}
    </div>
  )),
  Col: jest.fn(({ span, children }) => (
    <div data-testid="col" style={{ flex: `0 0 ${(span / 24) * 100}%` }}>
      {children}
    </div>
  )),
  Card: jest.fn(({ children }) => (
    <div data-testid="card">
      {children}
    </div>
  )),
  Statistic: jest.fn(({ title, value, prefix }) => (
    <div data-testid="statistic">
      <div>{title}</div>
      <div>{prefix}{value}</div>
    </div>
  )),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>
const mockUseNavigate = useNavigate as jest.MockedFunction<typeof useNavigate>

describe('CartPage Component', () => {
  const mockNavigate = jest.fn()
  
  beforeEach(() => {
    jest.clearAllMocks()
    mockUseNavigate.mockReturnValue(mockNavigate)
  })

  test('renders CartPage with empty state', async () => {
    // 模拟空购物车状态
    mockUseCartStore.mockReturnValue({
      items: [],
      total: 0,
      addItem: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <CartPage />
        </MemoryRouter>
      )
    })

    // 检查空状态
    expect(screen.getByText('购物车是空的')).toBeInTheDocument()
    expect(screen.getByText('继续购物')).toBeInTheDocument()
  })

  test('renders CartPage with items', async () => {
    const mockItems = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        product: {
          id: 101,
          name: '测试产品1',
          price: 100,
        },
      },
      {
        id: 2,
        product_id: 102,
        quantity: 1,
        product: {
          id: 102,
          name: '测试产品2',
          price: 150,
        },
      },
    ]

    // 模拟有商品的购物车状态
    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 350,
      addItem: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <CartPage />
        </MemoryRouter>
      )
    })

    // 检查购物车页面
    expect(screen.getByText('购物车')).toBeInTheDocument()
    expect(screen.getByTestId('cart-table')).toBeInTheDocument()
    expect(screen.getByTestId('cart-item-1')).toBeInTheDocument()
    expect(screen.getByTestId('cart-item-2')).toBeInTheDocument()
    expect(screen.getByText('总金额')).toBeInTheDocument()
    expect(screen.getByText('¥350')).toBeInTheDocument()
  })

  test('handles remove item', async () => {
    const mockItems = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        product: {
          id: 101,
          name: '测试产品1',
          price: 100,
        },
      },
    ]

    const mockRemoveItem = jest.fn()

    // 模拟有商品的购物车状态
    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      addItem: jest.fn(),
      removeItem: mockRemoveItem,
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <CartPage />
        </MemoryRouter>
      )
    })

    // 点击删除按钮
    const deleteButton = screen.getByText('删除')
    await act(async () => {
      fireEvent.click(deleteButton)
    })

    // 验证删除函数被调用
    expect(mockRemoveItem).toHaveBeenCalledWith(1)
  })

  test('handles update quantity', async () => {
    const mockItems = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        product: {
          id: 101,
          name: '测试产品1',
          price: 100,
        },
      },
    ]

    const mockUpdateQuantity = jest.fn()

    // 模拟有商品的购物车状态
    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      addItem: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: mockUpdateQuantity,
      clearCart: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <CartPage />
        </MemoryRouter>
      )
    })

    // 更新数量
    const inputNumber = screen.getByTestId('input-number') as HTMLInputElement
    await act(async () => {
      fireEvent.change(inputNumber, { target: { value: '3' } })
    })

    // 验证更新数量函数被调用
    expect(mockUpdateQuantity).toHaveBeenCalledWith(1, 3)
  })

  test('navigates to products page', async () => {
    const mockItems = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        product: {
          id: 101,
          name: '测试产品1',
          price: 100,
        },
      },
    ]

    // 模拟有商品的购物车状态
    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      addItem: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <CartPage />
        </MemoryRouter>
      )
    })

    // 点击继续购物按钮
    const continueShoppingButton = screen.getAllByText('继续购物')[0]
    await act(async () => {
      fireEvent.click(continueShoppingButton)
    })

    // 验证导航函数被调用
    expect(mockNavigate).toHaveBeenCalledWith('/products')
  })

  test('navigates to checkout page', async () => {
    const mockItems = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        product: {
          id: 101,
          name: '测试产品1',
          price: 100,
        },
      },
    ]

    // 模拟有商品的购物车状态
    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      addItem: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <CartPage />
        </MemoryRouter>
      )
    })

    // 点击去结算按钮
    const checkoutButton = screen.getByText('去结算')
    await act(async () => {
      fireEvent.click(checkoutButton)
    })

    // 验证导航函数被调用
    expect(mockNavigate).toHaveBeenCalledWith('/checkout')
  })

  test('navigates to product detail page', async () => {
    const mockItems = [
      {
        id: 1,
        product_id: 101,
        quantity: 2,
        product: {
          id: 101,
          name: '测试产品1',
          price: 100,
        },
      },
    ]

    // 模拟有商品的购物车状态
    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      addItem: jest.fn(),
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
      clearCart: jest.fn(),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <CartPage />
        </MemoryRouter>
      )
    })

    // 点击产品名称
    const productName = screen.getByText('测试产品1')
    await act(async () => {
      fireEvent.click(productName)
    })

    // 验证导航函数被调用
    expect(mockNavigate).toHaveBeenCalledWith('/products/101')
  })
})
