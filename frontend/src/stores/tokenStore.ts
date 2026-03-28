import { create } from 'zustand';
import { Token, TokenTransaction, UserAPIKey } from '@/types';
import { tokenService } from '@/services/token';

interface TokenState {
  balance: Token | null;
  transactions: TokenTransaction[];
  apiKeys: UserAPIKey[];
  isLoading: boolean;
  error: string | null;

  fetchBalance: () => Promise<void>;
  fetchTransactions: () => Promise<void>;
  fetchAPIKeys: () => Promise<void>;
  createAPIKey: (name: string) => Promise<boolean>;
  deleteAPIKey: (id: number) => Promise<boolean>;
  transfer: (recipientId: number, amount: number) => Promise<boolean>;
  clearError: () => void;
}

export const useTokenStore = create<TokenState>((set) => ({
  balance: null,
  transactions: [],
  apiKeys: [],
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
      set({ apiKeys: response.data?.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取API密钥失败';
      set({ error: message, isLoading: false });
    }
  },

  createAPIKey: async (name) => {
    set({ isLoading: true, error: null });
    try {
      await tokenService.createAPIKey(name);
      set({ isLoading: false });
      return true;
    } catch (error) {
      const message = error instanceof Error ? error.message : '创建API密钥失败';
      set({ error: message, isLoading: false });
      return false;
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
