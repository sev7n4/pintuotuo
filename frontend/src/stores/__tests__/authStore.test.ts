import { useAuthStore } from '../authStore';
import { authService } from '@/services/auth';
import { userService } from '@/services/user';

jest.mock('@/services/auth');
jest.mock('@/services/user');

const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

const sessionStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });
Object.defineProperty(window, 'sessionStorage', { value: sessionStorageMock });

const mockedAuthService = authService as jest.Mocked<typeof authService>;
const mockedUserService = userService as jest.Mocked<typeof userService>;

describe('AuthStore', () => {
  beforeEach(() => {
    localStorage.clear();
    sessionStorage.clear();
    jest.clearAllMocks();
    // 重置 store 状态
    useAuthStore.setState({
      user: null,
      token: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,
      rememberMe: false,
    });
  });

  describe('initial state', () => {
    it('should have correct initial values', () => {
      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe('login', () => {
    it('should login successfully with rememberMe true', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 1,
              email: 'test@example.com',
              name: 'Test User',
              role: 'user',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          },
        },
      };

      mockedAuthService.login.mockResolvedValueOnce(mockResponse as any);

      const store = useAuthStore.getState();
      await store.login('test@example.com', 'password', true);

      const newState = useAuthStore.getState();
      expect(newState.user?.email).toBe('test@example.com');
      expect(newState.token).toBe('test-token');
      expect(newState.isAuthenticated).toBe(true);
      expect(newState.isLoading).toBe(false);
      expect(localStorage.getItem('auth_token')).toBe('test-token');
      expect(localStorage.getItem('remember_me')).toBe('true');
    });

    it('should login successfully with rememberMe false (default)', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 1,
              email: 'test@example.com',
              name: 'Test User',
              role: 'user',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          },
        },
      };

      mockedAuthService.login.mockResolvedValueOnce(mockResponse as any);

      const store = useAuthStore.getState();
      await store.login('test@example.com', 'password');

      const newState = useAuthStore.getState();
      expect(newState.user?.email).toBe('test@example.com');
      expect(newState.token).toBe('test-token');
      expect(newState.isAuthenticated).toBe(true);
      expect(newState.isLoading).toBe(false);
      expect(sessionStorage.getItem('auth_token')).toBe('test-token');
      expect(localStorage.getItem('auth_token')).toBeNull();
    });

    it('should handle login error', async () => {
      const errorMessage = 'Invalid credentials';
      mockedAuthService.login.mockRejectedValueOnce(new Error(errorMessage));

      const store = useAuthStore.getState();
      await expect(store.login('invalid@example.com', 'wrong')).rejects.toThrow(errorMessage);

      const newState = useAuthStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
      expect(newState.token).toBeNull();
    });
  });

  describe('register', () => {
    it('should register successfully as user', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 1,
              email: 'new@example.com',
              name: 'New User',
              role: 'user',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'new-token',
          },
        },
      };

      mockedAuthService.register.mockResolvedValueOnce(mockResponse as any);

      const store = useAuthStore.getState();
      await store.register('new@example.com', 'New User', 'password');

      const newState = useAuthStore.getState();
      expect(newState.user?.email).toBe('new@example.com');
      expect(newState.token).toBe('new-token');
      expect(newState.isAuthenticated).toBe(true);
    });

    it('should register successfully as merchant', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 2,
              email: 'merchant@example.com',
              name: 'Merchant User',
              role: 'merchant',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'merchant-token',
          },
        },
      };

      mockedAuthService.register.mockResolvedValueOnce(mockResponse as any);

      const store = useAuthStore.getState();
      await store.register('merchant@example.com', 'Merchant User', 'password', 'merchant');

      const newState = useAuthStore.getState();
      expect(newState.user?.email).toBe('merchant@example.com');
      expect(newState.user?.role).toBe('merchant');
      expect(newState.token).toBe('merchant-token');
    });

    it('should handle registration error', async () => {
      const errorMessage = 'Email already exists';
      mockedAuthService.register.mockRejectedValueOnce(new Error(errorMessage));

      const store = useAuthStore.getState();
      await expect(store.register('existing@example.com', 'User', 'password')).rejects.toThrow(
        errorMessage
      );

      const newState = useAuthStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('logout', () => {
    it('should logout successfully', async () => {
      // 先登录
      const mockLoginResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 1,
              email: 'test@example.com',
              name: 'Test User',
              role: 'user',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          },
        },
      };
      mockedAuthService.login.mockResolvedValueOnce(mockLoginResponse as any);

      const store = useAuthStore.getState();
      await store.login('test@example.com', 'password');

      // 测试登出
      mockedAuthService.logout.mockResolvedValueOnce({
        data: { code: 0, message: 'success' },
      } as any);
      await store.logout();

      const newState = useAuthStore.getState();
      expect(newState.user).toBeNull();
      expect(newState.token).toBeNull();
      expect(newState.isAuthenticated).toBe(false);
      expect(localStorage.getItem('auth_token')).toBeNull();
    });

    it('should logout even if API call fails', async () => {
      // 先登录
      const mockLoginResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 1,
              email: 'test@example.com',
              name: 'Test User',
              role: 'user',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          },
        },
      };
      mockedAuthService.login.mockResolvedValueOnce(mockLoginResponse as any);

      const store = useAuthStore.getState();
      await store.login('test@example.com', 'password');

      // 测试登出失败的情况
      const error = new Error('Logout failed');
      mockedAuthService.logout.mockRejectedValueOnce(error);

      // 使用 try-catch 捕获错误
      try {
        await store.logout();
      } catch (err) {
        // 登出应该在 API 失败时仍然清理状态
      }

      const newState = useAuthStore.getState();
      expect(newState.user).toBeNull();
      expect(newState.token).toBeNull();
      expect(newState.isAuthenticated).toBe(false);
      expect(localStorage.getItem('auth_token')).toBeNull();
    });
  });

  describe('fetchUser', () => {
    it('should fetch user successfully', async () => {
      // 先登录
      const mockLoginResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 1,
              email: 'test@example.com',
              name: 'Test User',
              role: 'user',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          },
        },
      };
      mockedAuthService.login.mockResolvedValueOnce(mockLoginResponse as any);

      const store = useAuthStore.getState();
      await store.login('test@example.com', 'password');

      // 测试获取用户信息
      const mockUserResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            id: 1,
            email: 'test@example.com',
            name: 'Test User',
            role: 'user',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        },
      };

      mockedUserService.getCurrentUser.mockResolvedValueOnce(mockUserResponse as any);
      await store.fetchUser();

      const newState = useAuthStore.getState();
      expect(newState.user?.email).toBe('test@example.com');
      expect(newState.isLoading).toBe(false);
    });

    it('should not fetch user if no token', async () => {
      const store = useAuthStore.getState();
      await store.fetchUser();
      expect(mockedUserService.getCurrentUser).not.toHaveBeenCalled();
    });

    it('should handle fetch user error', async () => {
      // 先登录
      const mockLoginResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 1,
              email: 'test@example.com',
              name: 'Test User',
              role: 'user',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          },
        },
      };
      mockedAuthService.login.mockResolvedValueOnce(mockLoginResponse as any);

      const store = useAuthStore.getState();
      await store.login('test@example.com', 'password');

      // 测试获取用户信息失败
      const errorMessage = 'Failed to fetch user';
      mockedUserService.getCurrentUser.mockRejectedValueOnce(new Error(errorMessage));

      await store.fetchUser();

      const newState = useAuthStore.getState();
      expect(newState.user).toBeNull();
      expect(newState.isAuthenticated).toBe(false);
      expect(newState.error).toBeNull();
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('setUser', () => {
    it('should set user', () => {
      const testUser = {
        id: 1,
        email: 'test@example.com',
        name: 'Test User',
        role: 'user' as 'user' | 'merchant' | 'admin',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      const store = useAuthStore.getState();
      store.setUser(testUser);

      const newState = useAuthStore.getState();
      expect(newState.user).toEqual(testUser);
    });
  });

  describe('clearError', () => {
    it('should clear error', () => {
      const store = useAuthStore.getState();
      store.clearError();
      expect(store.error).toBeNull();
    });
  });

  describe('isAuthenticated', () => {
    it('should be true when token exists', async () => {
      const mockLoginResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            user: {
              id: 1,
              email: 'test@example.com',
              name: 'Test User',
              role: 'user',
              created_at: '2024-01-01T00:00:00Z',
              updated_at: '2024-01-01T00:00:00Z',
            },
            token: 'test-token',
          },
        },
      };
      mockedAuthService.login.mockResolvedValueOnce(mockLoginResponse as any);

      const store = useAuthStore.getState();
      await store.login('test@example.com', 'password');

      // 重新获取最新的 store 状态
      const updatedStore = useAuthStore.getState();
      expect(updatedStore.isAuthenticated).toBe(true);
    });

    it('should be false when no token', () => {
      // 清理 localStorage 并重置状态
      localStorage.clear();
      sessionStorage.clear();
      useAuthStore.setState({
        user: null,
        token: null,
        isLoading: false,
        error: null,
        isAuthenticated: false,
        rememberMe: false,
      });
      const store = useAuthStore.getState();
      expect(store.isAuthenticated).toBe(false);
    });
  });
});
