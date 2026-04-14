export interface EntitlementPackageItem {
  id: number;
  sku_id: number;
  sku_code: string;
  spu_name: string;
  sku_type: string;
  default_quantity: number;
  retail_price: number;
}

export interface EntitlementPackage {
  id: number;
  package_code: string;
  name: string;
  description?: string;
  status: 'active' | 'inactive';
  sort_order: number;
  start_at?: string;
  end_at?: string;
  is_featured: boolean;
  badge_text?: string;
  items: EntitlementPackageItem[];
  created_at: string;
  updated_at: string;
}

export interface EntitlementPackageUserView extends EntitlementPackage {
  covered_items: number;
  total_items: number;
  is_active: boolean;
}
