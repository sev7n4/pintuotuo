package capabilityprobe

import (
	"encoding/json"
	"strings"

	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
)

// BuildExecutionConfig builds ExecutionProviderConfig from model_provider row + merchant key (same as cmd/capability-probe).
func BuildExecutionConfig(
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
		BYOKRegion:      strings.TrimSpace(key.Region),
	}
	services.ConfigureGatewayMode(cfg)
	return cfg, nil
}
