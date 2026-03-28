import api from '../api';
import { productService } from '../product';

jest.mock('../api');

const mockedApi = api as jest.Mocked<typeof api>;

const createMockResponse = <T>(data: T) => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: { headers: {} },
});

describe('ProductService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('listProducts', () => {
    it('should call GET /products with filters', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
        data: {
          total: 2,
          page: 1,
          per_page: 20,
          data: [
            { id: 1, name: 'Product 1', price: 99.99 },
            { id: 2, name: 'Product 2', price: 49.99 },
          ],
        },
      });

      mockedApi.get.mockResolvedValueOnce(mockResponse as any);

      const result = await productService.listProducts({ page: 1, per_page: 20 });

      expect(mockedApi.get).toHaveBeenCalledWith('/products', {
        params: { page: 1, per_page: 20 },
      });
      expect(result.data.data?.total).toBe(2);
    });

    it('should call GET /products without filters', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
        data: { total: 0, page: 1, per_page: 20, data: [] },
      });

      mockedApi.get.mockResolvedValueOnce(mockResponse as any);

      await productService.listProducts();

      expect(mockedApi.get).toHaveBeenCalledWith('/products', { params: undefined });
    });
  });

  describe('getProductByID', () => {
    it('should call GET /products/:id', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
        data: {
          id: 1,
          merchant_id: 1,
          name: 'Test Product',
          description: 'Test Description',
          price: 99.99,
          stock: 100,
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      });

      mockedApi.get.mockResolvedValueOnce(mockResponse as any);

      const result = await productService.getProductByID(1);

      expect(mockedApi.get).toHaveBeenCalledWith('/products/1');
      expect(result.data.data?.name).toBe('Test Product');
    });
  });

  describe('searchProducts', () => {
    it('should call GET /products/search with query', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
        data: {
          total: 1,
          page: 1,
          per_page: 20,
          data: [{ id: 1, name: 'Test Product' }],
        },
      });

      mockedApi.get.mockResolvedValueOnce(mockResponse as any);

      const result = await productService.searchProducts('test');

      expect(mockedApi.get).toHaveBeenCalledWith('/products/search', {
        params: { q: 'test' },
      });
      expect(result.data.data?.total).toBe(1);
    });
  });

  describe('createProduct', () => {
    it('should call POST /products/merchants', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
        data: {
          id: 1,
          merchant_id: 1,
          name: 'New Product',
          description: 'Description',
          price: 99.99,
          stock: 100,
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      });

      mockedApi.post.mockResolvedValueOnce(mockResponse as any);

      const result = await productService.createProduct({
        name: 'New Product',
        description: 'Description',
        price: 99.99,
        stock: 100,
      });

      expect(mockedApi.post).toHaveBeenCalledWith('/products/merchants', {
        name: 'New Product',
        description: 'Description',
        price: 99.99,
        stock: 100,
      });
      expect(result.data.data?.name).toBe('New Product');
    });
  });

  describe('updateProduct', () => {
    it('should call PUT /products/merchants/:id', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
        data: {
          id: 1,
          merchant_id: 1,
          name: 'Updated Product',
          description: 'Description',
          price: 79.99,
          stock: 50,
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-02T00:00:00Z',
        },
      });

      mockedApi.put.mockResolvedValueOnce(mockResponse as any);

      const result = await productService.updateProduct(1, { price: 79.99 });

      expect(mockedApi.put).toHaveBeenCalledWith('/products/merchants/1', { price: 79.99 });
      expect(result.data.data?.price).toBe(79.99);
    });
  });

  describe('deleteProduct', () => {
    it('should call DELETE /products/merchants/:id', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
      });

      mockedApi.delete.mockResolvedValueOnce(mockResponse as any);

      await productService.deleteProduct(1);

      expect(mockedApi.delete).toHaveBeenCalledWith('/products/merchants/1');
    });
  });
});
