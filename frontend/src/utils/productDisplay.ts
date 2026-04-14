import type { Product } from '@/types';
import type { SKUWithSPU } from '@/types/sku';

/** 首页/卖场卡片一行卖点 */
export function getProductCardSubtitle(p: Product): string {
  const parts: string[] = [];
  if (p.token_count && p.token_count > 0) {
    const w =
      p.token_count >= 10000 ? `${(p.token_count / 10000).toFixed(0)}万` : String(p.token_count);
    parts.push(`${w} Token`);
  }
  if (p.validity_period) parts.push(p.validity_period);
  else if (p.category) parts.push(p.category);
  if (parts.length === 0 && p.models?.length) parts.push(p.models[0]);
  return parts.slice(0, 2).join(' · ') || '模型算力套餐';
}

export function getSkuCardSubtitle(s: SKUWithSPU): string {
  if (s.sku_type === 'token_pack' && s.token_amount) {
    return `${s.token_amount.toLocaleString()} tokens · ${s.model_name || ''}`.trim();
  }
  if (s.sku_type === 'subscription') {
    const m: Record<string, string> = { monthly: '月度', quarterly: '季度', yearly: '年度' };
    const gift =
      s.token_amount && s.token_amount > 0 ? ` · 赠送${s.token_amount.toLocaleString()}Token` : '';
    return `${m[s.subscription_period || 'monthly'] || '订阅'}${gift} · ${s.model_name || ''}`.trim();
  }
  if (s.sku_type === 'concurrent' && s.concurrent_requests) {
    return `${s.concurrent_requests} 并发`;
  }
  return s.spu_name || s.sku_code;
}

export const RECENT_SEARCH_KEY = 'pintuotuo_catalog_recent_searches';
export const MAX_RECENT_SEARCHES = 8;

export function readRecentSearches(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_SEARCH_KEY);
    if (!raw) return [];
    const arr = JSON.parse(raw) as unknown;
    return Array.isArray(arr) ? arr.filter((x): x is string => typeof x === 'string') : [];
  } catch {
    return [];
  }
}

export function pushRecentSearch(q: string): void {
  const t = q.trim();
  if (!t) return;
  const prev = readRecentSearches().filter((x) => x !== t);
  prev.unshift(t);
  localStorage.setItem(RECENT_SEARCH_KEY, JSON.stringify(prev.slice(0, MAX_RECENT_SEARCHES)));
}
