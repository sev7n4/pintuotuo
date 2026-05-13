package capabilityprobe

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

// AppendNonOpenAISkipRows appends skipped post_* rows for non-openai api_format (same notes as CLI).
func AppendNonOpenAISkipRows(out *[]Row, ts string, key *models.MerchantAPIKey, apiFormat, routeMode string) {
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
		appendRow(out, ts, key.ID, key.MerchantID, key.Provider, apiFormat, routeMode, probe, "0", "skipped", "api_format_not_openai_use_anthropic_native_or_other_adapter")
	}
}

// ProbeScannedKey runs get_models + openai-format probes (or skips) for one key row from the probe SQL.
func ProbeScannedKey(
	ctx context.Context,
	hc *services.HealthChecker,
	client *http.Client,
	longClient *http.Client,
	ts string,
	key *models.MerchantAPIKey,
	apiFormat string,
	mpCode, mpAPIBase, mpProviderRegion string,
	mpEndpoints, mpRouteStrategy []byte,
	pf ProbeFlags,
) []Row {
	var out []Row

	routeMode := ""
	if _, rm, e2 := services.ResolveMerchantAPIKeyUpstreamBase(ctx, key); e2 == nil {
		routeMode = rm
	}

	full, err := hc.FullVerification(ctx, key)
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
	appendRow(&out, ts, key.ID, key.MerchantID, key.Provider, apiFormat, routeMode, "get_models", itoa(httpCode), ok, truncate(note, 500))

	if strings.ToLower(strings.TrimSpace(apiFormat)) != "openai" {
		AppendNonOpenAISkipRows(&out, ts, key, apiFormat, routeMode)
		return out
	}

	runOpenAIFormatProbes(ctx, &out, ts, client, longClient, key, apiFormat, routeMode, mpCode, mpAPIBase, mpProviderRegion, mpEndpoints, mpRouteStrategy, pf)
	return out
}

// AdminRunOptions configures admin single-key capability probe (non-billable; shorter defaults than CLI).
type AdminRunOptions struct {
	SkipEmbeddings bool
	HTTPTimeout    time.Duration
	LongTimeout    time.Duration
}

const adminCapabilityProbeQuery = `
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
WHERE mak.id = $1
  AND mak.status = 'active'
  AND mp.status = 'active'
  AND m.status IN ('active', 'approved')
  AND COALESCE(m.lifecycle_status, 'active') <> 'suspended'
`

// RunForAdminAPIKeyID loads one active key (no verified_at gate) and returns Phase0-style probe rows.
func RunForAdminAPIKeyID(ctx context.Context, db *sql.DB, keyID int, opts AdminRunOptions) ([]Row, error) {
	if opts.HTTPTimeout <= 0 {
		opts.HTTPTimeout = 45 * time.Second
	}
	if opts.LongTimeout <= 0 {
		opts.LongTimeout = 90 * time.Second
	}

	var (
		key                          models.MerchantAPIKey
		routeConfigBytes             []byte
		mpCode, apiFormat            string
		mpAPIBase, mpProviderRegion  string
		mpEndpoints, mpRouteStrategy []byte
	)
	err := db.QueryRowContext(ctx, adminCapabilityProbeQuery, keyID).Scan(
		&key.ID, &key.MerchantID, &key.Name, &key.Provider, &key.APIKeyEncrypted,
		&key.EndpointURL, &key.FallbackEndpointURL, &key.RouteMode,
		&routeConfigBytes, &key.Region,
		&mpCode, &apiFormat, &mpAPIBase, &mpProviderRegion,
		&mpEndpoints, &mpRouteStrategy,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	if len(routeConfigBytes) > 0 {
		_ = json.Unmarshal(routeConfigBytes, &key.RouteConfig)
	}

	hc := services.NewHealthChecker()
	client := &http.Client{Timeout: opts.HTTPTimeout}
	longClient := &http.Client{Timeout: opts.LongTimeout}
	ts := time.Now().UTC().Format(time.RFC3339)
	pf := ProbeFlags{
		SkipEmbeddings:     opts.SkipEmbeddings,
		Billable:           false,
		EmbeddingModel:     "text-embedding-3-small",
		ModerationModel:    "omni-moderation-latest",
		ChatModel:          "gpt-4o-mini",
		ImageModel:         "dall-e-3",
		SpeechModel:        "tts-1",
		TranscriptionModel: "whisper-1",
	}
	return ProbeScannedKey(ctx, hc, client, longClient, ts, &key, apiFormat, mpCode, mpAPIBase, mpProviderRegion, mpEndpoints, mpRouteStrategy, pf), nil
}
