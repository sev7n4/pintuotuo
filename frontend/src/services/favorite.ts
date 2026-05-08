import api from './api';
import { APIResponse, Product } from '@/types';

/** 单品收藏（与历史接口字段兼容） */
export interface FavoriteSKUItem {
  item_type: 'sku';
  id: number;
  sku_id: number;
  product: Product;
  created_at: string;
}

/** 套餐包收藏（与套餐页 / entitlement-package 收藏 API 同源） */
export interface FavoriteEntitlementPackageItem {
  item_type: 'entitlement_package';
  id: number;
  entitlement_package_id: number;
  entitlement_package: {
    id: number;
    package_code: string;
    name: string;
    marketing_line?: string;
    status: string;
  };
  created_at: string;
}

export type FavoriteListItem = FavoriteSKUItem | FavoriteEntitlementPackageItem;

/** @deprecated 请用 FavoriteListItem；保留别名以免旧引用报错 */
export type FavoriteItem = FavoriteSKUItem;

export interface FavoriteListResponse {
  items: FavoriteListItem[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

export interface BrowseHistoryItem {
  id: number;
  sku_id: number;
  product: Product;
  view_count: number;
  viewed_at: string;
}

export interface BrowseHistoryListResponse {
  items: BrowseHistoryItem[];
  total: number;
  page: number;
  page_size: number;
  total_page: number;
}

export const favoriteService = {
  getFavorites: (page = 1, pageSize = 20) =>
    api.get<APIResponse<FavoriteListResponse>>('/favorites', {
      params: { page, page_size: pageSize },
    }),

  addFavorite: (skuId: number) =>
    api.post<APIResponse<{ id: number; sku_id: number }>>('/favorites', {
      sku_id: skuId,
    }),

  removeFavorite: (skuId: number) => api.delete<APIResponse<null>>(`/favorites/${skuId}`),

  checkFavorite: (skuId: number) =>
    api.get<APIResponse<{ is_favorite: boolean }>>(`/favorites/check/${skuId}`),
};

export const browseHistoryService = {
  getHistory: (params?: { endpoint_type?: string }, page = 1, pageSize = 20) =>
    api.get<APIResponse<BrowseHistoryListResponse>>('/browse-history', {
      params: { page, page_size: pageSize, ...params },
    }),

  addHistory: (skuId: number) =>
    api.post<APIResponse<null>>('/browse-history', {
      sku_id: skuId,
    }),

  clearHistory: () => api.delete<APIResponse<null>>('/browse-history'),

  removeHistoryItem: (skuId: number) => api.delete<APIResponse<null>>(`/browse-history/${skuId}`),
};
