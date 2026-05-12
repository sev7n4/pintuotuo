// Command capability-probe runs Phase 0 upstream capability checks using BYOK keys stored in PostgreSQL
// (merchant_api_keys.api_key_encrypted), decrypted with the same ENCRYPTION_KEY as the backend.
//
// Intended to run on the deployment host (or any host that can reach DATABASE_URL and upstreams), e.g.:
//
//	cd /opt/pintuotuo/backend && ./bin/capability-probe -out /tmp/cap.csv
//
// Do not pass vendor API keys on the command line; they are loaded from the DB only.
// For route_mode=litellm, Bearer may use LITELLM_MASTER_KEY from the process environment (same as health_checker).
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
)

func main() {
	outPath := flag.String("out", "", "write CSV to this path (default: stdout)")
	merchantID := flag.Int("merchant-id", 0, "only keys for this merchant_id (0 = all)")
	apiKeyID := flag.Int("api-key-id", 0, "only this merchant_api_keys.id (0 = all)")
	providerFilter := flag.String("provider", "", "only keys for this provider code (substring match)")
	limit := flag.Int("limit", 80, "max rows to process")
	skipEmbeddings := flag.Bool("skip-embeddings", false, "do not POST /v1/embeddings")
	embeddingModel := flag.String("embedding-model", "text-embedding-3-small", "model id for embeddings probe (OpenAI-format providers only)")
	flag.Parse()

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := config.InitDB(); err != nil {
		log.Fatalf("db: %v", err)
	}
	defer func() { _ = config.CloseDB() }()

	if err := utils.InitEncryption(); err != nil {
		log.Fatalf("encryption: %v", err)
	}

	out := io.Writer(os.Stdout)
	if strings.TrimSpace(*outPath) != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			log.Fatalf("create out: %v", err)
		}
		defer func() { _ = f.Close() }()
		out = f
	}

	w := csv.NewWriter(out)
	_ = w.Write([]string{
		"ts", "merchant_api_key_id", "merchant_id", "provider", "api_format", "route_mode",
		"probe", "http_code", "ok", "note",
	})
	w.Flush()

	db := config.GetDB()
	ctx := context.Background()
	hc := services.NewHealthChecker()

	args := []interface{}{}
	where := []string{
		"mak.status = 'active'",
		"mp.status = 'active'",
		"(mak.verified_at IS NOT NULL OR mak.verification_result = 'verified')",
		"m.status IN ('active', 'approved')",
		"COALESCE(m.lifecycle_status, 'active') <> 'suspended'",
	}
	if *merchantID > 0 {
		where = append(where, fmt.Sprintf("mak.merchant_id = $%d", len(args)+1))
		args = append(args, *merchantID)
	}
	if *apiKeyID > 0 {
		where = append(where, fmt.Sprintf("mak.id = $%d", len(args)+1))
		args = append(args, *apiKeyID)
	}
	if strings.TrimSpace(*providerFilter) != "" {
		where = append(where, fmt.Sprintf("lower(mak.provider) LIKE $%d", len(args)+1))
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(*providerFilter))+"%")
	}

	q := fmt.Sprintf(`
SELECT mak.id, mak.merchant_id, mak.name, mak.provider, mak.api_key_encrypted,
       COALESCE(NULLIF(TRIM(mak.endpoint_url), ''), '') AS endpoint_url,
       COALESCE(NULLIF(TRIM(mak.fallback_endpoint_url), ''), '') AS fallback_endpoint_url,
       COALESCE(NULLIF(TRIM(mak.route_mode), ''), 'auto') AS route_mode,
       COALESCE(mak.route_config, '{}'::jsonb),
       COALESCE(NULLIF(TRIM(mak.region), ''), 'domestic') AS region,
       mp.code AS mp_code,
       mp.api_format,
       COALESCE(NULLIF(TRIM(mp.api_base_url), ''), '') AS mp_api_base,
       COALESCE(NULLIF(TRIM(mp.provider_region), ''), 'domestic') AS mp_provider_region,
       COALESCE(mp.endpoints, '{}'::jsonb),
       COALESCE(mp.route_strategy, '{}'::jsonb)
FROM merchant_api_keys mak
INNER JOIN merchants m ON m.id = mak.merchant_id
INNER JOIN model_providers mp ON lower(trim(mp.code)) = lower(trim(mak.provider))
WHERE %s
ORDER BY mak.merchant_id, mak.provider, mak.id
LIMIT %d
`, strings.Join(where, " AND "), *limit)

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer func() { _ = rows.Close() }()

	ts := time.Now().UTC().Format(time.RFC3339)
	client := &http.Client{Timeout: 45 * time.Second}

	for rows.Next() {
		var (
			key                          models.MerchantAPIKey
			routeConfigBytes             []byte
			mpCode, apiFormat            string
			mpAPIBase, mpProviderRegion  string
			mpEndpoints, mpRouteStrategy []byte
		)
		if err := rows.Scan(
			&key.ID, &key.MerchantID, &key.Name, &key.Provider, &key.APIKeyEncrypted,
			&key.EndpointURL, &key.FallbackEndpointURL, &key.RouteMode,
			&routeConfigBytes, &key.Region,
			&mpCode, &apiFormat, &mpAPIBase, &mpProviderRegion,
			&mpEndpoints, &mpRouteStrategy,
		); err != nil {
			log.Fatalf("scan: %v", err)
		}
		if len(routeConfigBytes) > 0 {
			_ = json.Unmarshal(routeConfigBytes, &key.RouteConfig)
		}

		// --- GET models (same pipeline as health FullVerification) ---
		routeMode := ""
		if _, rm, e2 := services.ResolveMerchantAPIKeyUpstreamBase(ctx, &key); e2 == nil {
			routeMode = rm
		}

		full, err := hc.FullVerification(ctx, &key)
		note := ""
		httpCode := 0
		ok := "false"
		if err != nil {
			note = "health_error:" + err.Error()
		} else if full != nil {
			httpCode = full.StatusCode
			if full.Success {
				ok = "true"
				if httpCode == 0 {
					httpCode = 200
				}
			}
			note = strings.TrimSpace(full.ErrorMessage)
			if note == "" && len(full.ModelsFound) > 0 {
				note = fmt.Sprintf("models_count=%d", len(full.ModelsFound))
			}
		}
		_ = w.Write([]string{ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode, "get_models", itoa(httpCode), ok, truncate(note, 500)})
		w.Flush()

		if *skipEmbeddings {
			continue
		}
		if strings.ToLower(strings.TrimSpace(apiFormat)) != "openai" {
			_ = w.Write([]string{ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode, "post_embeddings", "0", "skipped", "api_format_not_openai"})
			w.Flush()
			continue
		}

		cfg, err := buildExecCfg(mpCode, mpAPIBase, apiFormat, mpProviderRegion, mpRouteStrategy, mpEndpoints, &key)
		if err != nil {
			_ = w.Write([]string{ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode, "post_embeddings", "0", "false", "build_cfg:" + err.Error()})
			w.Flush()
			continue
		}
		embedURL := services.ResolveEndpointByType(cfg, services.EndpointTypeEmbeddings)
		if embedURL == "" {
			_ = w.Write([]string{ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode, "post_embeddings", "0", "false", "empty_embed_url"})
			w.Flush()
			continue
		}

		decrypted, derr := utils.Decrypt(key.APIKeyEncrypted)
		if derr != nil || decrypted == "" {
			_ = w.Write([]string{ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode, "post_embeddings", "0", "false", "decrypt_failed"})
			w.Flush()
			continue
		}
		authTok := decrypted
		if cfg.GatewayMode == services.GatewayModeLitellm {
			if mk := strings.TrimSpace(os.Getenv("LITELLM_MASTER_KEY")); mk != "" {
				authTok = mk
			}
		}

		body := fmt.Sprintf(`{"model":%q,"input":"probe"}`, strings.TrimSpace(*embeddingModel))
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, embedURL, strings.NewReader(body))
		if err != nil {
			_ = w.Write([]string{ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode, "post_embeddings", "0", "false", err.Error()})
			w.Flush()
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authTok)

		resp, err := client.Do(req)
		if err != nil {
			_ = w.Write([]string{ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode, "post_embeddings", "0", "false", "http:" + err.Error()})
			w.Flush()
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		eok := "false"
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			eok = "true"
		}
		enote := resp.Status
		_ = w.Write([]string{ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode, "post_embeddings", itoa(resp.StatusCode), eok, truncate(enote, 200)})
		w.Flush()
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows: %v", err)
	}
	w.Flush()
}

