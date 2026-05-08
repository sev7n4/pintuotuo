package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
)

type LitellmTemplateEntry struct {
	Template  string
	APIKeyEnv string
	// APIBase：model_providers.litellm_gateway_api_base（网关专用覆盖，通常为空）
	APIBase string
	// ProviderAPIBaseURL：model_providers.api_base_url（厂商 OpenAI 兼容根地址；与 api_key_validator 探测一致）
	ProviderAPIBaseURL string
}

var (
	litellmCacheMu       sync.RWMutex
	litellmTemplateCache map[string]LitellmTemplateEntry
	litellmCacheLoaded   bool
	litellmCacheLoadOnce sync.Once
)

func LoadLitellmTemplateCache() {
	db := config.GetDB()
	if db == nil {
		return
	}

	rows, err := db.Query(`
		SELECT code,
		       COALESCE(NULLIF(TRIM(litellm_model_template), ''), ''),
		       COALESCE(NULLIF(TRIM(litellm_gateway_api_key_env), ''), ''),
		       COALESCE(NULLIF(TRIM(litellm_gateway_api_base), ''), ''),
		       COALESCE(NULLIF(TRIM(api_base_url), ''), '')
		FROM model_providers
		WHERE status = 'active'
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	newCache := make(map[string]LitellmTemplateEntry)
	for rows.Next() {
		var code, tpl, keyEnv, gwBase, providerBase string
		if err := rows.Scan(&code, &tpl, &keyEnv, &gwBase, &providerBase); err != nil {
			continue
		}
		code = strings.TrimSpace(code)
		tpl = strings.TrimSpace(tpl)
		if tpl == "" {
			continue
		}
		newCache[code] = LitellmTemplateEntry{
			Template:           tpl,
			APIKeyEnv:          strings.TrimSpace(keyEnv),
			APIBase:            strings.TrimSpace(gwBase),
			ProviderAPIBaseURL: strings.TrimSpace(providerBase),
		}
	}

	litellmCacheMu.Lock()
	litellmTemplateCache = newCache
	litellmCacheLoaded = true
	litellmCacheMu.Unlock()
}

func InitLitellmCache() {
	litellmCacheLoadOnce.Do(func() {
		LoadLitellmTemplateCache()
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				LoadLitellmTemplateCache()
			}
		}()
	})
}

func ResolveLitellmModelFromCache(provider, model string) (string, error) {
	litellmCacheMu.RLock()
	defer litellmCacheMu.RUnlock()

	if !litellmCacheLoaded {
		return "", errCacheNotLoaded
	}

	entry, ok := litellmTemplateCache[strings.TrimSpace(provider)]
	if !ok || entry.Template == "" {
		return "", errProviderNotInCache
	}

	modelName := model
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		modelName = model[idx+1:]
	}

	tpl := entry.Template
	if strings.Contains(tpl, "{model_id}") {
		return strings.ReplaceAll(tpl, "{model_id}", modelName), nil
	}

	return tpl + "/" + modelName, nil
}

func GetLitellmTemplateCache() map[string]LitellmTemplateEntry {
	litellmCacheMu.RLock()
	defer litellmCacheMu.RUnlock()

	result := make(map[string]LitellmTemplateEntry, len(litellmTemplateCache))
	for k, v := range litellmTemplateCache {
		result[k] = v
	}
	return result
}

func ResetLitellmCacheForTest() {
	litellmCacheMu.Lock()
	defer litellmCacheMu.Unlock()
	litellmTemplateCache = nil
	litellmCacheLoaded = false
	litellmCacheLoadOnce = sync.Once{}
}

func SetLitellmCacheForTest(cache map[string]LitellmTemplateEntry) {
	litellmCacheMu.Lock()
	defer litellmCacheMu.Unlock()
	litellmTemplateCache = cache
	litellmCacheLoaded = true
}

func ResolveLitellmUpstreamBaseURL(provider string) string {
	litellmCacheMu.RLock()
	defer litellmCacheMu.RUnlock()

	if !litellmCacheLoaded {
		return ""
	}

	entry, ok := litellmTemplateCache[strings.TrimSpace(provider)]
	if !ok {
		return ""
	}

	// 与 probeQuotaViaLitellmUserConfig 对齐：优先厂商 api_base_url，其次 litellm_gateway_api_base
	chosen := strings.TrimSpace(entry.ProviderAPIBaseURL)
	if chosen == "" {
		chosen = strings.TrimSpace(entry.APIBase)
	}
	return strings.TrimRight(chosen, "/")
}

var errCacheNotLoaded = fmt.Errorf("litellm template cache not loaded")
var errProviderNotInCache = fmt.Errorf("provider not found in litellm template cache")
