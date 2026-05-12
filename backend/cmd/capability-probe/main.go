// Command capability-probe runs Phase 0 upstream capability checks using BYOK keys stored in PostgreSQL
// (merchant_api_keys.api_key_encrypted), decrypted with the same ENCRYPTION_KEY as the backend.
//
// Intended to run inside the backend container on the deployment host:
//
//	docker exec pintuotuo-backend /app/capability-probe -out /tmp/cap.csv
//
// Vendor API keys are never read from operator-supplied OPENAI_* env vars; they come from the DB only.
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
	limit := flag.Int("limit", 80, "max BYOK rows to process")
	skipEmbeddings := flag.Bool("skip-embeddings", false, "skip POST embeddings")
	billable := flag.Bool("billable", false, "enable probes that incur upstream usage (chat, images, speech, STT)")
	httpTimeout := flag.Duration("http-timeout", 60*time.Second, "HTTP client timeout for most probes")
	longTimeout := flag.Duration("long-http-timeout", 180*time.Second, "timeout for images/audio/chat")

	embeddingModel := flag.String("embedding-model", "text-embedding-3-small", "embeddings model id")
	moderationModel := flag.String("moderation-model", "omni-moderation-latest", "moderations model id")
	chatModel := flag.String("chat-model", "gpt-4o-mini", "chat / responses model id")
	imageModel := flag.String("image-model", "dall-e-3", "images generations/edits model id")
	speechModel := flag.String("speech-model", "tts-1", "audio speech model id")
	transcriptionModel := flag.String("transcription-model", "whisper-1", "audio transcription/translation model id")
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
	client := &http.Client{Timeout: *httpTimeout}
	longClient := &http.Client{Timeout: *longTimeout}

	pf := probeFlags{
		skipEmbeddings:     *skipEmbeddings,
		billable:           *billable,
		embeddingModel:     strings.TrimSpace(*embeddingModel),
		moderationModel:    strings.TrimSpace(*moderationModel),
		chatModel:          strings.TrimSpace(*chatModel),
		imageModel:         strings.TrimSpace(*imageModel),
		speechModel:        strings.TrimSpace(*speechModel),
		transcriptionModel: strings.TrimSpace(*transcriptionModel),
	}

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

	for rows.Next() {
		var (
			key                           models.MerchantAPIKey
			routeConfigBytes              []byte
			mpCode, apiFormat             string
			mpAPIBase, mpProviderRegion   string
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
		writeProbeRow(w, ts, &key, apiFormat, routeMode, "get_models", itoa(httpCode), ok, truncate(note, 500))

		if strings.ToLower(strings.TrimSpace(apiFormat)) != "openai" {
			writeNonOpenAISkips(w, ts, &key, apiFormat, routeMode)
			continue
		}

		runOpenAIFormatProbes(ctx, w, ts, client, longClient, &key, apiFormat, routeMode, mpCode, mpAPIBase, mpProviderRegion, mpEndpoints, mpRouteStrategy, pf)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows: %v", err)
	}
	w.Flush()
}

func writeNonOpenAISkips(w *csv.Writer, ts string, key *models.MerchantAPIKey, apiFormat, routeMode string) {
	for _, probe := range []string{
		"post_" + services.EndpointTypeChatCompletions,
		"post_" + services.EndpointTypeEmbeddings,
		"post_" + services.EndpointTypeImagesGenerations,
		"post_" + services.EndpointTypeImagesVariations,
		"post_" + services.EndpointTypeImagesEdits,
		"post_" + services.EndpointTypeAudioTranscriptions,
		"post_" + services.EndpointTypeAudioTranslations,
		"post_" + services.EndpointTypeAudioSpeech,
		"post_" + services.EndpointTypeModerations,
		"post_" + services.EndpointTypeResponses,
	} {
		writeProbeRow(w, ts, key, apiFormat, routeMode, probe, "0", "skipped", "api_format_not_openai_use_anthropic_native_or_other_adapter")
	}
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
		Code:            mpCode,
		Name:            mpCode,
		APIBaseURL:      mpAPIBase,
		APIFormat:       apiFormat,
		ProviderRegion:  mpProviderRegion,
		RouteStrategy:   rs,
		Endpoints:       ep,
		BYOKEndpointURL: strings.TrimSpace(key.EndpointURL),
		BYOKRouteMode:   key.RouteMode,
		BYOKRouteConfig: key.RouteConfig,
		BYOKFallbackURL: strings.TrimSpace(key.FallbackEndpointURL),
	}
	services.ConfigureGatewayMode(cfg)
	return cfg, nil
}
