import api from './api';
import { Group, APIResponse, PaginatedResponse } from '@/types';

export interface CreateGroupRequest {
  sku_id: number;
  target_count: number;
  deadline: string;
}

export interface CreateGroupResponse {
  group: Group;
  order_id: number;
}

export interface JoinGroupResponse {
  group: Group;
  order_id: number;
}

/** GET /groups?scope= 与后端 ListGroups 一致 */
export type GroupListScope = 'all' | 'mine_involved' | 'mine_created' | 'mine_joined';

export const groupService = {
  // Create group
  createGroup: (data: CreateGroupRequest) =>
    api.post<APIResponse<CreateGroupResponse>>('/groups', data),

  // List groups（scope: all | mine_involved | mine_created | mine_joined）
  listGroups: (
    page?: number,
    per_page?: number,
    opts?: { scope?: GroupListScope; status?: string }
  ) =>
    api.get<APIResponse<PaginatedResponse<Group>>>('/groups', {
      params: {
        page,
        per_page,
        ...(opts?.scope && opts.scope !== 'all' ? { scope: opts.scope } : {}),
        ...(opts?.status ? { status: opts.status } : {}),
      },
    }),

  // Get group by ID
  getGroupByID: (id: number) => api.get<APIResponse<Group>>(`/groups/${id}`),

  // Join group
  joinGroup: (id: number) => api.post<APIResponse<JoinGroupResponse>>(`/groups/${id}/join`, {}),

  // Cancel group
  cancelGroup: (id: number) => api.delete<APIResponse<void>>(`/groups/${id}`),

  // Get group progress
  getGroupProgress: (id: number) => api.get<APIResponse<Group>>(`/groups/${id}/progress`),
};
