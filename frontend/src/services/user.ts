import api from './api'
import { User, APIResponse } from '@/types'

export const userService = {
  // Get current user
  getCurrentUser: () => api.get<APIResponse<User>>('/users/me'),

  // Update current user
  updateCurrentUser: (data: Partial<User>) =>
    api.put<APIResponse<User>>('/users/me', data),

  // Upload avatar
  uploadAvatar: async (file: File) => {
    const formData = new FormData()
    formData.append('avatar', file)
    return api.post<{ code: number; message: string; data: { url: string } }>(
      '/users/avatar',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )
  },

  // Get user by ID
  getUserByID: (id: number) =>
    api.get<APIResponse<User>>(`/users/${id}`),

  // Update user by ID (admin)
  updateUser: (id: number, data: Partial<User>) =>
    api.put<APIResponse<User>>(`/users/${id}`, data),
}
