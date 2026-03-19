import { render, screen, fireEvent, act, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route, Navigate } from 'react-router-dom'
import { RegisterPage } from '../pages/RegisterPage'
import ProductListPage from '../pages/ProductListPage'
import ProductDetailPage from '../pages/ProductDetailPage'
import OrderListPage from '../pages/OrderListPage'
import MerchantDashboard from '../pages/merchant/MerchantDashboard'
import CartPage from '../pages/CartPage'
import CheckoutPage from '../pages/CheckoutPage'
import PaymentPage from '../pages/PaymentPage'
import Layout from '../components/Layout'
import { useAuthStore } from '@/stores/authStore'
import { useProductStore } from '@/stores/productStore'
import { useOrderStore } from '@/stores/orderStore'
import { useCartStore } from '@/stores/cartStore'
import { useMerchantStore } from '@/stores/merchantStore'

jest.mock('@/stores/authStore')
jest.mock('@/stores/productStore')
jest.mock('@/stores/orderStore')
jest.mock('@/stores/cartStore')
jest.mock('@/stores/merchantStore')

jest.mock('../pages/merchant/MerchantDashboard.module.css', () => ({}))
jest.mock('../components/Layout.css', () => ({}))

jest.mock('antd', () => {
  const antd = jest.requireActual('antd');
  return {
    ...antd,
    message: {
      success: jest.fn(),
      error: jest.fn(),
    },
    Table: antd.Table,
    Tabs: antd.Tabs,
    TabPane: antd.Tabs.TabPane,
    Input: antd.Input,
    Button: antd.Button,
    Space: antd.Space,
    Spin: antd.Spin,
    Card: antd.Card,
    Row: antd.Row,
    Col: antd.Col,
    Typography: antd.Typography,
    Statistic: antd.Statistic,
    Form: antd.Form,
    Radio: antd.Radio,
    Tag: antd.Tag,
    Pagination: antd.Pagination,
    Modal: antd.Modal,
    Descriptions: antd.Descriptions,
    Divider: antd.Divider,
    InputNumber: antd.InputNumber,
    Empty: antd.Empty,
    Menu: antd.Menu,
    Dropdown: antd.Dropdown,
    Avatar: antd.Avatar,
    Layout: antd.Layout,
  };
});

const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>
const mockUseProductStore = useProductStore as jest.MockedFunction<typeof useProductStore>
const mockUseOrderStore = useOrderStore as jest.MockedFunction<typeof useOrderStore>
const mockUseCartStore = useCartStore as jest.MockedFunction<typeof useCartStore>
const mockUseMerchantStore = useMerchantStore as jest.MockedFunction<typeof useMerchantStore>

