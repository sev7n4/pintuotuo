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
  name: string;
  password: string;
  role?: string;
}

export const authService = {
  register: (data: RegisterRequest) =>
    api.post<APIResponse<LoginResponse>>('/users/register', data),

  login: (data: LoginRequest) => api.post<APIResponse<LoginResponse>>('/users/login', data),

  logout: () => api.post<APIResponse<void>>('/users/logout'),
};
