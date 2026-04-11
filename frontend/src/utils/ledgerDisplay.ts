/**
 * 与用户 `tokens.balance` / 代理扣费同口径的「内部可消费单位」（平台内部记账单位）。
 * `api_usage_logs.cost`、账单统计里的 `total_cost`、商户配额 `quota_used` 均为此口径，不是人民币标价。
 */
export function formatLedgerUnits(value: number | undefined | null): string {
  const n = Number(value ?? 0);
  return n.toLocaleString(undefined, { maximumFractionDigits: 6 });
}

export const ledgerUnitColumnTitle = '扣减（Token）';
export const ledgerUnitShort = 'Token';
