import axios, { AxiosInstance, AxiosError, AxiosResponse } from 'axios';

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

console.log('API Base URL:', BASE_URL);

const instance: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
});

instance.interceptors.request.use(
  (config) => {
    console.log('API Request:', config.method?.toUpperCase(), config.url);
    const token = localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

instance.interceptors.response.use(
  (response) => {
    console.log('API Response:', response.status, response.config.url);
    return response;
  },
  (error: AxiosError) => {
    console.error('API Error:', error.message, error.config?.url);
    if (error.response?.status === 401) {
      const url = String(error.config?.url ?? '');
      const isAuthLogin =
        url.includes('/users/login') ||
        url.includes('/users/sms/login') ||
        url.includes('/email/magic/verify');
      if (!isAuthLogin) {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('remember_me');
        sessionStorage.removeItem('auth_token');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export default instance as {
  get<T = unknown>(url: string, config?: object): Promise<AxiosResponse<T>>;
  post<T = unknown>(url: string, data?: unknown, config?: object): Promise<AxiosResponse<T>>;
  put<T = unknown>(url: string, data?: unknown, config?: object): Promise<AxiosResponse<T>>;
  patch<T = unknown>(url: string, data?: unknown, config?: object): Promise<AxiosResponse<T>>;
  delete<T = unknown>(url: string, config?: object): Promise<AxiosResponse<T>>;
};
