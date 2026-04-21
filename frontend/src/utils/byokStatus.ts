import type { MerchantAPIKey } from '@/types';

/** 与后端 strict 权益白名单 buildAllowlistEdgesForSKU 条件对齐 */
export function isStrictEntitlementEligible(k: MerchantAPIKey): boolean {
  const verifiedLine =
    !!(k.verified_at && String(k.verified_at).trim() !== '') ||
    k.verification_result === 'verified' ||
    k.verification_result === 'success';
  const h = (k.health_status || 'unknown').toLowerCase();
  const healthOk = h === 'healthy' || h === 'degraded';
  return verifiedLine && healthOk;
}

/** 启用中的 Key 未满足 strict 条件时与「Strict 权益 / 未满足」一致，需立即关注 */
export function keyNeedsAttentionActive(k: MerchantAPIKey): boolean {
  if (k.status !== 'active') return false;
  return !isStrictEntitlementEligible(k);
}

export type ByokMerchantLevel = 'none' | 'gray' | 'yellow' | 'green';

/** 与后端 services.AggregateMerchantBYOK 规则一致（仅 active 参与路由语义） */
export function aggregateMerchantBYOK(keys: MerchantAPIKey[]): {
  level: ByokMerchantLevel;
  hasRoutable: boolean;
  needAttentionActive: number;
  activeCount: number;
  totalCount: number;
} {
  const totalCount = keys.length;
  const active = keys.filter((k) => k.status === 'active');
  const activeCount = active.length;
  if (totalCount === 0) {
    return {
      level: 'none',
      hasRoutable: false,
      needAttentionActive: 0,
      activeCount: 0,
      totalCount: 0,
    };
  }
  if (activeCount === 0) {
    return {
      level: 'gray',
      hasRoutable: false,
      needAttentionActive: 0,
      activeCount: 0,
      totalCount,
    };
  }

  let hasRoutable = false;
  let needAttentionActive = 0;
  for (const k of active) {
    if (isStrictEntitlementEligible(k)) hasRoutable = true;
    if (keyNeedsAttentionActive(k)) needAttentionActive++;
  }
  if (hasRoutable) {
    return { level: 'green', hasRoutable: true, needAttentionActive, activeCount, totalCount };
  }
  for (const k of active) {
    const h = (k.health_status || 'unknown').toLowerCase();
    const vr = (k.verification_result || '').toLowerCase();
    if (h === 'unhealthy' || vr === 'failed') {
      return {
        level: 'yellow',
        hasRoutable: false,
        needAttentionActive,
        activeCount,
        totalCount,
      };
    }
  }
  return {
    level: 'gray',
    hasRoutable: false,
    needAttentionActive,
    activeCount,
    totalCount,
  };
}
