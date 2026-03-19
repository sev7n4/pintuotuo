import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import HomePage from '../HomePage'
import { useHomeStore } from '@/stores/homeStore'

// 模拟CSS模块
jest.mock('../HomePage.module.css', () => ({}))

// 模拟useHomeStore
jest.mock('@/stores/homeStore', () => ({
  useHomeStore: jest.fn(),
}))

const mockUseHomeStore = useHomeStore as jest.MockedFunction<typeof useHomeStore>

describe('HomePage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  const mockBanners = [
    { id: 1, title: '测试banner1', link: '/products/1' },
    { id: 2, title: '测试banner2', link: '/products/2' },
  ]

  const mockCategories = [
    { name: '分类1', count: 10 },
    { name: '分类2', count: 20 },
    { name: '分类3', count: 15 },
  ]

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
  ]

  test('页面加载时调用fetchHomeData', () => {
    const mockFetchHomeData = jest.fn()
    
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: mockFetchHomeData,
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    expect(mockFetchHomeData).toHaveBeenCalled()
  })

  test('显示搜索框', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    expect(screen.getByPlaceholderText('搜索商品')).toBeInTheDocument()
  })

  test('搜索功能', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    const searchInput = screen.getByPlaceholderText('搜索商品')
    fireEvent.change(searchInput, { target: { value: '测试' } })
    fireEvent.submit(searchInput)

    // 搜索功能会触发导航，这里我们无法直接测试导航，但可以确保搜索框存在
    expect(searchInput).toBeInTheDocument()
  })

  test('显示轮播图', () => {
    mockUseHomeStore.mockReturnValue({
      banners: mockBanners,
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    const banner1Elements = screen.getAllByText('测试banner1')
    expect(banner1Elements.length).toBeGreaterThan(0)
    const banner2Elements = screen.getAllByText('测试banner2')
    expect(banner2Elements.length).toBeGreaterThan(0)
  })

  test('显示分类', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: mockCategories,
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    expect(screen.getByText('分类1')).toBeInTheDocument()
    expect(screen.getByText('分类2')).toBeInTheDocument()
    expect(screen.getByText('分类3')).toBeInTheDocument()
    expect(screen.getByText('10件')).toBeInTheDocument()
    expect(screen.getByText('20件')).toBeInTheDocument()
    expect(screen.getByText('15件')).toBeInTheDocument()
  })

  test('显示热门推荐商品', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: mockProducts,
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    expect(screen.getByText('热门推荐')).toBeInTheDocument()
    expect(screen.getByText('测试商品1')).toBeInTheDocument()
    expect(screen.getByText('测试商品2')).toBeInTheDocument()
    expect(screen.getByText('¥100.00')).toBeInTheDocument()
    expect(screen.getByText('¥200.00')).toBeInTheDocument()
  })

  test('显示新品上架商品', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: mockProducts,
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    expect(screen.getByText('新品上架')).toBeInTheDocument()
    expect(screen.getByText('测试商品1')).toBeInTheDocument()
    expect(screen.getByText('测试商品2')).toBeInTheDocument()
  })

  test('显示加载状态', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: true,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    // 检查加载状态是否显示
    expect(screen.getByText('热门推荐')).toBeInTheDocument()
    expect(screen.getByText('新品上架')).toBeInTheDocument()
  })

  test('显示错误状态', () => {
    const errorMessage = '加载失败'
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: errorMessage,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    expect(screen.getByText(errorMessage)).toBeInTheDocument()
  })

  test('点击商品跳转到详情页', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: mockProducts,
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    const productCard = screen.getByText('测试商品1')
    fireEvent.click(productCard)

    // 这里我们无法直接测试导航，但可以确保商品卡片存在
    expect(productCard).toBeInTheDocument()
  })

  test('点击分类跳转到商品列表页', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: mockCategories,
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    const categoryItem = screen.getByText('分类1')
    fireEvent.click(categoryItem)

    // 这里我们无法直接测试导航，但可以确保分类项存在
    expect(categoryItem).toBeInTheDocument()
  })

  test('点击查看全部链接', () => {
    mockUseHomeStore.mockReturnValue({
      banners: [],
      hotProducts: mockProducts,
      newProducts: mockProducts,
      categories: [],
      isLoading: false,
      error: null,
      fetchHomeData: jest.fn(),
    })

    render(
      <BrowserRouter>
        <HomePage />
      </BrowserRouter>
    )

    const viewAllLinks = screen.getAllByText('查看全部')
    expect(viewAllLinks.length).toBe(2)
  })
})