describe('Page Navigation Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('User Registration Navigation', () => {
    test('should navigate to products page after C-end user registration', async () => {
      const mockRegister = jest.fn().mockResolvedValue({ token: 'user-token' })

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: mockRegister,
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

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
      })

      render(
        <MemoryRouter initialEntries={['/register']}>
          <Routes>
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/products" element={<ProductListPage />} />
            <Route path="/merchant/dashboard" element={<MerchantDashboard />} />
            <Route path="/*" element={<Navigate to="/register" />} />
          </Routes>
        </MemoryRouter>
      )

      const emailInput = screen.getByPlaceholderText('example@email.com')
      const nameInput = screen.getByPlaceholderText('输入你的名字')
      const passwordInput = screen.getByPlaceholderText('设置密码')
      const confirmPasswordInput = screen.getByPlaceholderText('再次输入密码')
      const registerButton = screen.getByText('创建账户')

      const userRadio = screen.getByText('普通用户').closest('label')?.querySelector('input')

      await act(async () => {
        if (userRadio) {
          fireEvent.click(userRadio)
        }
        fireEvent.change(emailInput, { target: { value: 'cuser@example.com' } })
        fireEvent.change(nameInput, { target: { value: 'C端用户' } })
        fireEvent.change(passwordInput, { target: { value: 'password123' } })
        fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } })
        fireEvent.click(registerButton)
      })

      await waitFor(() => {
        expect(mockRegister).toHaveBeenCalledWith('cuser@example.com', 'C端用户', 'password123', 'user')
      })
    })

    test('should navigate to merchant dashboard after B-end user registration', async () => {
      const mockRegister = jest.fn().mockResolvedValue({ token: 'merchant-token' })

      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: mockRegister,
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

      mockUseMerchantStore.mockReturnValue({
        stats: { total_products: 0, active_products: 0, month_sales: 0, month_orders: 0 },
        orders: [],
        products: [],
        isLoading: false,
        error: null,
        fetchStats: jest.fn(),
        fetchOrders: jest.fn(),
        fetchProducts: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/register']}>
          <Routes>
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/products" element={<ProductListPage />} />
            <Route path="/merchant/dashboard" element={<MerchantDashboard />} />
            <Route path="/*" element={<Navigate to="/register" />} />
          </Routes>
        </MemoryRouter>
      )

      const emailInput = screen.getByPlaceholderText('example@email.com')
      const nameInput = screen.getByPlaceholderText('输入你的名字')
      const passwordInput = screen.getByPlaceholderText('设置密码')
      const confirmPasswordInput = screen.getByPlaceholderText('再次输入密码')
      const registerButton = screen.getByText('创建账户')

      const merchantRadio = screen.getByText('商家').closest('label')?.querySelector('input')

      await act(async () => {
        if (merchantRadio) {
          fireEvent.click(merchantRadio)
        }
        fireEvent.change(emailInput, { target: { value: 'merchant@example.com' } })
        fireEvent.change(nameInput, { target: { value: 'B端商家' } })
        fireEvent.change(passwordInput, { target: { value: 'password123' } })
        fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } })
        fireEvent.click(registerButton)
      })

      await waitFor(() => {
        expect(mockRegister).toHaveBeenCalledWith('merchant@example.com', 'B端商家', 'password123', 'merchant')
      })
    })

    test('should show role selection UI correctly', async () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/register']}>
          <Routes>
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/*" element={<Navigate to="/register" />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('普通用户')).toBeInTheDocument()
      expect(screen.getByText('商家')).toBeInTheDocument()
      expect(screen.getByText('购买 Token、参与拼团')).toBeInTheDocument()
      expect(screen.getByText('上架商品、管理订单')).toBeInTheDocument()
    })
  })

  describe('Product Navigation', () => {
    test('should navigate from product list to product detail', async () => {
      const mockFetchProducts = jest.fn().mockResolvedValue(undefined)
      const mockFetchProductByID = jest.fn().mockResolvedValue({
        id: 1,
        name: '测试产品',
        description: '这是一个测试产品',
        price: 99.99,
        stock: 100,
        status: 'active',
      })

      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseProductStore.mockReturnValue({
        products: [
          { id: 1, name: '测试产品', description: '描述', price: 99.99, stock: 100, status: 'active' },
          { id: 2, name: '另一个产品', description: '描述2', price: 199.99, stock: 50, status: 'active' },
        ],
        total: 2,
        filters: { page: 1, per_page: 20 },
        isLoading: false,
        error: null,
        fetchProducts: mockFetchProducts,
        setFilters: jest.fn(),
        searchProducts: jest.fn(),
        fetchProductByID: mockFetchProductByID,
      })

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
      })

      render(
        <MemoryRouter initialEntries={['/products']}>
          <Routes>
            <Route path="/products" element={<ProductListPage />} />
            <Route path="/products/:id" element={<ProductDetailPage />} />
            <Route path="/*" element={<Navigate to="/products" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('测试产品')).toBeInTheDocument()
      })

      const detailButton = screen.getAllByText('详情')[0]
      expect(detailButton).toBeInTheDocument()
    })

    test('should display product list with correct columns', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseProductStore.mockReturnValue({
        products: [
          { id: 1, name: '产品A', description: '描述A', price: 50, stock: 10, status: 'active' },
        ],
        total: 1,
        filters: { page: 1, per_page: 20 },
        isLoading: false,
        error: null,
        fetchProducts: jest.fn(),
        setFilters: jest.fn(),
        searchProducts: jest.fn(),
        fetchProductByID: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/products']}>
          <Routes>
            <Route path="/products" element={<ProductListPage />} />
            <Route path="/*" element={<Navigate to="/products" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('产品A')).toBeInTheDocument()
        expect(screen.getByText('¥50.00')).toBeInTheDocument()
        expect(screen.getByText('上架')).toBeInTheDocument()
      })
    })

    test('should navigate to cart after adding product to cart', async () => {
      const mockAddItem = jest.fn()

      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseProductStore.mockReturnValue({
        products: [],
        total: 0,
        filters: { page: 1, per_page: 20 },
        isLoading: false,
        error: null,
        fetchProducts: jest.fn(),
        setFilters: jest.fn(),
        searchProducts: jest.fn(),
        fetchProductByID: jest.fn().mockResolvedValue({
          id: 1,
          name: '加入购物车测试产品',
          description: '测试描述',
          price: 88,
          stock: 20,
          status: 'active',
        }),
      })

      mockUseCartStore.mockReturnValue({
        items: [],
        total: 0,
        isLoading: false,
        error: null,
        addItem: mockAddItem,
        removeItem: jest.fn(),
        updateQuantity: jest.fn(),
        clear: jest.fn(),
        getTotal: jest.fn().mockReturnValue(0),
      })

      render(
        <MemoryRouter initialEntries={['/products/1']}>
          <Routes>
            <Route path="/products/:id" element={<ProductDetailPage />} />
            <Route path="/cart" element={<CartPage />} />
            <Route path="/*" element={<Navigate to="/products/1" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('加入购物车测试产品')).toBeInTheDocument()
      })

      const addToCartButton = screen.getByText('加入购物车')
      expect(addToCartButton).toBeInTheDocument()
    })
  })

  describe('Order Navigation', () => {
    test('should display order list with pending payment status', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseOrderStore.mockReturnValue({
        orders: [
          {
            id: 1,
            product_id: 1,
            quantity: 2,
            total_price: 199.98,
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
          },
          {
            id: 2,
            product_id: 2,
            quantity: 1,
            total_price: 99.99,
            status: 'paid',
            created_at: '2024-01-02T00:00:00Z',
          },
        ],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/orders']}>
          <Routes>
            <Route path="/orders" element={<OrderListPage />} />
            <Route path="/*" element={<Navigate to="/orders" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('订单列表')).toBeInTheDocument()
        expect(screen.getByText('待支付')).toBeInTheDocument()
        expect(screen.getByText('已支付')).toBeInTheDocument()
      })
    })

    test('should show payment button for pending orders', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseOrderStore.mockReturnValue({
        orders: [
          {
            id: 1,
            product_id: 1,
            quantity: 1,
            total_price: 100,
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
          },
        ],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/orders']}>
          <Routes>
            <Route path="/orders" element={<OrderListPage />} />
            <Route path="/payment/:id" element={<div>Payment Page</div>} />
            <Route path="/*" element={<Navigate to="/orders" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('支付')).toBeInTheDocument()
      })
    })

    test('should show order details in modal', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseOrderStore.mockReturnValue({
        orders: [
          {
            id: 1,
            product_id: 1,
            quantity: 3,
            total_price: 299.97,
            status: 'pending',
            created_at: '2024-01-15T10:30:00Z',
            group_id: 100,
          },
        ],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/orders']}>
          <Routes>
            <Route path="/orders" element={<OrderListPage />} />
            <Route path="/*" element={<Navigate to="/orders" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('详情')).toBeInTheDocument()
      })

      const detailButton = screen.getByText('详情')
      await act(async () => {
        fireEvent.click(detailButton)
      })

      await waitFor(() => {
        expect(screen.getByText('订单详情 #1')).toBeInTheDocument()
      })
    })

    test('should display completed orders correctly', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseOrderStore.mockReturnValue({
        orders: [
          {
            id: 1,
            product_id: 1,
            quantity: 1,
            total_price: 50,
            status: 'completed',
            created_at: '2024-01-01T00:00:00Z',
          },
        ],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/orders']}>
          <Routes>
            <Route path="/orders" element={<OrderListPage />} />
            <Route path="/*" element={<Navigate to="/orders" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('已完成')).toBeInTheDocument()
      })
    })
  })

  describe('Merchant Dashboard Navigation', () => {
    test('should display merchant dashboard after B-end registration', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'merchant@example.com', role: 'merchant' },
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
      })

      mockUseMerchantStore.mockReturnValue({
        stats: {
          total_products: 10,
          active_products: 8,
          month_sales: 5000.00,
          month_orders: 25,
        },
        orders: [],
        products: [],
        isLoading: false,
        error: null,
        fetchStats: jest.fn(),
        fetchOrders: jest.fn(),
        fetchProducts: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/merchant/dashboard']}>
          <Routes>
            <Route path="/merchant/dashboard" element={<MerchantDashboard />} />
            <Route path="/*" element={<Navigate to="/merchant/dashboard" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('数据概览')).toBeInTheDocument()
        expect(screen.getByText('商品总数')).toBeInTheDocument()
        expect(screen.getByText('在售商品')).toBeInTheDocument()
        expect(screen.getByText('本月销售额')).toBeInTheDocument()
        expect(screen.getByText('本月订单')).toBeInTheDocument()
      })
    })

    test('should display merchant statistics correctly', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'merchant@example.com', role: 'merchant' },
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
      })

      mockUseMerchantStore.mockReturnValue({
        stats: {
          total_products: 50,
          active_products: 45,
          month_sales: 15000.00,
          month_orders: 100,
        },
        orders: [],
        products: [],
        isLoading: false,
        error: null,
        fetchStats: jest.fn(),
        fetchOrders: jest.fn(),
        fetchProducts: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/merchant/dashboard']}>
          <Routes>
            <Route path="/merchant/dashboard" element={<MerchantDashboard />} />
            <Route path="/*" element={<Navigate to="/merchant/dashboard" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('数据概览')).toBeInTheDocument()
      })
    })

    test('should display recent orders on merchant dashboard', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'merchant@example.com', role: 'merchant' },
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
      })

      mockUseMerchantStore.mockReturnValue({
        stats: {
          total_products: 10,
          active_products: 8,
          month_sales: 5000.00,
          month_orders: 25,
        },
        orders: [
          {
            id: 1,
            product_name: '测试商品',
            quantity: 2,
            total_price: 199.98,
            status: 'paid',
            created_at: '2024-01-15T10:30:00Z',
          },
        ],
        products: [],
        isLoading: false,
        error: null,
        fetchStats: jest.fn(),
        fetchOrders: jest.fn(),
        fetchProducts: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/merchant/dashboard']}>
          <Routes>
            <Route path="/merchant/dashboard" element={<MerchantDashboard />} />
            <Route path="/*" element={<Navigate to="/merchant/dashboard" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('最近订单')).toBeInTheDocument()
        expect(screen.getByText('测试商品')).toBeInTheDocument()
      })
    })
  })

  describe('Cart Navigation', () => {
    test('should display empty cart message when no items', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

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
      })

      render(
        <MemoryRouter initialEntries={['/cart']}>
          <Routes>
            <Route path="/cart" element={<CartPage />} />
            <Route path="/*" element={<Navigate to="/cart" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/购物车/i)).toBeInTheDocument()
      })
    })

    test('should display cart items with correct information', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseCartStore.mockReturnValue({
        items: [
          {
            id: '1-0-123',
            product_id: 1,
            product: {
              id: 1,
              name: '购物车商品A',
              price: 100,
              image: 'a.jpg',
            },
            quantity: 2,
            group_id: null,
          },
        ],
        total: 200,
        isLoading: false,
        error: null,
        addItem: jest.fn(),
        removeItem: jest.fn(),
        updateQuantity: jest.fn(),
        clear: jest.fn(),
        getTotal: jest.fn().mockReturnValue(200),
      })

      render(
        <MemoryRouter initialEntries={['/cart']}>
          <Routes>
            <Route path="/cart" element={<CartPage />} />
            <Route path="/*" element={<Navigate to="/cart" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('购物车商品A')).toBeInTheDocument()
      })
    })
  })

  describe('Cross-Page Navigation Flow', () => {
    test('should navigate from product list to cart flow', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseProductStore.mockReturnValue({
        products: [
          { id: 1, name: '跨页测试产品', description: '描述', price: 88, stock: 10, status: 'active' },
        ],
        total: 1,
        filters: { page: 1, per_page: 20 },
        isLoading: false,
        error: null,
        fetchProducts: jest.fn(),
        setFilters: jest.fn(),
        searchProducts: jest.fn(),
        fetchProductByID: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/products']}>
          <Routes>
            <Route path="/products" element={<ProductListPage />} />
            <Route path="/products/:id" element={<ProductDetailPage />} />
            <Route path="/cart" element={<CartPage />} />
            <Route path="/*" element={<Navigate to="/products" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('跨页测试产品')).toBeInTheDocument()
      })

      const addToCartButton = screen.getAllByText('加购')[0]
      expect(addToCartButton).toBeInTheDocument()
    })

    test('should navigate from order list to payment flow', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseOrderStore.mockReturnValue({
        orders: [
          {
            id: 1,
            product_id: 1,
            quantity: 1,
            total_price: 150,
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
          },
        ],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/orders']}>
          <Routes>
            <Route path="/orders" element={<OrderListPage />} />
            <Route path="/payment/:id" element={<div>Payment Page</div>} />
            <Route path="/*" element={<Navigate to="/orders" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('订单列表')).toBeInTheDocument()
        expect(screen.getByText('支付')).toBeInTheDocument()
      })
    })

    test('should handle complete user journey from registration to order', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'journey@example.com', role: 'user' },
        token: 'journey-token',
        isLoading: false,
        error: null,
        isAuthenticated: true,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

      mockUseProductStore.mockReturnValue({
        products: [
          { id: 1, name: '旅程产品', description: '完整流程测试', price: 199, stock: 50, status: 'active' },
        ],
        total: 1,
        filters: { page: 1, per_page: 20 },
        isLoading: false,
        error: null,
        fetchProducts: jest.fn(),
        setFilters: jest.fn(),
        searchProducts: jest.fn(),
        fetchProductByID: jest.fn(),
      })

      mockUseCartStore.mockReturnValue({
        items: [
          {
            id: '1-0-456',
            product_id: 1,
            product: {
              id: 1,
              name: '旅程产品',
              price: 199,
              image: 'journey.jpg',
            },
            quantity: 1,
            group_id: null,
          },
        ],
        total: 199,
        isLoading: false,
        error: null,
        addItem: jest.fn(),
        removeItem: jest.fn(),
        updateQuantity: jest.fn(),
        clear: jest.fn(),
        getTotal: jest.fn().mockReturnValue(199),
      })

      mockUseOrderStore.mockReturnValue({
        orders: [
          {
            id: 1,
            product_id: 1,
            quantity: 1,
            total_price: 199,
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
          },
        ],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/products']}>
          <Routes>
            <Route path="/products" element={<ProductListPage />} />
            <Route path="/cart" element={<CartPage />} />
            <Route path="/orders" element={<OrderListPage />} />
            <Route path="/*" element={<Navigate to="/products" />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('旅程产品')).toBeInTheDocument()
      })
    })
  })

  describe('Layout Authentication State', () => {
    test('should show login and register links when user is not authenticated', async () => {
      mockUseAuthStore.mockReturnValue({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        login: jest.fn(),
        register: jest.fn(),
        logout: jest.fn(),
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

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
      })

      render(
        <MemoryRouter initialEntries={['/products']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="products" element={<ProductListPage />} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('登录')).toBeInTheDocument()
      expect(screen.getByText('注册')).toBeInTheDocument()
      expect(screen.queryByTestId('user-dropdown')).not.toBeInTheDocument()
    })

    test('should show user info instead of login/register when authenticated', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'test@example.com', name: '测试用户', role: 'user' },
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
      })

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
      })

      render(
        <MemoryRouter initialEntries={['/products']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="products" element={<ProductListPage />} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('测试用户')).toBeInTheDocument()
      expect(screen.queryByText('登录')).not.toBeInTheDocument()
      expect(screen.queryByText('注册')).not.toBeInTheDocument()
    })

    test('should show user email when name is not available', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'no-name@example.com', role: 'user' },
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
      })

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
      })

      render(
        <MemoryRouter initialEntries={['/products']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="products" element={<ProductListPage />} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('no-name@example.com')).toBeInTheDocument()
      expect(screen.queryByText('登录')).not.toBeInTheDocument()
      expect(screen.queryByText('注册')).not.toBeInTheDocument()
    })

    test('should navigate to login page after logout', async () => {
      const mockLogout = jest.fn().mockResolvedValue(undefined)

      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'logout-test@example.com', name: '退出测试用户', role: 'user' },
        token: 'test-token',
        isLoading: false,
        error: null,
        isAuthenticated: true,
        login: jest.fn(),
        register: jest.fn(),
        logout: mockLogout,
        fetchUser: jest.fn(),
        setUser: jest.fn(),
        clearError: jest.fn(),
      })

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
      })

      render(
        <MemoryRouter initialEntries={['/products']}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route path="products" element={<ProductListPage />} />
              <Route path="login" element={<div>登录页面</div>} />
            </Route>
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('退出测试用户')).toBeInTheDocument()
    })
  })

  describe('Checkout and Payment Flow', () => {
    test('should display checkout page with cart items', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseCartStore.mockReturnValue({
        items: [
          {
            id: '1-0-789',
            product_id: 1,
            product: {
              id: 1,
              name: '结算商品',
              price: 150,
              stock: 10,
              status: 'active',
            },
            quantity: 2,
            group_id: null,
          },
        ],
        total: 300,
        isLoading: false,
        error: null,
        addItem: jest.fn(),
        removeItem: jest.fn(),
        updateQuantity: jest.fn(),
        clear: jest.fn(),
        getTotal: jest.fn().mockReturnValue(300),
      })

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn().mockResolvedValue(undefined),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/checkout']}>
          <Routes>
            <Route path="/checkout" element={<CheckoutPage />} />
            <Route path="/cart" element={<div>购物车页面</div>} />
            <Route path="/orders" element={<div>订单页面</div>} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('确认订单')).toBeInTheDocument()
      expect(screen.getByText('结算商品')).toBeInTheDocument()
      expect(screen.getByText('订单结算')).toBeInTheDocument()
    })

    test('should show empty state when cart is empty on checkout', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

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
      })

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/checkout']}>
          <Routes>
            <Route path="/checkout" element={<CheckoutPage />} />
            <Route path="/products" element={<div>产品列表</div>} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText('购物车是空的')).toBeInTheDocument()
    })

    test('should display payment page with order info', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: {
          id: 100,
          user_id: 1,
          product_id: 1,
          group_id: null,
          quantity: 2,
          total_price: 299,
          status: 'pending',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn().mockResolvedValue(undefined),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/100']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
            <Route path="/orders" element={<div>订单列表</div>} />
            <Route path="/products" element={<div>产品列表</div>} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('订单支付')).toBeInTheDocument()
      })
      expect(screen.getByText('#100')).toBeInTheDocument()
      expect(screen.getByText('选择支付方式')).toBeInTheDocument()
    })

    test('should show already paid message for completed order', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: {
          id: 101,
          user_id: 1,
          product_id: 1,
          group_id: null,
          quantity: 1,
          total_price: 199,
          status: 'paid',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn().mockResolvedValue(undefined),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/payment/101']}>
          <Routes>
            <Route path="/payment/:id" element={<PaymentPage />} />
            <Route path="/orders" element={<div>订单列表</div>} />
            <Route path="/products" element={<div>产品列表</div>} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('订单已支付')).toBeInTheDocument()
      })
    })

    test('should navigate from cart to checkout', async () => {
      mockUseAuthStore.mockReturnValue({
        user: { id: 1, email: 'user@example.com', role: 'user' },
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
      })

      mockUseCartStore.mockReturnValue({
        items: [
          {
            id: '1-0-999',
            product_id: 1,
            product: {
              id: 1,
              name: '导航测试商品',
              price: 88,
              stock: 10,
              status: 'active',
            },
            quantity: 1,
            group_id: null,
          },
        ],
        total: 88,
        isLoading: false,
        error: null,
        addItem: jest.fn(),
        removeItem: jest.fn(),
        updateQuantity: jest.fn(),
        clear: jest.fn(),
        getTotal: jest.fn().mockReturnValue(88),
      })

      mockUseOrderStore.mockReturnValue({
        orders: [],
        currentOrder: null,
        isLoading: false,
        error: null,
        fetchOrders: jest.fn(),
        fetchOrderByID: jest.fn(),
        createOrder: jest.fn(),
        cancelOrder: jest.fn(),
        clearError: jest.fn(),
      })

      render(
        <MemoryRouter initialEntries={['/cart']}>
          <Routes>
            <Route path="/cart" element={<CartPage />} />
            <Route path="/checkout" element={<CheckoutPage />} />
            <Route path="/products" element={<ProductListPage />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('导航测试商品')).toBeInTheDocument()
      })

      const checkoutButton = screen.getByText('去结算')
      expect(checkoutButton).toBeInTheDocument()
    })
  })
})
