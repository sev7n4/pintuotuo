package models

import "time"

// PricingVersion is a published retail price list snapshot (order-time binding).
type PricingVersion struct {
	ID            int       `json:"id"`
	Code          string    `json:"code"`
	Description   string    `json:"description,omitempty"`
	EffectiveFrom time.Time `json:"effective_from"`
	CreatedAt     time.Time `json:"created_at"`
}

// PricingVersionSPURate is per-SPU input/output rate under a version (CNY per 1K tokens).
type PricingVersionSPURate struct {
	ID                 int       `json:"id"`
	PricingVersionID   int       `json:"pricing_version_id"`
	SPUID              int       `json:"spu_id"`
	ProviderInputRate  float64   `json:"provider_input_rate"`
	ProviderOutputRate float64   `json:"provider_output_rate"`
	CreatedAt          time.Time `json:"created_at"`
}
