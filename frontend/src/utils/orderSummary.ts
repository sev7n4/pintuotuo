import type { Order } from '@/types';

/** 列表/摘要：替代「产品 ID」空列，展示用户可读的商品摘要 */
export function getOrderProductSummary(order: Order): string {
  if (order.product_id != null && Number(order.product_id) > 0) {
    return `商品 #${order.product_id}`;
  }
  const items = order.items || [];
  if (items.length === 0) return '—';
  const first = items[0];
  const name = (first.spu_name && first.spu_name.trim()) || `规格 #${first.sku_id}`;
  if (items.length === 1) return name;
  return `${name} 等 ${items.length} 项`;
}

export function canReorderFromOrder(order: Order): boolean {
  const n = order.items?.length ?? 0;
  if (n === 0) return false;
  return ['completed', 'paid', 'processing'].includes(order.status);
}

function skuTypeLabel(t: string | undefined): string {
  if (!t) return '—';
  const m: Record<string, string> = {
    subscription: '订阅',
    token_pack: 'Token',
    trial: '试用',
    concurrent: '并发',
  };
  return m[t] || t;
}

export function orderItemLineTitle(item: {
  spu_name?: string;
  sku_code?: string;
  sku_id: number;
  sku_type?: string;
}): string {
  const title = item.spu_name?.trim() || `规格 #${item.sku_id}`;
  const code = item.sku_code ? ` · ${item.sku_code}` : '';
  return `${title}${code}（${skuTypeLabel(item.sku_type)}）`;
}
