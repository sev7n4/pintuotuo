package services

import (
	"database/sql"
	"strings"
)

// AnthropicSiblingProviderSuffix 为写死的产品后缀：主厂商 code + 此后缀 = Anthropic 出站用 model_providers.code。
const AnthropicSiblingProviderSuffix = "_anthropic"

// AnthropicSiblingProviderCode 返回主厂商对应的 Anthropic 影子厂商 code（小写 trim + 后缀）。
func AnthropicSiblingProviderCode(primaryProvider string) string {
	p := strings.ToLower(strings.TrimSpace(primaryProvider))
	if p == "" {
		return ""
	}
	return p + AnthropicSiblingProviderSuffix
}

// ModelProviderCodeExistsActive 判断 model_providers 是否存在且 status=active。
func ModelProviderCodeExistsActive(db *sql.DB, code string) bool {
	if db == nil {
		return false
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	var one int
	err := db.QueryRow(
		`SELECT 1 FROM model_providers WHERE lower(trim(code)) = lower(trim($1)) AND status = 'active' LIMIT 1`,
		code,
	).Scan(&one)
	return err == nil
}
