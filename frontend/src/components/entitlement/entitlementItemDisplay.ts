import type { EntitlementPackageItem } from '@/types/entitlementPackage';

export function skuTypeLabel(skuType: string): string {
  const m: Record<string, string> = {
    subscription: '订阅',
    token_pack: 'Token',
    trial: '试用',
    concurrent: '并发',
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
