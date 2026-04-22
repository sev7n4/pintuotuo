package utils

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidJSON        = errors.New("invalid JSON format")
	ErrInvalidMode        = errors.New("invalid route mode")
	ErrInvalidURL         = errors.New("invalid endpoint URL")
	ErrInvalidProxyMode   = errors.New("proxy mode requires proxy_endpoint")
)

var validModes = map[string]bool{
	"auto":    true,
	"direct":  true,
	"litellm": true,
	"proxy":   true,
}

func ParseRouteStrategy(input string) (map[string]interface{}, error) {
	if input == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return nil, ErrInvalidJSON
	}

	return result, nil
}

func ParseEndpoints(input string) (map[string]interface{}, error) {
	if input == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return nil, ErrInvalidJSON
	}

	return result, nil
}

func ValidateRouteStrategy(strategy map[string]interface{}) error {
	if len(strategy) == 0 {
		return nil
	}

	if defaultMode, ok := strategy["default_mode"].(string); ok {
		if !validModes[defaultMode] {
			return ErrInvalidMode
		}
	}

	userTypes := []string{"domestic_users", "overseas_users", "enterprise_users"}
	for _, userType := range userTypes {
		if userStrategy, ok := strategy[userType].(map[string]interface{}); ok {
			if mode, ok := userStrategy["mode"].(string); ok {
				if !validModes[mode] {
					return ErrInvalidMode
				}
			}

			if fallbackMode, ok := userStrategy["fallback_mode"].(string); ok {
				if !validModes[fallbackMode] {
					return ErrInvalidMode
				}
			}

			if mode, ok := userStrategy["mode"].(string); ok && mode == "proxy" {
				if _, ok := userStrategy["proxy_endpoint"]; !ok {
					return ErrInvalidProxyMode
				}
			}
		}
	}

	return nil
}

func ValidateEndpoints(endpoints map[string]interface{}) error {
	if len(endpoints) == 0 {
		return nil
	}

	for _, endpointsByRegion := range endpoints {
		if endpointsMap, ok := endpointsByRegion.(map[string]interface{}); ok {
			for _, endpoint := range endpointsMap {
				if endpointStr, ok := endpoint.(string); ok && endpointStr != "" {
					if !isValidURL(endpointStr) {
						return ErrInvalidURL
					}
				}
			}
		}
	}

	return nil
}

func isValidURL(rawURL string) bool {
	if rawURL == "" {
		return true
	}

	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		_, err := url.Parse(rawURL)
		return err == nil
	}

	return false
}
