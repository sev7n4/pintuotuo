export interface ValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

export interface RouteStrategyItem {
  mode: string;
  weight?: number;
  fallback_mode?: string;
  conditions?: Record<string, any>;
}

export interface EndpointConfig {
  url: string;
  timeout?: number;
  retry_count?: number;
}

export function validateRouteStrategy(
  strategy: Record<string, RouteStrategyItem>
): ValidationResult {
  const errors: string[] = [];
  const warnings: string[] = [];

  if (!strategy || typeof strategy !== 'object') {
    errors.push('路由策略配置必须是一个对象');
    return { valid: false, errors, warnings };
  }

  const validModes = ['direct', 'litellm', 'proxy', 'auto'];
  const validUserTypes = [
    'domestic_users',
    'overseas_users',
    'enterprise_users',
    'default_mode',
  ];

  const userTypes = Object.keys(strategy);

  if (userTypes.length === 0) {
    warnings.push('路由策略配置为空，将使用默认路由模式');
    return { valid: true, errors, warnings };
  }

  for (const userType of userTypes) {
    if (!validUserTypes.includes(userType)) {
      errors.push(`无效的用户类型: ${userType}`);
      continue;
    }

    const item = strategy[userType];

    if (!item || typeof item !== 'object') {
      errors.push(`用户类型 ${userType} 的配置必须是一个对象`);
      continue;
    }

    if (!item.mode) {
      errors.push(`用户类型 ${userType} 缺少必填字段: mode`);
    } else if (!validModes.includes(item.mode)) {
      errors.push(
        `用户类型 ${userType} 的 mode 字段值无效: ${
          item.mode
        }，有效值为: ${validModes.join(', ')}`
      );
    }

    if (item.weight !== undefined) {
      if (typeof item.weight !== 'number' || item.weight < 0 || item.weight > 100) {
        errors.push(`用户类型 ${userType} 的 weight 字段必须在 0-100 之间`);
      }
    }

    if (item.fallback_mode !== undefined) {
      if (!validModes.includes(item.fallback_mode)) {
        errors.push(
          `用户类型 ${userType} 的 fallback_mode 字段值无效: ${item.fallback_mode}`
        );
      }
    }

    if (item.mode === 'auto' && !item.conditions) {
      warnings.push(`用户类型 ${userType} 使用 auto 模式但未配置 conditions 字段`);
    }
  }

  if (!strategy.default_mode) {
    warnings.push('建议配置 default_mode 作为默认路由策略');
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

export function validateEndpoints(
  endpoints: Record<string, EndpointConfig>
): ValidationResult {
  const errors: string[] = [];
  const warnings: string[] = [];

  if (!endpoints || typeof endpoints !== 'object') {
    errors.push('端点配置必须是一个对象');
    return { valid: false, errors, warnings };
  }

  const validModes = ['direct', 'litellm', 'proxy'];
  const modes = Object.keys(endpoints);

  if (modes.length === 0) {
    warnings.push('端点配置为空，将使用默认端点');
    return { valid: true, errors, warnings };
  }

  const urlPattern = /^https?:\/\/.+/;

  for (const mode of modes) {
    if (!validModes.includes(mode)) {
      errors.push(`无效的网关模式: ${mode}`);
      continue;
    }

    const endpoint = endpoints[mode];

    if (!endpoint || typeof endpoint !== 'object') {
      errors.push(`网关模式 ${mode} 的配置必须是一个对象`);
      continue;
    }

    if (!endpoint.url) {
      errors.push(`网关模式 ${mode} 缺少必填字段: url`);
    } else if (typeof endpoint.url !== 'string') {
      errors.push(`网关模式 ${mode} 的 url 字段必须是字符串`);
    } else if (!urlPattern.test(endpoint.url)) {
      errors.push(`网关模式 ${mode} 的 url 字段格式无效，必须以 http:// 或 https:// 开头`);
    }

    if (endpoint.timeout !== undefined) {
      if (
        typeof endpoint.timeout !== 'number' ||
        endpoint.timeout < 1000 ||
        endpoint.timeout > 300000
      ) {
        errors.push(`网关模式 ${mode} 的 timeout 字段必须在 1000-300000 毫秒之间`);
      }
    }

    if (endpoint.retry_count !== undefined) {
      if (
        typeof endpoint.retry_count !== 'number' ||
        endpoint.retry_count < 0 ||
        endpoint.retry_count > 10
      ) {
        errors.push(`网关模式 ${mode} 的 retry_count 字段必须在 0-10 之间`);
      }
    }
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

export function validateProviderConfig(config: {
  route_strategy?: Record<string, RouteStrategyItem>;
  endpoints?: Record<string, EndpointConfig>;
}): ValidationResult {
  const errors: string[] = [];
  const warnings: string[] = [];

  if (config.route_strategy) {
    const strategyResult = validateRouteStrategy(config.route_strategy);
    errors.push(...strategyResult.errors);
    warnings.push(...strategyResult.warnings);
  }

  if (config.endpoints) {
    const endpointsResult = validateEndpoints(config.endpoints);
    errors.push(...endpointsResult.errors);
    warnings.push(...endpointsResult.warnings);
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  };
}
