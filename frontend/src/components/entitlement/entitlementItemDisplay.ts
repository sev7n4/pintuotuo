import type { EntitlementPackageItem } from '@/types/entitlementPackage';

/** 类型展示：concurrent 对用户展示为「智能路由选线」 */
export function skuTypeLabel(skuType: string): string {
  const m: Record<string, string> = {
    subscription: '订阅',
    token_pack: 'Token',
    trial: '试用',
    concurrent: '智能路由选线',
  };
  return m[skuType] || skuType || '—';
}

export function lineDisplayName(it: EntitlementPackageItem): string {
  const d = it.display_name?.trim();
  if (d) return d;
  return it.spu_name || it.sku_code || '—';
}

export function lineValueHint(it: EntitlementPackageItem): string {
  const note = it.value_note?.trim();
  if (note) return note;
  const unit = Number(it.retail_price || 0);
  const q = Number(it.default_quantity || 1);
  const sub = unit * q;
  return `单价 ¥${unit.toFixed(2)} × ${q} = ¥${sub.toFixed(2)}`;
}

/** 「套餐包含」四条要点（第 1 条来自组合内模型名；2～4 条据 SKU 类型与数量生成，缺类时给占位说明） */
export type PackageIncludeBullet = { key: string; text: string };

export function buildPackageIncludeBullets(
  items: EntitlementPackageItem[]
): PackageIncludeBullet[] {
  const list = items ?? [];
  if (list.length === 0) return [];

  const names = list.map((it) => lineDisplayName(it).trim()).filter((s) => s.length > 0);
  const inner = names.length ? names.join('+') : '—';

  const hasSubscription = list.some((it) => it.sku_type === 'subscription');
  let tokenTotal = 0;
  for (const it of list) {
    if (it.sku_type === 'token_pack') {
      const base = Number(it.token_amount ?? 0);
      const qty = it.default_quantity || 1;
      tokenTotal += base * qty;
    }
  }
  const hasConcurrent = list.some((it) => it.sku_type === 'concurrent');

  return [
    { key: 'combo', text: `「${inner}」` },
    {
      key: 'subscription',
      text: hasSubscription
        ? '模型调用权益 · 订阅 · 包月（续费规则以订单为准）'
        : '—（本组合暂无订阅类 SKU，若需此项请选购含订阅项的套餐）',
    },
    {
      key: 'token',
      text:
        tokenTotal > 0
          ? `赠送 Token · 约 ${tokenTotal.toLocaleString('zh-CN')}（按批次入账）`
          : '—（本组合暂无加油包/Token 包 SKU，或 Token 数为 0）',
    },
    {
      key: 'routing',
      text: hasConcurrent ? '智能线路选择/模型自由切换' : '—（本组合暂无智能线路/并发类 SKU）',
    },
  ];
}
