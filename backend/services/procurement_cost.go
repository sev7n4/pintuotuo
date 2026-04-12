package services

// ProcurementCostCNY 按商户 SKU 配置的元/1K 输入输出单价计算采购成本（人民币元）。
// 与对用户扣费的 CostFromPer1KRates 公式一致，数据源为 merchant_skus 成本字段而非 SPU 参考价。
func ProcurementCostCNY(costInputPer1K, costOutputPer1K float64, inputTokens, outputTokens int) float64 {
	return CostFromPer1KRates(costInputPer1K, costOutputPer1K, inputTokens, outputTokens)
}
