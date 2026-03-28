import { create } from 'zustand';
import { User, APIResponse } from '@/types';
import { authService } from '@/services/auth';
import { userService } from '@/services/user';

interface LoginResponse {
  user: User;
  token: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
  isAuthenticated: boolean;
  rememberMe: boolean;

  login: (email: string, password: string, rememberMe?: boolean) => Promise<void>;
  register: (email: string, name: string, password: string, role?: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchUser: () => Promise<void>;
  setUser: (user: User) => void;
  clearError: () => void;
  setRememberMe: (remember: boolean) => void;
}

const getTokenFromStorage = (): string | null => {
  return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token');
};

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: getTokenFromStorage(),
  isLoading: false,
  error: null,
  isAuthenticated: !!getTokenFromStorage(),
  rememberMe: localStorage.getItem('remember_me') === 'true',

  login: async (email, password, rememberMe = false) => {
    set({ isLoading: true, error: null });
    try {
      const response = await authService.login({ email, password });
      const apiResponse = response.data as APIResponse<LoginResponse>;
      const data = apiResponse.data;
      if (data) {
        const { user, token } = data;
        if (rememberMe) {
          localStorage.setItem('auth_token', token);
          localStorage.setItem('remember_me', 'true');
        } else {
          sessionStorage.setItem('auth_token', token);
          localStorage.removeItem('remember_me');
        }
        set({ user, token, isAuthenticated: true, isLoading: false, rememberMe });
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : '登录失败';
      set({ error: message, isLoading: false });
      throw error;
    }
  },

  register: async (email, name, password, role = 'user') => {
    set({ isLoading: true, error: null });
    try {
      const response = await authService.register({ email, name, password, role });
      const apiResponse = response.data as APIResponse<LoginResponse>;
      const data = apiResponse.data;
      if (data) {
        const { user, token } = data;
        localStorage.setItem('auth_token', token);
        localStorage.setItem('remember_me', 'true');
        set({ user, token, isAuthenticated: true, isLoading: false, rememberMe: true });
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : '注册失败';
      set({ error: message, isLoading: false });
      throw error;
    }
  },

  logout: async () => {
    set({ isLoading: true });
    try {
      await authService.logout();
    } finally {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('remember_me');
      sessionStorage.removeItem('auth_token');
      set({ user: null, token: null, isAuthenticated: false, isLoading: false, rememberMe: false });
    }
  },

  fetchUser: async () => {
    const token = getTokenFromStorage();
    if (!token) {
      set({ isAuthenticated: false, user: null });
      return;
    }

    set({ isLoading: true, error: null });
    try {
      const response = await userService.getCurrentUser();
      const apiResponse = response.data as APIResponse<User>;
      const fetchedUser = apiResponse.data || null;
      set({ user: fetchedUser, isLoading: false, isAuthenticated: true });
    } catch (error) {
      localStorage.removeItem('auth_token');
      sessionStorage.removeItem('auth_token');
      set({ user: null, token: null, isAuthenticated: false, isLoading: false });
    }
  },

  setUser: (user) => set({ user }),

  clearError: () => set({ error: null }),

  setRememberMe: (remember) => {
    if (remember) {
      localStorage.setItem('remember_me', 'true');
    } else {
      localStorage.removeItem('remember_me');
    }
    set({ rememberMe: remember });
  },
}));
