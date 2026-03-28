import api from './api';
import { APIResponse, Product } from '@/types';

export interface FavoriteItem {
  id: number;
  product_id: number;
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
  product_id: number;
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

  addFavorite: (productId: number) =>
    api.post<APIResponse<{ id: number; product_id: number }>>('/favorites', {
      product_id: productId,
    }),

  removeFavorite: (productId: number) => api.delete<APIResponse<null>>(`/favorites/${productId}`),

  checkFavorite: (productId: number) =>
    api.get<APIResponse<{ is_favorite: boolean }>>(`/favorites/check/${productId}`),
};

export const browseHistoryService = {
  getHistory: (page = 1, pageSize = 20) =>
    api.get<APIResponse<BrowseHistoryListResponse>>('/browse-history', {
      params: { page, page_size: pageSize },
    }),

  addHistory: (productId: number) =>
    api.post<APIResponse<null>>('/browse-history', {
      product_id: productId,
    }),

  clearHistory: () => api.delete<APIResponse<null>>('/browse-history'),

  removeHistoryItem: (productId: number) =>
    api.delete<APIResponse<null>>(`/browse-history/${productId}`),
};
