import api from './api';
import type { EntitlementPackage, EntitlementPackageUserView } from '@/types/entitlementPackage';

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
