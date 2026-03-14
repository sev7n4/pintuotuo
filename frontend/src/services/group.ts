import api from './api'
import { Group, APIResponse, PaginatedResponse } from '@types/index'

interface CreateGroupRequest {
  product_id: number
  target_count: number
  deadline: string
}

export const groupService = {
  // Create group
  createGroup: (data: CreateGroupRequest) =>
    api.post<APIResponse<Group>>('/groups', data),

  // List groups
  listGroups: (page?: number, per_page?: number) =>
    api.get<APIResponse<PaginatedResponse<Group>>>('/groups', {
      params: { page, per_page },
    }),

  // Get group by ID
  getGroupByID: (id: number) =>
    api.get<APIResponse<Group>>(`/groups/${id}`),

  // Join group
  joinGroup: (id: number) =>
    api.post<APIResponse<Group>>(`/groups/${id}/join`, {}),

  // Cancel group
  cancelGroup: (id: number) =>
    api.delete<APIResponse<void>>(`/groups/${id}`),

  // Get group progress
  getGroupProgress: (id: number) =>
    api.get<APIResponse<Group>>(`/groups/${id}/progress`),
}
