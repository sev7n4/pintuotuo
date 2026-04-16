package services

import (
	"fmt"
	"strings"
)

// OrderLinePolicyInput carries checkout-time SKU/SPU facts for fuel-pack policy checks.
type OrderLinePolicyInput struct {
	SKUType         string
	ModelProvider   string
	ModelName       string
	ProviderModelID string
}

func normalizeModelProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

// IsStrictModelSKU returns true when an SPU represents a real model entitlement under strict matching semantics.
func IsStrictModelSKU(modelProvider, modelName, providerModelID string) bool {
	provider := normalizeModelProvider(modelProvider)
	if provider == "" {
		return false
	}
	// Keep aligned with strict entitlement intent: internal virtual SPUs do not grant API model entitlement.
	if provider == "internal" || provider == "virtual_goods" {
		return false
	}
	if strings.TrimSpace(providerModelID) != "" {
		return true
	}
	return strings.TrimSpace(modelName) != ""
}

func isTokenPackSKUType(skuType string) bool {
	return strings.EqualFold(strings.TrimSpace(skuType), "token_pack")
}

// ValidateFuelPackBundle checks "token_pack cannot be purchased alone":
// if any token_pack exists, at least one strict model SKU must exist in the same order.
func ValidateFuelPackBundle(lines []OrderLinePolicyInput) error {
	hasTokenPack := false
	hasModelSKU := false
	for _, line := range lines {
		if isTokenPackSKUType(line.SKUType) {
			hasTokenPack = true
		}
		if IsStrictModelSKU(line.ModelProvider, line.ModelName, line.ProviderModelID) {
			hasModelSKU = true
		}
	}
	if hasTokenPack && !hasModelSKU {
		return fmt.Errorf("加油包不可单独购买，请至少搭配一个在售模型 SKU 或套餐包")
	}
	return nil
}
