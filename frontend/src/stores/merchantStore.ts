import { create } from 'zustand';
import {
  Merchant,
  MerchantStats,
  MerchantSettlement,
  MerchantOrder,
  Product,
  MerchantAPIKey,
  APIKeyUsage,
  PaginatedResponse,
} from '@/types';
import { merchantService } from '@/services/merchant';

interface MerchantState {
  profile: Merchant | null;
  stats: MerchantStats | null;
  products: Product[];
  orders: MerchantOrder[];
  settlements: MerchantSettlement[];
  apiKeys: MerchantAPIKey[];
  apiKeyUsage: APIKeyUsage[];
  isLoading: boolean;
  error: string | null;

  fetchProfile: () => Promise<void>;
  updateProfile: (data: Partial<Merchant>) => Promise<boolean>;
  fetchStats: () => Promise<void>;
  fetchProducts: (page?: number, perPage?: number, status?: string) => Promise<void>;
  fetchOrders: (page?: number, perPage?: number, status?: string) => Promise<void>;
  fetchSettlements: () => Promise<void>;
  requestSettlement: () => Promise<boolean>;
  fetchAPIKeys: () => Promise<void>;
  createAPIKey: (data: {
    name: string;
    provider: string;
    api_key: string;
    api_secret?: string;
    quota_limit?: number;
  }) => Promise<boolean>;
  updateAPIKey: (id: number, data: Partial<MerchantAPIKey>) => Promise<boolean>;
  deleteAPIKey: (id: number) => Promise<boolean>;
  fetchAPIKeyUsage: () => Promise<void>;
  clearError: () => void;
}

export const useMerchantStore = create<MerchantState>((set) => ({
  profile: null,
  stats: null,
  products: [],
  orders: [],
  settlements: [],
  apiKeys: [],
  apiKeyUsage: [],
  isLoading: false,
  error: null,

  fetchProfile: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await merchantService.getProfile();
      set({ profile: response.data, isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取商家信息失败';
      set({ error: message, isLoading: false });
    }
  },

  updateProfile: async (data: Partial<Merchant>) => {
    set({ isLoading: true, error: null });
    try {
      const response = await merchantService.updateProfile(data);
      set({ profile: response.data, isLoading: false });
      return true;
    } catch (error) {
      const message = error instanceof Error ? error.message : '更新商家信息失败';
      set({ error: message, isLoading: false });
      return false;
    }
  },

  fetchStats: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await merchantService.getStats();
      set({ stats: response.data, isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取统计数据失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchProducts: async (page = 1, perPage = 20, status?: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await merchantService.getProducts(page, perPage, status);
      const data = response.data as unknown as PaginatedResponse<Product>;
      set({ products: data?.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取商品列表失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchOrders: async (page = 1, perPage = 20, status?: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await merchantService.getOrders(page, perPage, status);
      const data = response.data as unknown as PaginatedResponse<MerchantOrder>;
      set({ orders: data?.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取订单列表失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchSettlements: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await merchantService.getSettlements();
      set({ settlements: response.data?.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取结算记录失败';
      set({ error: message, isLoading: false });
    }
  },

  requestSettlement: async () => {
    set({ isLoading: true, error: null });
    try {
      await merchantService.requestSettlement();
      set({ isLoading: false });
      return true;
    } catch (error) {
      const message = error instanceof Error ? error.message : '申请结算失败';
      set({ error: message, isLoading: false });
      return false;
    }
  },

  fetchAPIKeys: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await merchantService.getAPIKeys();
      set({ apiKeys: response.data?.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取API密钥失败';
      set({ error: message, isLoading: false });
    }
  },

  createAPIKey: async (data) => {
    set({ isLoading: true, error: null });
    try {
      await merchantService.createAPIKey(data);
      set({ isLoading: false });
      return true;
    } catch (error) {
      const message = error instanceof Error ? error.message : '创建API密钥失败';
      set({ error: message, isLoading: false });
      return false;
    }
  },

  updateAPIKey: async (id, data) => {
    set({ isLoading: true, error: null });
    try {
      await merchantService.updateAPIKey(id, data);
      set({ isLoading: false });
      return true;
    } catch (error) {
      const message = error instanceof Error ? error.message : '更新API密钥失败';
      set({ error: message, isLoading: false });
      return false;
    }
  },

  deleteAPIKey: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await merchantService.deleteAPIKey(id);
      set({ isLoading: false });
      return true;
    } catch (error) {
      const message = error instanceof Error ? error.message : '删除API密钥失败';
      set({ error: message, isLoading: false });
      return false;
    }
  },

  fetchAPIKeyUsage: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await merchantService.getAPIKeyUsage();
      set({ apiKeyUsage: response.data?.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取使用情况失败';
      set({ error: message, isLoading: false });
    }
  },

  clearError: () => set({ error: null }),
}));
