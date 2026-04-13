import api from './api';
import { User, APIResponse } from '@/types';

interface LoginRequest {
  email: string;
  password: string;
}

interface LoginResponse {
  user: User;
  token: string;
}

interface RegisterRequest {
  email: string;
  password: string;
  role?: string;
}

export const authService = {
  register: (data: RegisterRequest) => api.post<APIResponse<LoginResponse>>('/users/register', data),

  login: (data: LoginRequest) => api.post<APIResponse<LoginResponse>>('/users/login', data),

  logout: () => api.post<APIResponse<void>>('/users/logout'),

  sendSmsCode: (phone: string, scene?: string) =>
    api.post<{ message?: string; debug_code?: string }>('/users/sms/send', { phone, scene }),

  sendEmailMagicLink: (email: string) =>
    api.post<{ message?: string; debug_link?: string }>('/users/email/magic/send', { email }),

  registerWithSms: (data: { phone: string; code: string; password: string; role?: string }) =>
    api.post<APIResponse<LoginResponse>>('/users/sms/register', data),

  loginWithSms: (phone: string, code: string) =>
    api.post<APIResponse<LoginResponse>>('/users/sms/login', { phone, code }),
};
