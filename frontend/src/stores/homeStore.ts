import { create } from 'zustand';
import { Product, Category, Banner, APIResponse, ScenarioCategoryItem } from '@/types';
import { productService } from '@/services/product';

interface HomeState {
  banners: Banner[];
  hotProducts: Product[];
  newProducts: Product[];
  categories: Category[];
  scenarioCategories: ScenarioCategoryItem[];
  isLoading: boolean;
  error: string | null;

  fetchHomeData: () => Promise<void>;
  fetchHotProducts: (limit?: number) => Promise<void>;
  fetchNewProducts: (limit?: number) => Promise<void>;
  fetchCategories: () => Promise<void>;
  clearError: () => void;
}

export const useHomeStore = create<HomeState>((set) => ({
  banners: [],
  hotProducts: [],
  newProducts: [],
  categories: [],
  scenarioCategories: [],
  isLoading: false,
  error: null,

  fetchHomeData: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await productService.getHomeData();
      const data = response.data;
      set({
        banners: data.banners || [],
        hotProducts: data.hot || [],
        newProducts: data.new || [],
        categories: data.categories || [],
        scenarioCategories: data.scenario_categories || [],
        isLoading: false,
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取首页数据失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchHotProducts: async (limit = 10) => {
    set({ isLoading: true, error: null });
    try {
      const response = await productService.getHotProducts(limit);
      const apiResponse = response.data as APIResponse<Product[]>;
      set({ hotProducts: apiResponse.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取热门商品失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchNewProducts: async (limit = 10) => {
    set({ isLoading: true, error: null });
    try {
      const response = await productService.getNewProducts(limit);
      const apiResponse = response.data as APIResponse<Product[]>;
      set({ newProducts: apiResponse.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取新品上架失败';
      set({ error: message, isLoading: false });
    }
  },

  fetchCategories: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await productService.getCategories();
      const apiResponse = response.data as APIResponse<Category[]>;
      set({ categories: apiResponse.data || [], isLoading: false });
    } catch (error) {
      const message = error instanceof Error ? error.message : '获取分类失败';
      set({ error: message, isLoading: false });
    }
  },

  clearError: () => set({ error: null }),
}));
