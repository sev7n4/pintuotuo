package utils

import (
	"encoding/json"
	"testing"
)

func TestParseRouteStrategy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		hasError bool
	}{
		{
			name:  "valid route strategy",
			input: `{"default_mode":"auto","domestic_users":{"mode":"litellm"}}`,
			expected: map[string]interface{}{
				"default_mode": "auto",
				"domestic_users": map[string]interface{}{
					"mode": "litellm",
				},
			},
			hasError: false,
		},
		{
			name:     "empty route strategy",
			input:    `{}`,
			expected: map[string]interface{}{},
			hasError: false,
		},
		{
			name:     "invalid JSON",
			input:    `{invalid}`,
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseRouteStrategy(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !compareMaps(result, tt.expected) {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestParseEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
		hasError bool
	}{
		{
			name:  "valid endpoints",
			input: `{"direct":{"domestic":"https://api.example.com"},"litellm":{"domestic":"http://litellm:4000"}}`,
			expected: map[string]interface{}{
				"direct": map[string]interface{}{
					"domestic": "https://api.example.com",
				},
				"litellm": map[string]interface{}{
					"domestic": "http://litellm:4000",
				},
			},
			hasError: false,
		},
		{
			name:     "empty endpoints",
			input:    `{}`,
			expected: map[string]interface{}{},
			hasError: false,
		},
		{
			name:     "invalid JSON",
			input:    `{invalid}`,
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEndpoints(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !compareMaps(result, tt.expected) {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestValidateRouteStrategy(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		hasError bool
	}{
		{
			name: "valid route strategy with all fields",
			input: map[string]interface{}{
				"default_mode": "auto",
				"domestic_users": map[string]interface{}{
					"mode":          "litellm",
					"fallback_mode": "proxy",
				},
			},
			hasError: false,
		},
		{
			name:     "empty route strategy",
			input:    map[string]interface{}{},
			hasError: false,
		},
		{
			name: "invalid default_mode",
			input: map[string]interface{}{
				"default_mode": "invalid_mode",
			},
			hasError: true,
		},
		{
			name: "invalid user mode",
			input: map[string]interface{}{
				"domestic_users": map[string]interface{}{
					"mode": "invalid_mode",
				},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRouteStrategy(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		hasError bool
	}{
		{
			name: "valid endpoints",
			input: map[string]interface{}{
				"direct": map[string]interface{}{
					"domestic": "https://api.example.com",
				},
				"litellm": map[string]interface{}{
					"domestic": "http://litellm:4000",
				},
			},
			hasError: false,
		},
		{
			name:     "empty endpoints",
			input:    map[string]interface{}{},
			hasError: false,
		},
		{
			name: "invalid URL",
			input: map[string]interface{}{
				"direct": map[string]interface{}{
					"domestic": "not-a-valid-url",
				},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpoints(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func compareMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok {
			return false
		} else {
			switch av := v.(type) {
			case map[string]interface{}:
				if bvm, ok := bv.(map[string]interface{}); !ok || !compareMaps(av, bvm) {
					return false
				}
			default:
				if v != bv {
					return false
				}
			}
		}
	}
	return true
}

func TestParseRouteStrategyFromJSON(t *testing.T) {
	jsonData := `{
		"default_mode": "auto",
		"domestic_users": {
			"mode": "litellm",
			"fallback_mode": "proxy",
			"proxy_endpoint": "gaap"
		},
		"overseas_users": {
			"mode": "direct"
		}
	}`

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result["default_mode"] != "auto" {
		t.Errorf("expected default_mode to be 'auto', got %v", result["default_mode"])
	}

	domesticUsers, ok := result["domestic_users"].(map[string]interface{})
	if !ok {
		t.Fatal("domestic_users is not a map")
	}

	if domesticUsers["mode"] != "litellm" {
		t.Errorf("expected domestic_users.mode to be 'litellm', got %v", domesticUsers["mode"])
	}
}

func TestParseEndpointsFromJSON(t *testing.T) {
	jsonData := `{
		"direct": {
			"domestic": "https://api.deepseek.com/v1",
			"overseas": "https://api.openai.com/v1"
		},
		"litellm": {
			"domestic": "http://litellm-domestic:4000/v1",
			"overseas": "http://litellm-overseas:4000/v1"
		},
		"proxy": {
			"gaap": "https://gaap.example.com",
			"nginx_hk": "https://proxy-hk.example.com"
		}
	}`

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 endpoint types, got %d", len(result))
	}

	direct, ok := result["direct"].(map[string]interface{})
	if !ok {
		t.Fatal("direct is not a map")
	}

	if direct["domestic"] != "https://api.deepseek.com/v1" {
		t.Errorf("unexpected domestic endpoint: %v", direct["domestic"])
	}
}
