import { AxiosError } from 'axios';

/** 解析后端 { code, message, error } 响应，用于下单等业务错误提示 */
export function getApiErrorMessage(err: unknown, fallback?: string): string {
  if (err instanceof AxiosError && err.response?.data) {
    const d = err.response.data as { message?: string; error?: string; code?: string };
    if (typeof d.message === 'string' && d.message.trim()) {
      return mapOrderErrorCode(d.code, d.message);
    }
    if (typeof d.error === 'string' && d.error.trim()) {
      return mapOrderErrorCode(d.code, d.error);
    }
  }
  if (err instanceof Error) return err.message;
  return fallback?.trim() || '请求失败';
}

function mapOrderErrorCode(code: string | undefined, fallback: string): string {
  const map: Record<string, string> = {
    ORDER_LINE_SKU_UNAVAILABLE: '包含不可售或已下架的商品，请刷新页面后重试',
    PRODUCT_NOT_FOUND: '商品已下架或不存在，请刷新后重试',
    INSUFFICIENT_STOCK: '库存不足，请稍后再试或联系客服',
    ENTITLEMENT_SKU_NOT_SELLABLE: '权益包内含有未上架 SKU，请等待运营更新配置',
    ENTITLEMENT_SPU_NOT_SELLABLE: '权益包内含有 SPU 未在售的 SKU，请等待运营更新配置',
    ENTITLEMENT_SKU_INSUFFICIENT_STOCK: '权益包内 SKU 库存不足，请等待补货或联系运营',
  };
  if (code && map[code]) return map[code];
  return fallback;
}
