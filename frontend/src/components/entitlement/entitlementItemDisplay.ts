import type { EntitlementPackageItem } from '@/types/entitlementPackage';

export function subscriptionPeriodLabel(period: string | undefined): string {
  if (!period?.trim()) return '';
  const m: Record<string, string> = {
    monthly: '包月',
    quarterly: '包季',
    yearly: '包年',
  };
  return m[period] || period;
}

/** 类型展示：concurrent 对用户展示为「智能路由选线」，不写「并发」 */
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

/** 单行 SKU 的规格要点（与后台 skus 配置一致，供「套餐包含」与明细共用） */
export function packageItemSpecParts(it: EntitlementPackageItem): string[] {
  const parts: string[] = [];
  const qty = it.default_quantity || 1;

  switch (it.sku_type) {
    case 'subscription': {
      const p = subscriptionPeriodLabel(it.subscription_period);
      parts.push(p ? `订阅 · ${p}` : '订阅');
      break;
    }
    case 'token_pack': {
      const base = Number(it.token_amount ?? 0);
      const total = base * qty;
      if (total > 0) {
        parts.push(`约 ${total.toLocaleString('zh-CN')} Token（SKU 规格 × 数量）`);
      } else {
        parts.push('Token 包');
      }
      break;
    }
    case 'concurrent':
      parts.push('智能路由选线 · 支持自由切换');
      break;
    case 'trial':
      parts.push('试用');
      break;
    default:
      parts.push(skuTypeLabel(it.sku_type));
  }

  if (it.valid_days != null && it.valid_days > 0) {
    parts.push(`有效期：${it.valid_days} 天`);
  }

  return parts;
}

/**
 * 「A+B+C」模型组合（仅展示模型/线路名称，不列 SKU 编码；具体对应关系见下方分项明细）。
 */
export function packageModelComboSummary(items: EntitlementPackageItem[]): string {
  const list = items ?? [];
  const names = list.map((it) => lineDisplayName(it).trim()).filter((s) => s.length > 0);
  const inner = names.length > 0 ? names.join('+') : '—';
  return `「${inner}」模型`;
}

/** 多项时在组合行之下，提示下方为分项明细 */
export function packageIncludeHeadline(items: EntitlementPackageItem[]): string | null {
  const n = items?.length ?? 0;
  if (n <= 1) return null;
  return '分项规格如下（与上方「模型组合」逐项对应）：';
}
