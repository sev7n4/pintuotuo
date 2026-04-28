import api from './api';

export interface BYOKRoutingItem {
  id: number;
  merchant_id: number;
  company_name: string;
  byok_type: 'official' | 'reseller' | 'self_hosted';
  provider: string;
  name: string;
  region: 'domestic' | 'overseas';
  route_mode: 'auto' | 'direct' | 'litellm' | 'proxy';
  endpoint_url: string;
  fallback_endpoint_url: string;
  route_config: Record<string, unknown>;
  health_status: string;
  verification_result: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface BYOKRoutingListResponse {
  data: BYOKRoutingItem[];
  total: number;
}

export interface UpdateRouteConfigRequest {
  route_mode?: string;
  endpoint_url?: string;
  fallback_endpoint_url?: string;
  route_config?: Record<string, unknown>;
}

const getByokRoutingList = async (params?: {
  merchant_id?: number;
  byok_type?: string;
  provider?: string;
  region?: string;
  route_mode?: string;
  health_status?: string;
}): Promise<{ data: BYOKRoutingListResponse }> => {
  const queryParams = new URLSearchParams();
  if (params?.merchant_id) queryParams.append('merchant_id', String(params.merchant_id));
  if (params?.byok_type) queryParams.append('byok_type', params.byok_type);
  if (params?.provider) queryParams.append('provider', params.provider);
  if (params?.region) queryParams.append('region', params.region);
  if (params?.route_mode) queryParams.append('route_mode', params.route_mode);
  if (params?.health_status) queryParams.append('health_status', params.health_status);
  return api.get(`/admin/byok-routing?${queryParams.toString()}`);
};

const updateRouteConfig = async (
  id: number,
  data: UpdateRouteConfigRequest
): Promise<{ data: { message: string; api_key_id: number } }> => {
  return api.put(`/admin/byok-routing/${id}/route-config`, data);
};

const triggerProbe = async (
  id: number
): Promise<{ data: { message: string; api_key_id: number } }> => {
  return api.post(`/admin/byok-routing/${id}/probe`);
};

const triggerLightVerify = async (
  id: number
): Promise<{ data: { message: string; api_key_id: number; verification_type: string } }> => {
  return api.post(`/admin/byok-routing/${id}/light-verify`);
};

const triggerDeepVerify = async (
  id: number
): Promise<{ data: { message: string; api_key_id: number; verification_type: string } }> => {
  return api.post(`/admin/byok-routing/${id}/deep-verify`);
};

export const adminByokRoutingService = {
  getByokRoutingList,
  updateRouteConfig,
  triggerProbe,
  triggerLightVerify,
  triggerDeepVerify,
};
