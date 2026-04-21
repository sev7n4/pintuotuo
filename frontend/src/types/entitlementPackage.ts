export interface EntitlementPackageItem {
  id: number;
  sku_id: number;
  sku_code: string;
  spu_name: string;
  sku_type: string;
  default_quantity: number;
  retail_price: number;
  /** 运营配置：对用户展示的短名称，可覆盖敏感/技术向 SPU 名 */
  display_name?: string;
  /** 运营配置：单项价值说明 */
  value_note?: string;
  sku_status?: string;
  spu_status?: string;
  stock?: number;
  line_purchasable?: boolean;
  line_issue?: string;
  /** 仅「我的权益」接口：当前账号是否已具备该明细 */
  line_covered?: boolean;
  /** 来自 skus，与后台套餐明细接口一致 */
  token_amount?: number;
  subscription_period?: string;
  valid_days?: number;
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
  category_code?: string;
  badge_text_secondary?: string;
  marketing_line?: string;
  promo_label?: string;
  promo_ends_at?: string;
  /** 后端聚合：是否满足 SKU+SPU 在售且库存足够 */
  purchasable?: boolean;
  unavailable_reason?: string;
  items: EntitlementPackageItem[];
  created_at: string;
  updated_at: string;
}

export interface EntitlementPackageUserView extends EntitlementPackage {
  covered_items: number;
  total_items: number;
  is_active: boolean;
}

/** GET /entitlement-packages/stats 单条（登录时可能含 user_*） */
export interface EntitlementPackageStatRow {
  package_id: number;
  favorite_count: number;
  like_count: number;
  sales_count: number;
  review_count: number;
  user_favorited?: boolean;
  user_liked?: boolean;
  user_reviewed?: boolean;
}

/** 前台 /packages 分类筛选（不含 general，「通用」类套餐仅在「全部」下展示） */
export const ENTITLEMENT_PACKAGE_FILTER_OPTIONS: { value: string; label: string }[] = [
  { value: 'all', label: '全部' },
  { value: 'personal', label: '个人月包' },
  { value: 'boost', label: '加油包' },
  { value: 'team', label: '小团队' },
  { value: 'enterprise', label: '企业' },
];

/** 管理端编辑分类（含 general） */
export const ENTITLEMENT_CATEGORY_ADMIN_OPTIONS: { value: string; label: string }[] = [
  { value: 'general', label: '通用' },
  { value: 'personal', label: '个人月包' },
  { value: 'boost', label: '加油包' },
  { value: 'team', label: '小团队' },
  { value: 'enterprise', label: '企业' },
];
