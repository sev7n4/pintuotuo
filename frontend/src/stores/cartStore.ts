import { create } from 'zustand';
import { CartItem, Product } from '@/types';

interface CartState {
  items: CartItem[];
  total: number;

  addItem: (product: Product, quantity: number, groupId?: number) => void;
  removeItem: (id: string) => void;
  updateQuantity: (id: string, quantity: number) => void;
  clear: () => void;
  getTotal: () => number;
}

export const useCartStore = create<CartState>((set, get) => ({
  items: [],
  total: 0,

  addItem: (product, quantity, groupId) => {
    set((state) => {
      const existingItem = state.items.find(
        (item) => item.sku_id === product.id && item.group_id === groupId
      );

      let newItems: CartItem[];
      if (existingItem) {
        newItems = state.items.map((item) =>
          item.id === existingItem.id ? { ...item, quantity: item.quantity + quantity } : item
        );
      } else {
        newItems = [
          ...state.items,
          {
            id: `${product.id}-${groupId || 0}-${Date.now()}`,
            sku_id: product.id,
            product,
            quantity,
            group_id: groupId,
          },
        ];
      }

      return {
        items: newItems,
        total: newItems.reduce((sum, item) => sum + item.product.price * item.quantity, 0),
      };
    });
  },

  removeItem: (id) => {
    set((state) => {
      const newItems = state.items.filter((item) => item.id !== id);
      return {
        items: newItems,
        total: newItems.reduce((sum, item) => sum + item.product.price * item.quantity, 0),
      };
    });
  },

  updateQuantity: (id, quantity) => {
    set((state) => {
      const newItems = state.items.map((item) => (item.id === id ? { ...item, quantity } : item));
      return {
        items: newItems,
        total: newItems.reduce((sum, item) => sum + item.product.price * item.quantity, 0),
      };
    });
  },

  clear: () => set({ items: [], total: 0 }),

  getTotal: () => get().total,
}));
