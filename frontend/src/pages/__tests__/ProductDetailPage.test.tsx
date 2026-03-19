
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BrowserRouter, useParams } from 'react-router-dom'
import { message } from 'antd'
import ProductDetailPage from '../ProductDetailPage'
import { useProductStore } from '@/stores/productStore'
import { useCartStore } from '@/stores/cartStore'

// 模拟useParams
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useParams: jest.fn(),
}))

// 模拟useProductStore
jest.mock('@/stores/productStore', () => ({
  useProductStore: jest.fn(),
}))

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

const mockUseParams = useParams as jest.MockedFunction<typeof useParams>
const mockUseProductStore = useProductStore as jest.MockedFunction<typeof useProductStore>
const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>
const mockMessageSuccess = message.success as jest.MockedFunction<typeof message.success>

describe('ProductDetailPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // 模拟默认的useParams返回值
    mockUseParams.mockReturnValue({ id: '1' })
  })

  const mockProduct = {
    id: 1,
    name: '测试商品',
    description: '这是一个测试商品',
    price: 100,
    stock: 50,
    sold_count: 10,
    category: '测试分类',
    status: 'active',
  }

  test('显示加载状态', () => {
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn(),
      isLoading: true,
      error: null,
    })

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
    })

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    )

    // 检查加载状态是否显示
    expect(screen.queryByText('测试商品')).not.toBeInTheDocument() // 确保商品信息未显示
  })

  test('显示错误状态', () => {
    const errorMessage = '加载失败'
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn(),
      isLoading: false,
      error: errorMessage,
    })

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
    })

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    )

    expect(screen.getByText(`错误: ${errorMessage}`)).toBeInTheDocument()
  })

  test('显示产品详情', async () => {
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn().mockResolvedValue(mockProduct),
      isLoading: false,
      error: null,
    })

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
    })

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    )

    // 等待产品加载完成
    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument()
    })

    expect(screen.getByText('这是一个测试商品')).toBeInTheDocument()
    expect(screen.getByText('价格')).toBeInTheDocument()
    expect(screen.getByText('库存')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument() // 价格数字
    expect(screen.getByText('50')).toBeInTheDocument() // 库存数字
    expect(screen.getByText('购买数量:')).toBeInTheDocument()
    expect(screen.getByText('加入购物车')).toBeInTheDocument()
  })

  test('产品无库存时显示暂无库存', async () => {
    const outOfStockProduct = {
      ...mockProduct,
      stock: 0,
    }

    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn().mockResolvedValue(outOfStockProduct),
      isLoading: false,
      error: null,
    })

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
    })

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('暂无库存')).toBeInTheDocument()
    })

    // 找到包含"暂无库存"文本的按钮
    const addToCartButton = screen.getByRole('button', { name: /暂无库存/ })
    expect(addToCartButton).toBeDisabled()
  })

  test('调整购买数量', async () => {
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn().mockResolvedValue(mockProduct),
      isLoading: false,
      error: null,
    })

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
    })

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument()
    })

    // 找到数量输入框并修改
    const inputElements = screen.getAllByRole('spinbutton')
    expect(inputElements.length).toBe(1)
    
    fireEvent.change(inputElements[0], { target: { value: '5' } })
  })

  test('添加到购物车', async () => {
    const mockAddItem = jest.fn()
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn().mockResolvedValue(mockProduct),
      isLoading: false,
      error: null,
    })

    mockUseCartStore.mockReturnValue({
      addItem: mockAddItem,
    })

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument()
    })

    const addToCartButton = screen.getByText('加入购物车')
    fireEvent.click(addToCartButton)

    expect(mockAddItem).toHaveBeenCalledWith(mockProduct, 1)
    expect(mockMessageSuccess).toHaveBeenCalledWith('已添加 1 件到购物车')
  })

  test('返回按钮', async () => {
    mockUseProductStore.mockReturnValue({
      fetchProductByID: jest.fn().mockResolvedValue(mockProduct),
      isLoading: false,
      error: null,
    })

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
    })

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    )

    await waitFor(() => {
      expect(screen.getByText('测试商品')).toBeInTheDocument()
    })

    const backButton = screen.getByText('返回列表')
    expect(backButton).toBeInTheDocument()
  })

  test('无产品ID时不加载', () => {
    mockUseParams.mockReturnValue({ id: undefined })
    const mockFetchProductByID = jest.fn()

    mockUseProductStore.mockReturnValue({
      fetchProductByID: mockFetchProductByID,
      isLoading: false,
      error: null,
    })

    mockUseCartStore.mockReturnValue({
      addItem: jest.fn(),
    })

    render(
      <BrowserRouter>
        <ProductDetailPage />
      </BrowserRouter>
    )

    expect(mockFetchProductByID).not.toHaveBeenCalled()
  })
})
