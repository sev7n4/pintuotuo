package services

import (
	"database/sql"
	"strings"
)

// AnthropicSiblingProviderSuffix 为写死的产品后缀：主厂商 code + 此后缀 = Anthropic 出站用 model_providers.code。
const AnthropicSiblingProviderSuffix = "_anthropic"

// PrimaryProviderFromAnthropicSibling 从影子厂商 code 还原主厂商 code（如 alibaba_anthropic → alibaba）。
func PrimaryProviderFromAnthropicSibling(provider string) string {
	p := strings.ToLower(strings.TrimSpace(provider))
	if p == "" || !strings.HasSuffix(p, AnthropicSiblingProviderSuffix) {
		return ""
	}
	return strings.TrimSuffix(p, AnthropicSiblingProviderSuffix)
}

// AnthropicSiblingProviderCode 返回主厂商对应的 Anthropic 影子厂商 code（小写 trim + 后缀）。
func AnthropicSiblingProviderCode(primaryProvider string) string {
	p := strings.ToLower(strings.TrimSpace(primaryProvider))
	if p == "" {
		return ""
	}
	return p + AnthropicSiblingProviderSuffix
}

// ProviderUsesAnthropicHTTP 判断出站 HTTP 应使用 Anthropic Messages 鉴权（x-api-key + anthropic-version）。
// apiFormat 来自 model_providers.api_format；provider 为 merchant_api_keys.provider。
func ProviderUsesAnthropicHTTP(provider, apiFormat string) bool {
	if strings.EqualFold(strings.TrimSpace(apiFormat), modelProviderAnthropic) {
		return true
	}
	p := strings.ToLower(strings.TrimSpace(provider))
	if p == modelProviderAnthropic {
		return true
	}
	return strings.HasSuffix(p, AnthropicSiblingProviderSuffix)
}

// SiblingOpenAIBaseFromDB 返回影子厂商对应主厂商的 OpenAI 兼容 api_base_url（用于 Anthropic 线路拉取模型列表）。
func SiblingOpenAIBaseFromDB(db *sql.DB, anthropicProvider string) string {
	primary := PrimaryProviderFromAnthropicSibling(anthropicProvider)
	if primary == "" || db == nil {
		return ""
	}
	var baseURL, apiFormat string
	err := db.QueryRow(
		`SELECT COALESCE(NULLIF(TRIM(api_base_url), ''), ''), COALESCE(NULLIF(TRIM(api_format), ''), $1)
		 FROM model_providers WHERE lower(trim(code)) = lower(trim($2)) AND status = 'active'`,
		modelProviderOpenAI, primary,
	).Scan(&baseURL, &apiFormat)
	if err != nil || !strings.EqualFold(strings.TrimSpace(apiFormat), modelProviderOpenAI) {
		return ""
	}
	return strings.TrimRight(strings.TrimSpace(baseURL), "/")
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
