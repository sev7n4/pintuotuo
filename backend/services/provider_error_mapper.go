package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	errorCategoryAuthInvalidKey       = "AUTH_INVALID_KEY"
	errorCategoryAuthPermissionDenied = "AUTH_PERMISSION_DENIED"
	errorCategoryQuotaInsufficient    = "QUOTA_INSUFFICIENT"
	errorCategoryRateLimited          = "RATE_LIMITED"
	errorCategoryModelNotFound        = "MODEL_NOT_FOUND"
	errorCategoryContextTooLong       = "CONTEXT_WINDOW_EXCEEDED"
	errorCategoryServiceUnavailable   = "SERVICE_UNAVAILABLE"
	errorCategoryNetworkTimeout       = "NETWORK_TIMEOUT"
	errorCategoryNetworkDNS           = "NETWORK_DNS"
	errorCategoryUpstreamBadRequest   = "UPSTREAM_BAD_REQUEST"
	errorCategoryUnknown              = "UNKNOWN"
)

type ProviderErrorInfo struct {
	Category          string
	ProviderCode      string
	ProviderMessage   string
	HTTPStatusCode    int
	ProviderRequestID string
	Retryable         bool
	RawErrorExcerpt   string
}

func MapProviderError(statusCode int, providerCode, providerMessage string, headers http.Header, netErr error, rawBody string) ProviderErrorInfo {
	code := strings.ToLower(strings.TrimSpace(providerCode))
	msg := strings.ToLower(strings.TrimSpace(providerMessage))
	netMsg := ""
	if netErr != nil {
		netMsg = strings.ToLower(strings.TrimSpace(netErr.Error()))
	}

	reqID := firstNonEmpty(
		headerValue(headers, "x-request-id"),
		headerValue(headers, "x-amzn-requestid"),
		headerValue(headers, "x-amz-request-id"),
		headerValue(headers, "request-id"),
	)

	info := ProviderErrorInfo{
		Category:          errorCategoryUnknown,
		ProviderCode:      strings.TrimSpace(providerCode),
		ProviderMessage:   strings.TrimSpace(providerMessage),
		HTTPStatusCode:    statusCode,
		ProviderRequestID: strings.TrimSpace(reqID),
		RawErrorExcerpt:   truncateErrorExcerpt(rawBody, 512),
	}

	if netErr != nil {
		switch {
		case strings.Contains(netMsg, "timeout"), strings.Contains(netMsg, "deadline exceeded"):
			info.Category = errorCategoryNetworkTimeout
			info.Retryable = true
		case strings.Contains(netMsg, "no such host"), strings.Contains(netMsg, "lookup"):
			info.Category = errorCategoryNetworkDNS
			info.Retryable = true
		default:
			info.Category = errorCategoryServiceUnavailable
			info.Retryable = true
		}
		return info
	}

	switch {
	case statusCode == http.StatusUnauthorized || code == "invalid_api_key" || strings.Contains(code, "invalid_key") || strings.Contains(msg, "invalid api key"):
		info.Category = errorCategoryAuthInvalidKey
	case statusCode == http.StatusForbidden || strings.Contains(code, "permission") || strings.Contains(msg, "permission"):
		info.Category = errorCategoryAuthPermissionDenied
	case statusCode == http.StatusTooManyRequests || strings.Contains(code, "rate") || strings.Contains(msg, "rate limit"):
		info.Category = errorCategoryRateLimited
		info.Retryable = true
	case statusCode == http.StatusPaymentRequired || strings.Contains(code, "quota") || strings.Contains(code, "insufficient") || strings.Contains(msg, "quota") || strings.Contains(msg, "insufficient"):
		info.Category = errorCategoryQuotaInsufficient
	case statusCode == http.StatusNotFound || strings.Contains(code, "model_not_found") || strings.Contains(msg, "model not found"):
		info.Category = errorCategoryModelNotFound
	case statusCode == http.StatusBadRequest && (strings.Contains(code, "context") || strings.Contains(msg, "context") || strings.Contains(msg, "max tokens")):
		info.Category = errorCategoryContextTooLong
	case statusCode == http.StatusBadRequest:
		info.Category = errorCategoryUpstreamBadRequest
	case statusCode >= http.StatusInternalServerError:
		info.Category = errorCategoryServiceUnavailable
		info.Retryable = true
	default:
		info.Category = errorCategoryUnknown
	}

	return info
}

// SuggestModelFallbackAfterFailure 判断上游失败后是否应尝试配置中的下一备用模型（鉴权/配额等不应切换模型）。
func SuggestModelFallbackAfterFailure(info ProviderErrorInfo) bool {
	switch info.Category {
	case errorCategoryAuthInvalidKey, errorCategoryAuthPermissionDenied, errorCategoryQuotaInsufficient:
		return false
	case errorCategoryContextTooLong, errorCategoryUpstreamBadRequest:
		return false
	}
	if info.Retryable {
		return true
	}
	switch info.Category {
	case errorCategoryModelNotFound, errorCategoryRateLimited, errorCategoryServiceUnavailable,
		errorCategoryNetworkTimeout, errorCategoryNetworkDNS:
		return true
	case errorCategoryUnknown:
		return info.HTTPStatusCode == http.StatusTooManyRequests || info.HTTPStatusCode >= http.StatusInternalServerError
	default:
		return false
	}
}

func headerValue(h http.Header, key string) string {
	if h == nil {
		return ""
	}
	return strings.TrimSpace(h.Get(key))
}

func truncateErrorExcerpt(raw string, max int) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// HTTPUpstreamRetryable 基于结构化错误映射判断代理层是否应对该 HTTP 响应重试（与验证/探测语义一致）。
func HTTPUpstreamRetryable(statusCode int, body []byte, headers http.Header) bool {
	code, msg := ExtractProviderError(body)
	info := MapProviderError(statusCode, code, msg, headers, nil, string(body))
	return info.Retryable
}

// ExtractProviderError 解析常见 OpenAI 风格 error JSON（与验证、探测、健康检查共用）。
func ExtractProviderError(body []byte) (string, string) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "", ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", trimmed
	}

	if errNode, ok := payload["error"].(map[string]interface{}); ok {
		return getString(errNode, "code"), firstNonEmpty(
			getString(errNode, "message"),
			getString(errNode, "msg"),
		)
	}
	return getString(payload, "code"), firstNonEmpty(
		getString(payload, "message"),
		getString(payload, "msg"),
		trimmed,
	)
}

func getString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", s)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
