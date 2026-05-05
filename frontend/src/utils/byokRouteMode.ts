export type RouteMode = 'direct' | 'litellm' | 'proxy' | 'auto';

export const ROUTE_MODE_LABELS: Record<RouteMode, string> = {
  direct: '直连',
  litellm: 'LiteLLM',
  proxy: '代理',
  auto: '自动',
};

export const ROUTE_MODE_COLORS: Record<RouteMode, string> = {
  direct: '#52c41a',
  litellm: '#1890ff',
  proxy: '#722ed1',
  auto: '#faad14',
};

export const getRouteModeLabel = (mode?: string): string => {
  if (!mode) return '未知';
  return ROUTE_MODE_LABELS[mode as RouteMode] || '未知';
};

export const getRouteModeColor = (mode?: string): string => {
  if (!mode) return '#d9d9d9';
  return ROUTE_MODE_COLORS[mode as RouteMode] || '#d9d9d9';
};

export type ErrorCategory =
  | 'AUTHENTICATION_ERROR'
  | 'RATE_LIMIT_ERROR'
  | 'NETWORK_ERROR'
  | 'PROVIDER_ERROR'
  | 'QUOTA_EXCEEDED'
  | 'INVALID_REQUEST'
  | 'MODEL_NOT_FOUND'
  | 'INSUFFICIENT_QUOTA'
  | 'UNKNOWN_ERROR';

export const ERROR_CATEGORY_LABELS: Record<ErrorCategory, string> = {
  AUTHENTICATION_ERROR: '认证错误',
  RATE_LIMIT_ERROR: '速率限制',
  NETWORK_ERROR: '网络错误',
  PROVIDER_ERROR: '服务商错误',
  QUOTA_EXCEEDED: '配额超限',
  INVALID_REQUEST: '无效请求',
  MODEL_NOT_FOUND: '模型不存在',
  INSUFFICIENT_QUOTA: '余额不足',
  UNKNOWN_ERROR: '未知错误',
};

export const ERROR_CATEGORY_COLORS: Record<ErrorCategory, string> = {
  AUTHENTICATION_ERROR: '#f5222d',
  RATE_LIMIT_ERROR: '#fa8c16',
  NETWORK_ERROR: '#722ed1',
  PROVIDER_ERROR: '#eb2f96',
  QUOTA_EXCEEDED: '#f5222d',
  INVALID_REQUEST: '#faad14',
  MODEL_NOT_FOUND: '#faad14',
  INSUFFICIENT_QUOTA: '#f5222d',
  UNKNOWN_ERROR: '#8c8c8c',
};

export const getErrorCategoryLabel = (category?: string): string => {
  if (!category) return '未知错误';
  return ERROR_CATEGORY_LABELS[category as ErrorCategory] || '未知错误';
};

export const getErrorCategoryColor = (category?: string): string => {
  if (!category) return '#8c8c8c';
  return ERROR_CATEGORY_COLORS[category as ErrorCategory] || '#8c8c8c';
};

export const getErrorCategoryByCode = (errorCode?: string): ErrorCategory => {
  if (!errorCode) return 'UNKNOWN_ERROR';

  if (errorCode.includes('AUTH') || errorCode.includes('401') || errorCode.includes('403')) {
    return 'AUTHENTICATION_ERROR';
  }
  if (errorCode.includes('RATE') || errorCode.includes('429')) {
    return 'RATE_LIMIT_ERROR';
  }
  if (
    errorCode.includes('NETWORK') ||
    errorCode.includes('TIMEOUT') ||
    errorCode.includes('ECONNREFUSED')
  ) {
    return 'NETWORK_ERROR';
  }
  if (errorCode.includes('QUOTA') || errorCode.includes('INSUFFICIENT')) {
    return 'QUOTA_EXCEEDED';
  }
  if (errorCode.includes('MODEL') || errorCode.includes('NOT_FOUND')) {
    return 'MODEL_NOT_FOUND';
  }
  if (errorCode.includes('INVALID') || errorCode.includes('400')) {
    return 'INVALID_REQUEST';
  }

  return 'UNKNOWN_ERROR';
};
