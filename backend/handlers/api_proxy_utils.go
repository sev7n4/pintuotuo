package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
)

func candidateScoreFromMap(m map[string]any) float64 {
	if m == nil {
		return 0
	}
	for _, k := range []string{"Score", "score"} {
		if v, ok := m[k]; ok {
			switch x := v.(type) {
			case float64:
				return x
			case json.Number:
				f, _ := x.Float64()
				return f
			}
		}
	}
	return 0
}

func intFromMap(m map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch x := v.(type) {
			case float64:
				return int(x)
			case int:
				return x
			case json.Number:
				i64, _ := x.Int64()
				return int(i64)
			}
		}
	}
	return 0
}

func stringFromMap(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func summarizeRoutingCandidatesForTrace(candidatesJSON json.RawMessage) (int, *traceTopCandidate) {
	if len(candidatesJSON) == 0 {
		return 0, nil
	}
	var items []map[string]any
	if err := json.Unmarshal(candidatesJSON, &items); err != nil || len(items) == 0 {
		return 0, nil
	}
	bestIdx := 0
	bestScore := candidateScoreFromMap(items[0])
	for i := 1; i < len(items); i++ {
		s := candidateScoreFromMap(items[i])
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	best := items[bestIdx]
	return len(items), &traceTopCandidate{
		APIKeyID: intFromMap(best, "APIKeyID", "api_key_id", "ApiKeyID"),
		Provider: stringFromMap(best, "Provider", "provider"),
		Model:    stringFromMap(best, "Model", "model"),
		Score:    candidateScoreFromMap(best),
	}
}

func normalizeEffectivePolicySource(stored string) string {
	switch strings.TrimSpace(stored) {
	case policySourceEnv, policySourceDB, policySourceDefault:
		return strings.TrimSpace(stored)
	default:
		return policySourceDefault
	}
}

func parseRoutingDecisionPayload(raw json.RawMessage) (routingDecisionPayload, error) {
	if len(raw) == 0 {
		return routingDecisionPayload{}, nil
	}
	var wrapped routingDecisionPayload
	if err := json.Unmarshal(raw, &wrapped); err == nil && (len(wrapped.Candidates) > 0 || wrapped.StrategyRuntime.StrategyCode != "" || strings.TrimSpace(wrapped.EffectivePolicySource) != "") {
		return wrapped, nil
	}
	return routingDecisionPayload{}, fmt.Errorf("invalid routing decision payload")
}

func legacyProviderList() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":         "openai",
			"display_name": "OpenAI",
			"models": []map[string]interface{}{
				{"name": "gpt-4o", "display_name": "GPT-4o"},
				{"name": "gpt-4o-mini", "display_name": "GPT-4o Mini"},
				{"name": "gpt-4-turbo", "display_name": "GPT-4 Turbo"},
				{"name": "gpt-3.5-turbo", "display_name": "GPT-3.5 Turbo"},
			},
		},
		{
			"name":         "anthropic",
			"display_name": "Anthropic",
			"models": []map[string]interface{}{
				{"name": "claude-3-5-sonnet-20241022", "display_name": "Claude 3.5 Sonnet"},
				{"name": "claude-3-opus-20240229", "display_name": "Claude 3 Opus"},
				{"name": "claude-3-sonnet-20240229", "display_name": "Claude 3 Sonnet"},
				{"name": "claude-3-haiku-20240307", "display_name": "Claude 3 Haiku"},
			},
		},
	}
}

func estimateInputTokens(messages []ChatMessage) int {
	total := 0
	for _, m := range messages {
		total += len(m.Content) / 4
	}
	if total == 0 {
		total = 100
	}
	return total
}
