import { create } from 'zustand';
import { Token, TokenTransaction, UserAPIKey, RechargeOrder } from '@/types';
import { tokenService } from '@/services/token';

interface TokenState {
  balance: Token | null;
  transactions: TokenTransaction[];
  apiKeys: UserAPIKey[];
  rechargeOrders: RechargeOrder[];
  isLoading: boolean;
  error: string | null;

  fetchBalance: () => Promise<void>;
  fetchTransactions: () => Promise<void>;
  fetchAPIKeys: () => Promise<void>;
  fetchRechargeOrders: () => Promise<void>;
  createAPIKey: (name: string) => Promise<string | null>;
  deleteAPIKey: (id: number) => Promise<boolean>;
  createRechargeOrder: (
    amount: number,
    method: 'alipay' | 'wechat' | 'balance'
  ) => Promise<RechargeOrder | null>;
  transfer: (recipientId: number, amount: number) => Promise<boolean>;
  clearError: () => void;
}

export const useTokenStore = create<TokenState>((set) => ({
  balance: null,
  transactions: [],
  apiKeys: [],
  rechargeOrders: [],
  isLoading: false,
  error: null,

  fetchBalance: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.getBalance();
      set({ balance: response.data, isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取余额失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchTransactions: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.getConsumption();
      set({ transactions: response.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取交易记录失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchAPIKeys: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.getAPIKeys();
      const payload = response.data;
      const apiKeys = Array.isArray(payload) ? payload : payload?.data || [];
      set({ apiKeys, isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取API密钥失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchRechargeOrders: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.getRechargeOrders(1, 20);
      const rechargeOrders = response.data?.data?.data || [];
      set({ rechargeOrders, isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取充值订单失败';
      set({ error: message, isLoading: false });
    }
  },

  createAPIKey: async (name) => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.createAPIKey(name);
      set({ isLoading: false });
      return response.data?.key || null;
    } catch (error) {
      const message = error instanceof Error ? error.message : '创建API密钥失败';
      set({ error: message, isLoading: false });
      return null;
    }
  },

  deleteAPIKey: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await tokenService.deleteAPIKey(id);
      set({ isLoading: false });
      return true;
    } catch (error) {
      const message = error instanceof Error ? error.message : '删除API密钥失败';
      set({ error: message, isLoading: false });
      return false;
    }
  },

  createRechargeOrder: async (amount, method) => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.createRechargeOrder(amount, method);
      set({ isLoading: false });
      return response.data?.data || null;
    } catch (error) {
      const message = error instanceof Error ? error.message : '创建充值订单失败';
      set({ error: message, isLoading: false });
      return null;
    }
  },

  transfer: async (recipientId, amount) => {
    set({ isLoading: true, error: null });
    try {
      await tokenService.transfer(recipientId, amount);
      set({ isLoading: false });
      return true;
    } catch (error) {
      const message = error instanceof Error ? error.message : '转账失败';
      set({ error: message, isLoading: false });
      return false;
    }
  },

  clearError: () => set({ error: null }),
}));
