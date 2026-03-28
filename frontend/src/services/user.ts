import api from './api';
import { User, APIResponse } from '@/types';

export const userService = {
  // Get current user
  getCurrentUser: () => api.get<APIResponse<User>>('/users/me'),

  // Update current user
  updateCurrentUser: (data: Partial<User>) => api.put<APIResponse<User>>('/users/me', data),

  // Get user by ID
  getUserByID: (id: number) => api.get<APIResponse<User>>(`/users/${id}`),

  // Update user by ID (admin)
  updateUser: (id: number, data: Partial<User>) => api.put<APIResponse<User>>(`/users/${id}`, data),
};
