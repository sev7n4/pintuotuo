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
  health_error_message: string;
  health_error_category: string;
  health_error_code: string;
  last_health_check_at: string | null;
  verification_result: string;
  verification_message: string;
  models_supported: string[];
  verified_at: string | null;
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

export interface VerificationResult {
  id: number;
  api_key_id: number;
  verification_type: string;
  status: 'success' | 'failed' | 'in_progress' | 'pending';
  connection_test: boolean;
  connection_latency_ms: number;
  models_found: string[];
  models_count: number;
  pricing_verified: boolean;
  pricing_info?: Record<string, unknown>;
  error_code?: string;
  error_message?: string;
  started_at: string;
  completed_at?: string;
  retry_count: number;
}

export interface VerificationDetailsResponse {
  api_key: {
    id: number;
    merchant_id: number;
    provider: string;
    verification_result: string;
    verified_at: string | null;
    models_supported: string[];
    verification_message: string;
  };
  history: VerificationResult[];
}

const getVerificationDetails = async (
  id: number
): Promise<{ data: VerificationDetailsResponse }> => {
  return api.get(`/admin/byok-routing/${id}/verification`);
};

export const adminByokRoutingService = {
  getByokRoutingList,
  updateRouteConfig,
  triggerProbe,
  triggerLightVerify,
  triggerDeepVerify,
  getVerificationDetails,
};
