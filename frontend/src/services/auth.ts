import api from './api'
import { User, APIResponse } from '@types/index'

interface LoginRequest {
  email: string
  password: string
}

interface LoginResponse {
  user: User
  token: string
}

interface RegisterRequest {
  email: string
  name: string
  password: string
}

export const authService = {
  // Register new user
  register: (data: RegisterRequest) =>
    api.post<APIResponse<LoginResponse>>('/users/register', data),

  // Login user
  login: (data: LoginRequest) =>
    api.post<APIResponse<LoginResponse>>('/users/login', data),

  // Logout user
  logout: () =>
    api.post<APIResponse<void>>('/users/logout'),
}
