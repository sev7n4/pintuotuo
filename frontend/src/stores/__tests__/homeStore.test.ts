import { useHomeStore } from '../homeStore';
import { productService } from '@/services/product';
import { Product, Category, Banner } from '@/types';

jest.mock('@/services/product');

const mockedProductService = productService as jest.Mocked<typeof productService>;

const mockProduct: Product = {
  id: 1,
  merchant_id: 1,
  name: 'Test Product',
  description: 'Test Description',
  price: 99.99,
  stock: 100,
  status: 'active',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

const mockCategory: Category = {
  name: 'Test Category',
  count: 10,
};

const mockBanner: Banner = {
  id: 1,
  title: 'Test Banner',
  image: 'https://example.com/banner.jpg',
  link: '/product/1',
};

describe('HomeStore', () => {
  beforeEach(() => {
    useHomeStore.setState({
      banners: [],
      hotProducts: [],
      newProducts: [],
      categories: [],
      isLoading: false,
      error: null,
    });
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('should have correct initial values', () => {
      const state = useHomeStore.getState();
      expect(state.banners).toEqual([]);
      expect(state.hotProducts).toEqual([]);
      expect(state.newProducts).toEqual([]);
      expect(state.categories).toEqual([]);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });
  });

  describe('fetchHomeData', () => {
    it('should fetch home data successfully', async () => {
      const mockResponse = {
        data: {
          banners: [mockBanner],
          hot: [mockProduct],
          new: [mockProduct],
          categories: [mockCategory],
        },
      };

      mockedProductService.getHomeData.mockResolvedValueOnce(mockResponse as any);

      const store = useHomeStore.getState();
      await store.fetchHomeData();

      const newState = useHomeStore.getState();
      expect(newState.banners).toEqual([mockBanner]);
      expect(newState.hotProducts).toEqual([mockProduct]);
      expect(newState.newProducts).toEqual([mockProduct]);
      expect(newState.categories).toEqual([mockCategory]);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle fetch home data error', async () => {
      const errorMessage = 'Failed to fetch home data';
      mockedProductService.getHomeData.mockRejectedValueOnce(new Error(errorMessage));

      const store = useHomeStore.getState();
      await store.fetchHomeData();

      const newState = useHomeStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('fetchHotProducts', () => {
    it('should fetch hot products successfully', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: [mockProduct],
        },
      };

      mockedProductService.getHotProducts.mockResolvedValueOnce(mockResponse as any);

      const store = useHomeStore.getState();
      await store.fetchHotProducts();

      const newState = useHomeStore.getState();
      expect(newState.hotProducts).toEqual([mockProduct]);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle fetch hot products error', async () => {
      const errorMessage = 'Failed to fetch hot products';
      mockedProductService.getHotProducts.mockRejectedValueOnce(new Error(errorMessage));

      const store = useHomeStore.getState();
      await store.fetchHotProducts();

      const newState = useHomeStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('fetchNewProducts', () => {
    it('should fetch new products successfully', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: [mockProduct],
        },
      };

      mockedProductService.getNewProducts.mockResolvedValueOnce(mockResponse as any);

      const store = useHomeStore.getState();
      await store.fetchNewProducts();

      const newState = useHomeStore.getState();
      expect(newState.newProducts).toEqual([mockProduct]);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle fetch new products error', async () => {
      const errorMessage = 'Failed to fetch new products';
      mockedProductService.getNewProducts.mockRejectedValueOnce(new Error(errorMessage));

      const store = useHomeStore.getState();
      await store.fetchNewProducts();

      const newState = useHomeStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('fetchCategories', () => {
    it('should fetch categories successfully', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: [mockCategory],
        },
      };

      mockedProductService.getCategories.mockResolvedValueOnce(mockResponse as any);

      const store = useHomeStore.getState();
      await store.fetchCategories();

      const newState = useHomeStore.getState();
      expect(newState.categories).toEqual([mockCategory]);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle fetch categories error', async () => {
      const errorMessage = 'Failed to fetch categories';
      mockedProductService.getCategories.mockRejectedValueOnce(new Error(errorMessage));

      const store = useHomeStore.getState();
      await store.fetchCategories();

      const newState = useHomeStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('clearError', () => {
    it('should clear error', () => {
      useHomeStore.setState({ error: 'Test error' });

      const store = useHomeStore.getState();
      store.clearError();

      // 重新获取 store 状态来验证错误是否被清除
      const updatedStore = useHomeStore.getState();
      expect(updatedStore.error).toBeNull();
    });
  });
});
