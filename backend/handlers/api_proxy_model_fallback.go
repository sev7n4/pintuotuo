package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/pintuotuo/backend/services"
)

// apiProxyModelFallbackEnabled 默认开启；设置 API_PROXY_MODEL_FALLBACK=0|false|off 可关闭运行时模型链 fallback。
func apiProxyModelFallbackEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("API_PROXY_MODEL_FALLBACK")))
	return v != "0" && v != "false" && v != "off"
}

type proxyCatalogAttempt struct {
	provider string
	model    string
}

// buildProxyCatalogAttempts 首元素为当前请求的 provider/model，其后为 DB 配置的备用（若能解析目录主键）。
func buildProxyCatalogAttempts(ctx context.Context, db *sql.DB, req APIProxyRequest) []proxyCatalogAttempt {
	first := proxyCatalogAttempt{provider: strings.TrimSpace(req.Provider), model: strings.TrimSpace(req.Model)}
	out := []proxyCatalogAttempt{first}
	if !apiProxyModelFallbackEnabled() || db == nil {
		return out
	}
	canonKey := fmt.Sprintf("%s/%s", first.provider, first.model)
	canonical, err := services.CanonicalCatalogModelKey(ctx, db, canonKey)
	if err != nil {
		return out
	}
	chain, err := services.GetEnabledFallbackChain(ctx, db, canonical)
	if err != nil || len(chain) == 0 {
		return out
	}
	for _, fb := range chain {
		p, m, err := services.SplitCatalogModelKey(fb)
		if err != nil {
			continue
		}
		out = append(out, proxyCatalogAttempt{provider: p, model: m})
	}
	return out
}

// chatCompletionJSONIndicatesProxySuccess HTTP 200 且 OpenAI 风格 body 无 error 字段。
func chatCompletionJSONIndicatesProxySuccess(statusCode int, body []byte) bool {
	if statusCode != http.StatusOK {
		return false
	}
	var apiResp APIProxyResponse
	if json.Unmarshal(body, &apiResp) != nil {
		return false
	}
	return apiResp.Error == nil
}

func providerInfoFromUpstreamFailure(statusCode int, body []byte, headers http.Header, netErr error) services.ProviderErrorInfo {
	if netErr != nil {
		return services.MapProviderError(0, "", netErr.Error(), nil, netErr, "")
	}
	code, msg := services.ExtractProviderError(body)
	return services.MapProviderError(statusCode, code, msg, headers, nil, string(body))
}
