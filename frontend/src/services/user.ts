import api from './api';
import { User, UserIdentity, APIResponse } from '@/types';

export const userService = {
  // Get current user
  getCurrentUser: () => api.get<APIResponse<User>>('/users/me'),

  /** 已绑定的第三方身份（微信/GitHub 等） */
  getMyIdentities: () => api.get<{ code: number; data: UserIdentity[] }>('/users/me/identities'),

  // Update current user
  updateCurrentUser: (data: Partial<User>) => api.put<APIResponse<User>>('/users/me', data),

  // Get user by ID
  getUserByID: (id: number) => api.get<APIResponse<User>>(`/users/${id}`),

  // Update user by ID (admin)
  updateUser: (id: number, data: Partial<User>) => api.put<APIResponse<User>>(`/users/${id}`, data),
};
