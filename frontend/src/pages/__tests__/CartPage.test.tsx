import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { message } from 'antd'
import CartPage from '../CartPage'
import { useCartStore } from '@/stores/cartStore'

// 模拟useCartStore
jest.mock('@/stores/cartStore', () => ({
  useCartStore: jest.fn(),
}))

// 模拟message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
  },
}))

const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>
const mockMessageSuccess = message.success as jest.MockedFunction<typeof message.success>

describe('CartPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('购物车为空时显示空状态', () => {
    mockUseCartStore.mockReturnValue({
      items: [],
      total: 0,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
    })

    render(
      <BrowserRouter>
        <CartPage />
      </BrowserRouter>
    )

    expect(screen.getByText('购物车是空的')).toBeInTheDocument()
    expect(screen.getByText('继续购物')).toBeInTheDocument()
  })

  test('购物车有商品时显示商品列表', () => {
    const mockItems = [
      {
        id: 1,
        product_id: 1,
        quantity: 2,
        product: {
          id: 1,
          name: '测试商品1',
          price: 100,
        },
      },
      {
        id: 2,
        product_id: 2,
        quantity: 1,
        product: {
          id: 2,
          name: '测试商品2',
          price: 200,
        },
      },
    ]

    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 400,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
    })

    render(
      <BrowserRouter>
        <CartPage />
      </BrowserRouter>
    )

    expect(screen.getByText('购物车')).toBeInTheDocument()
    expect(screen.getByText('测试商品1')).toBeInTheDocument()
    expect(screen.getByText('测试商品2')).toBeInTheDocument()
    expect(screen.getByText('¥100.00')).toBeInTheDocument()
    const priceElements = screen.getAllByText('¥200.00')
    expect(priceElements.length).toBeGreaterThanOrEqual(2)
    expect(screen.getByText('总金额')).toBeInTheDocument()
    expect(screen.getByText('继续购物')).toBeInTheDocument()
    expect(screen.getByText('去结算')).toBeInTheDocument()
  })

  test('点击商品链接跳转到商品详情页', () => {
    const mockItems = [
      {
        id: 1,
        product_id: 1,
        quantity: 2,
        product: {
          id: 1,
          name: '测试商品1',
          price: 100,
        },
      },
    ]

    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
    })

    const { container } = render(
      <BrowserRouter>
        <CartPage />
      </BrowserRouter>
    )

    const productLink = screen.getByText('测试商品1')
    fireEvent.click(productLink)

    // 这里我们无法直接测试导航，但可以确保链接存在
    expect(productLink).toBeInTheDocument()
  })

  test('修改商品数量', () => {
    const mockUpdateQuantity = jest.fn()
    const mockItems = [
      {
        id: 1,
        product_id: 1,
        quantity: 2,
        product: {
          id: 1,
          name: '测试商品1',
          price: 100,
        },
      },
    ]

    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      removeItem: jest.fn(),
      updateQuantity: mockUpdateQuantity,
    })

    render(
      <BrowserRouter>
        <CartPage />
      </BrowserRouter>
    )

    // 找到数量输入框并修改
    const inputElements = screen.getAllByRole('spinbutton')
    expect(inputElements.length).toBe(1)
    
    fireEvent.change(inputElements[0], { target: { value: '3' } })
    expect(mockUpdateQuantity).toHaveBeenCalledWith(1, 3)
  })

  test('删除商品', () => {
    const mockRemoveItem = jest.fn()
    const mockItems = [
      {
        id: 1,
        product_id: 1,
        quantity: 2,
        product: {
          id: 1,
          name: '测试商品1',
          price: 100,
        },
      },
    ]

    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      removeItem: mockRemoveItem,
      updateQuantity: jest.fn(),
    })

    render(
      <BrowserRouter>
        <CartPage />
      </BrowserRouter>
    )

    const deleteButton = screen.getByText('删除')
    fireEvent.click(deleteButton)

    expect(mockRemoveItem).toHaveBeenCalledWith(1)
    expect(mockMessageSuccess).toHaveBeenCalledWith('已删除')
  })

  test('购物车有商品时显示继续购物按钮', () => {
    const mockItems = [
      {
        id: 1,
        product_id: 1,
        quantity: 2,
        product: {
          id: 1,
          name: '测试商品1',
          price: 100,
        },
      },
    ]

    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
    })

    render(
      <BrowserRouter>
        <CartPage />
      </BrowserRouter>
    )

    const continueShoppingButtons = screen.getAllByText('继续购物')
    expect(continueShoppingButtons.length).toBe(1)
  })

  test('购物车为空时显示继续购物按钮', () => {
    mockUseCartStore.mockReturnValue({
      items: [],
      total: 0,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
    })

    render(
      <BrowserRouter>
        <CartPage />
      </BrowserRouter>
    )

    const emptyCartContinueButton = screen.getByText('继续购物')
    expect(emptyCartContinueButton).toBeInTheDocument()
  })

  test('点击去结算按钮', () => {
    const mockItems = [
      {
        id: 1,
        product_id: 1,
        quantity: 2,
        product: {
          id: 1,
          name: '测试商品1',
          price: 100,
        },
      },
    ]

    mockUseCartStore.mockReturnValue({
      items: mockItems,
      total: 200,
      removeItem: jest.fn(),
      updateQuantity: jest.fn(),
    })

    render(
      <BrowserRouter>
        <CartPage />
      </BrowserRouter>
    )

    const checkoutButton = screen.getByText('去结算')
    expect(checkoutButton).toBeInTheDocument()
  })
})
