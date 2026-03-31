import { create } from 'zustand';
import axios from 'axios';
import { Token, TokenTransaction, UserAPIKey, RechargeOrder } from '@/types';
import { tokenService } from '@/services/token';

function getApiErrorMessage(err: unknown, fallback: string): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data as { message?: string; error?: string } | undefined;
    if (data && typeof data === 'object') {
      if (typeof data.message === 'string' && data.message) return data.message;
      if (typeof data.error === 'string' && data.error) return data.error;
    }
  }
  if (err instanceof Error) return err.message;
  return fallback;
}

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
  transfer: (
    amount: number,
    opts: { recipientId?: number; recipientEmail?: string }
  ) => Promise<boolean>;
  mockCompleteRechargeOrder: (orderId: number) => Promise<boolean>;
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
      set({ error: getApiErrorMessage(error, '获取余额失败'), isLoading: false });
    }
  },

  fetchTransactions: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.getConsumption();
      set({ transactions: response.data || [], isLoading: false });
    } catch (error) {
      set({ error: getApiErrorMessage(error, '获取交易记录失败'), isLoading: false });
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
      set({ error: getApiErrorMessage(error, '获取API密钥失败'), isLoading: false });
    }
  },

  fetchRechargeOrders: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.getRechargeOrders(1, 20);
      const rechargeOrders = response.data?.data?.data || [];
      set({ rechargeOrders, isLoading: false });
    } catch (error) {
      set({ error: getApiErrorMessage(error, '获取充值订单失败'), isLoading: false });
    }
  },

  createAPIKey: async (name) => {
    set({ isLoading: true, error: null });
    try {
      const response = await tokenService.createAPIKey(name);
      set({ isLoading: false });
      return response.data?.key || null;
    } catch (error) {
      set({ error: getApiErrorMessage(error, '创建API密钥失败'), isLoading: false });
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
      set({ error: getApiErrorMessage(error, '删除API密钥失败'), isLoading: false });
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
      set({ error: getApiErrorMessage(error, '创建充值订单失败'), isLoading: false });
      return null;
    }
  },

  transfer: async (amount, opts) => {
    set({ isLoading: true, error: null });
    try {
      await tokenService.transfer(amount, opts);
      set({ isLoading: false });
      return true;
    } catch (error) {
      set({ error: getApiErrorMessage(error, '转账失败'), isLoading: false });
      return false;
    }
  },

  mockCompleteRechargeOrder: async (orderId) => {
    set({ isLoading: true, error: null });
    try {
      await tokenService.mockCompleteRechargeOrder(orderId);
      set({ isLoading: false });
      return true;
    } catch (error) {
      set({ error: getApiErrorMessage(error, '模拟支付失败'), isLoading: false });
      return false;
    }
  },

  clearError: () => set({ error: null }),
}));