func itoa(v int) string {
	return fmt.Sprintf("%d", v)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func buildExecCfg(
	mpCode, mpAPIBase, apiFormat, mpProviderRegion string,
	mpRouteStrategy, mpEndpoints []byte,
	key *models.MerchantAPIKey,
) (*services.ExecutionProviderConfig, error) {
	var rs, ep map[string]interface{}
	if err := json.Unmarshal(mpRouteStrategy, &rs); err != nil {
		rs = map[string]interface{}{}
	}
	if err := json.Unmarshal(mpEndpoints, &ep); err != nil {
		ep = map[string]interface{}{}
	}
	cfg := &services.ExecutionProviderConfig{
		Code:             mpCode,
		Name:             mpCode,
		APIBaseURL:       mpAPIBase,
		APIFormat:        apiFormat,
		ProviderRegion:   mpProviderRegion,
		RouteStrategy:    rs,
		Endpoints:        ep,
		BYOKEndpointURL:  strings.TrimSpace(key.EndpointURL),
		BYOKRouteMode:    key.RouteMode,
		BYOKRouteConfig:  key.RouteConfig,
		BYOKFallbackURL:  strings.TrimSpace(key.FallbackEndpointURL),
	}
	services.ConfigureGatewayMode(cfg)
	return cfg, nil
}
