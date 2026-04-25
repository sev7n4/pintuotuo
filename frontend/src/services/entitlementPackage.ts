import api from './api';
import type {
  EntitlementPackage,
  EntitlementPackageStatRow,
  EntitlementPackageUserView,
} from '@/types/entitlementPackage';

export type EntitlementPackageUpsertPayload = {
  package_code?: string;
  name: string;
  description?: string;
  status?: 'active' | 'inactive';
  sort_order?: number;
  start_at?: string;
  end_at?: string;
  is_featured?: boolean;
  badge_text?: string;
  category_code?: string;
  badge_text_secondary?: string;
  marketing_line?: string;
  promo_label?: string;
  promo_ends_at?: string;
  items: Array<{
    sku_id: number;
    default_quantity: number;
    display_name?: string;
    value_note?: string;
  }>;
};

export const entitlementPackageService = {
  /** 批量社交与销量统计；未登录也可调用（无 user_* 字段） */
  batchStats: (ids: number[]) =>
    api.get<{ data: EntitlementPackageStatRow[] }>('/entitlement-packages/stats', {
      params: { ids: ids.join(',') },
    }),
  addFavorite: (packageId: number) =>
    api.post<{ data: { favorited: boolean; favorite_count: number } }>(
      `/entitlement-packages/${packageId}/favorite`
    ),
  removeFavorite: (packageId: number) =>
    api.delete<{ data: { favorited: boolean; favorite_count: number } }>(
      `/entitlement-packages/${packageId}/favorite`
    ),
  toggleLike: (packageId: number) =>
    api.post<{ data: { liked: boolean; like_count: number } }>(
      `/entitlement-packages/${packageId}/like`
    ),
  upsertReview: (packageId: number, body: { rating: number; comment?: string }) =>
    api.post<{ data: { review_count: number } }>(
      `/entitlement-packages/${packageId}/reviews`,
      body
    ),
  listAdmin: () => api.get<{ data: EntitlementPackage[] }>('/admin/entitlement-packages'),
  createAdmin: (data: EntitlementPackageUpsertPayload) =>
    api.post('/admin/entitlement-packages', data),
  updateAdmin: (id: number, data: EntitlementPackageUpsertPayload) =>
    api.put(`/admin/entitlement-packages/${id}`, data),
  deleteAdmin: (id: number) => api.delete(`/admin/entitlement-packages/${id}`),
  listPublic: () => api.get<{ data: EntitlementPackage[] }>('/entitlement-packages'),
  listMine: () =>
    api.get<{ data: EntitlementPackageUserView[] }>('/users/me/entitlements/packages'),
};
