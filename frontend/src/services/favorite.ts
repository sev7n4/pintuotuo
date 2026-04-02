import api from './api';
import { APIResponse, Product } from '@/types';

export interface FavoriteItem {
  id: number;
  sku_id: number;
  product: Product;
  created_at: string;
}

export interface FavoriteListResponse {
  items: FavoriteItem[];
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
  getHistory: (page = 1, pageSize = 20) =>
    api.get<APIResponse<BrowseHistoryListResponse>>('/browse-history', {
      params: { page, page_size: pageSize },
    }),

  addHistory: (skuId: number) =>
    api.post<APIResponse<null>>('/browse-history', {
      sku_id: skuId,
    }),

  clearHistory: () => api.delete<APIResponse<null>>('/browse-history'),

  removeHistoryItem: (skuId: number) => api.delete<APIResponse<null>>(`/browse-history/${skuId}`),
};
