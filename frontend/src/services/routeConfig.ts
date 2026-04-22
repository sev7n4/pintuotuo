import api from './api';

export interface ProviderRouteConfig {
  id: number;
  code: string;
  name: string;
  provider_region: string;
  route_strategy: Record<string, any>;
  endpoints: Record<string, any>;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface MerchantRouteConfig {
  id: number;
  name: string;
  merchant_type: string;
  region: string;
  route_preference: Record<string, any>;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface RouteTestResult {
  mode: string;
  endpoint: string;
  fallback_mode: string;
  fallback_endpoint: string;
  reason: string;
  provider_config: Record<string, any>;
  merchant_config: Record<string, any>;
}

export const routeConfigService = {
  getProviderRouteConfigs: async (params?: { region?: string; status?: string }) => {
    const response = await api.get<{ data: ProviderRouteConfig[] }>(
      '/admin/route-configs/providers',
      { params }
    );
    return response.data.data || [];
  },

  getProviderRouteConfig: async (code: string) => {
    const response = await api.get<{ data: ProviderRouteConfig }>(
      `/admin/route-configs/providers/${code}`
    );
    return response.data.data;
  },

  updateProviderRouteConfig: async (code: string, data: Partial<ProviderRouteConfig>) => {
    const response = await api.put(`/admin/route-configs/providers/${code}`, data);
    return response.data;
  },

  getMerchantRouteConfigs: async (params?: { type?: string; region?: string; status?: string }) => {
    const response = await api.get<{ data: MerchantRouteConfig[] }>(
      '/admin/route-configs/merchants',
      { params }
    );
    return response.data.data || [];
  },

  updateMerchantRouteConfig: async (id: number, data: Partial<MerchantRouteConfig>) => {
    const response = await api.put(`/admin/route-configs/merchants/${id}`, data);
    return response.data;
  },

  testRouteDecision: async (providerCode: string, merchantId: number) => {
    const response = await api.post<{ data: RouteTestResult }>('/admin/route-configs/test', {
      provider_code: providerCode,
      merchant_id: merchantId,
    });
    return response.data.data;
  },
};

export default routeConfigService;
