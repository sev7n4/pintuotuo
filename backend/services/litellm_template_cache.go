package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
)

type litellmTemplateEntry struct {
	Template  string
	APIKeyEnv string
	APIBase   string
}

var (
	litellmCacheMu      sync.RWMutex
	litellmTemplateCache map[string]litellmTemplateEntry
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
		       COALESCE(NULLIF(TRIM(litellm_gateway_api_base), ''), '')
		FROM model_providers
		WHERE status = 'active'
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	newCache := make(map[string]litellmTemplateEntry)
	for rows.Next() {
		var code, tpl, keyEnv, apiBase string
		if err := rows.Scan(&code, &tpl, &keyEnv, &apiBase); err != nil {
			continue
		}
		code = strings.TrimSpace(code)
		tpl = strings.TrimSpace(tpl)
		if tpl == "" {
			continue
		}
		newCache[code] = litellmTemplateEntry{
			Template:  tpl,
			APIKeyEnv: strings.TrimSpace(keyEnv),
			APIBase:   strings.TrimSpace(apiBase),
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

func GetLitellmTemplateCache() map[string]litellmTemplateEntry {
	litellmCacheMu.RLock()
	defer litellmCacheMu.RUnlock()

	result := make(map[string]litellmTemplateEntry, len(litellmTemplateCache))
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

func SetLitellmCacheForTest(cache map[string]litellmTemplateEntry) {
	litellmCacheMu.Lock()
	defer litellmCacheMu.Unlock()
	litellmTemplateCache = cache
	litellmCacheLoaded = true
}

var errCacheNotLoaded = fmt.Errorf("litellm template cache not loaded")
var errProviderNotInCache = fmt.Errorf("provider not found in litellm template cache")
