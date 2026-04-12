/**
 * 用户 Token 余额 / 消费扣减的展示格式化。
 * 单位：模型 Token（input_tokens + output_tokens）
 * 
 * 计费单位口径统一后：
 * - 用户侧：使用模型 Token 数量（token_usage = input + output）
 * - 商户侧：使用人民币金额（total_sales_cny）
 */
export function formatLedgerUnits(value: number | undefined | null): string {
  const n = Number(value ?? 0);
  return n.toLocaleString(undefined, { maximumFractionDigits: 0 });
}

export const ledgerUnitColumnTitle = 'Token';
export const ledgerUnitShort = 'Token';
