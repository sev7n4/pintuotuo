import { useCartStore } from '../cartStore';
import { Product } from '@/types';

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

const mockProduct2: Product = {
  id: 2,
  merchant_id: 1,
  name: 'Test Product 2',
  description: 'Test Description 2',
  price: 49.99,
  stock: 50,
  status: 'active',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('CartStore', () => {
  beforeEach(() => {
    useCartStore.setState({ items: [], total: 0 });
  });

  describe('addItem', () => {
    it('should add a new item to the cart', () => {
      const { addItem } = useCartStore.getState();

      addItem(mockProduct, 2);

      const newState = useCartStore.getState();
      expect(newState.items).toHaveLength(1);
      expect(newState.items[0].sku_id).toBe(1);
      expect(newState.items[0].quantity).toBe(2);
      expect(newState.total).toBe(199.98);
    });

    it('should update quantity when adding existing item', () => {
      const { addItem } = useCartStore.getState();

      addItem(mockProduct, 2);
      addItem(mockProduct, 3);

      const newState = useCartStore.getState();
      expect(newState.items).toHaveLength(1);
      expect(newState.items[0].quantity).toBe(5);
      expect(newState.total).toBe(499.95);
    });

    it('should add item with group_id', () => {
      const { addItem } = useCartStore.getState();

      addItem(mockProduct, 1, 5);

      const newState = useCartStore.getState();
      expect(newState.items[0].group_id).toBe(5);
    });

    it('should add different products as separate items', () => {
      const { addItem } = useCartStore.getState();

      addItem(mockProduct, 2);
      addItem(mockProduct2, 1);

      const newState = useCartStore.getState();
      expect(newState.items).toHaveLength(2);
      expect(newState.total).toBe(249.97);
    });
  });

  describe('removeItem', () => {
    it('should remove item from cart', () => {
      const { addItem, removeItem } = useCartStore.getState();

      addItem(mockProduct, 2);
      const itemId = useCartStore.getState().items[0].id;
      removeItem(itemId);

      const newState = useCartStore.getState();
      expect(newState.items).toHaveLength(0);
      expect(newState.total).toBe(0);
    });

    it('should update total after removal', () => {
      const { addItem, removeItem } = useCartStore.getState();

      addItem(mockProduct, 2);
      addItem(mockProduct2, 1);
      const itemId = useCartStore.getState().items[0].id;
      removeItem(itemId);

      const newState = useCartStore.getState();
      expect(newState.items).toHaveLength(1);
      expect(newState.total).toBe(49.99);
    });
  });

  describe('updateQuantity', () => {
    it('should update item quantity', () => {
      const { addItem, updateQuantity } = useCartStore.getState();

      addItem(mockProduct, 2);
      const itemId = useCartStore.getState().items[0].id;
      updateQuantity(itemId, 5);

      const newState = useCartStore.getState();
      expect(newState.items[0].quantity).toBe(5);
      expect(newState.total).toBe(499.95);
    });

    it('should update total when quantity changes', () => {
      const { addItem, updateQuantity } = useCartStore.getState();

      addItem(mockProduct, 1);
      const itemId = useCartStore.getState().items[0].id;
      updateQuantity(itemId, 3);

      const newState = useCartStore.getState();
      expect(newState.total).toBeCloseTo(299.97, 2);
    });
  });

  describe('clear', () => {
    it('should clear all items from cart', () => {
      const { addItem, clear } = useCartStore.getState();

      addItem(mockProduct, 2);
      addItem(mockProduct2, 1);
      clear();

      const newState = useCartStore.getState();
      expect(newState.items).toHaveLength(0);
      expect(newState.total).toBe(0);
    });
  });

  describe('getTotal', () => {
    it('should return current total', () => {
      const { addItem, getTotal } = useCartStore.getState();

      addItem(mockProduct, 2);

      expect(getTotal()).toBe(199.98);
    });

    it('should return 0 for empty cart', () => {
      const { getTotal } = useCartStore.getState();

      expect(getTotal()).toBe(0);
    });
  });
});
