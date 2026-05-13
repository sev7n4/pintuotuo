package capabilityprobe

import (
	"fmt"
	"strings"

	"github.com/pintuotuo/backend/services"
)

var allowedAdminNonChatProbeAliases = map[string]string{
	"embeddings":  services.EndpointTypeEmbeddings,
	"embedding":   services.EndpointTypeEmbeddings,
	"moderations": services.EndpointTypeModerations,
	"moderation":  services.EndpointTypeModerations,
	"responses":   services.EndpointTypeResponses,
	"response":    services.EndpointTypeResponses,
}

// NormalizeAdminNonChatProbes maps admin JSON probe names to canonical endpoint IDs and applies skip_embeddings.
// Returns non-empty slice on success (for ProbeFlags.NonChatProbeIDs).
func NormalizeAdminNonChatProbes(requested []string, skipEmbeddings bool) ([]string, error) {
	var out []string
	seen := map[string]bool{}
	for _, raw := range requested {
		k := strings.ToLower(strings.TrimSpace(raw))
		if k == "" {
			continue
		}
		canon, ok := allowedAdminNonChatProbeAliases[k]
		if !ok {
			return nil, fmt.Errorf("unknown probe %q (allowed: embeddings, moderations, responses)", raw)
		}
		if seen[canon] {
			continue
		}
		seen[canon] = true
		out = append(out, canon)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid probes in list")
	}
	if skipEmbeddings {
		var filtered []string
		for _, id := range out {
			if id == services.EndpointTypeEmbeddings {
				continue
			}
			filtered = append(filtered, id)
		}
		out = filtered
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("skip_embeddings leaves no probes to run; include moderations or responses")
	}
	return out, nil
}
