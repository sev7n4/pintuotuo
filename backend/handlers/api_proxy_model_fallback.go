package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/pintuotuo/backend/logger"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
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

// resolveProxyAttemptRuntime 统一流式/非流式在 fallback 尝试项上的 key+provider 解析逻辑。
// skip=true 表示该尝试项应跳过（例如无可用 key、解密失败、provider 不存在）；fatalErr 需终止请求并返回 5xx。
func resolveProxyAttemptRuntime(
	ctx context.Context,
	db *sql.DB,
	userID int,
	merchantID int,
	req APIProxyRequest,
	att proxyCatalogAttempt,
	baseKey models.MerchantAPIKey,
	baseDecryptedKey string,
	baseProviderCfg providerRuntimeConfig,
	entCtx *services.EntitlementRoutingContext,
	requestID string,
) (pk models.MerchantAPIKey, decryptedKey string, pcfg providerRuntimeConfig, skip bool, fatalErr error) {
	if att.provider == req.Provider {
		return baseKey, baseDecryptedKey, baseProviderCfg, false, nil
	}

	modReq := req
	modReq.Provider = att.provider
	modReq.Model = att.model
	if selErr := selectAPIKeyForRequest(db, userID, merchantID, modReq, &pk, entCtx); selErr != nil {
		logger.LogWarn(ctx, "api_proxy", "model fallback key selection skipped", map[string]interface{}{
			"request_id": requestID, "provider": att.provider, "model": att.model, "error": selErr.Error(),
		})
		return models.MerchantAPIKey{}, "", providerRuntimeConfig{}, true, nil
	}

	var decErr error
	decryptedKey, decErr = utils.Decrypt(pk.APIKeyEncrypted)
	if decErr != nil {
		return models.MerchantAPIKey{}, "", providerRuntimeConfig{}, true, nil
	}

	pcfg, fatalErr = getProviderRuntimeConfig(db, att.provider)
	if fatalErr != nil {
		if errors.Is(fatalErr, sql.ErrNoRows) {
			return models.MerchantAPIKey{}, "", providerRuntimeConfig{}, true, nil
		}
		return models.MerchantAPIKey{}, "", providerRuntimeConfig{}, false, fatalErr
	}
	return pk, decryptedKey, pcfg, false, nil
}
