package models

import (
	"testing"
)

func TestSKU_EndpointType(t *testing.T) {
	sku := SKU{
		ID:           1,
		SPUID:        1,
		SKUCode:      "test-sku",
		EndpointType: "chat_completions",
	}

	if sku.EndpointType != "chat_completions" {
		t.Errorf("EndpointType = %v, want chat_completions", sku.EndpointType)
	}
}

func TestSKUCreateRequest_EndpointType(t *testing.T) {
	req := SKUCreateRequest{
		SPUID:        1,
		SKUCode:      "test-sku",
		SKUType:      "token_pack",
		EndpointType: "embeddings",
		RetailPrice:  100.0,
	}

	if req.EndpointType != "embeddings" {
		t.Errorf("EndpointType = %v, want embeddings", req.EndpointType)
	}
}

func TestSKUUpdateRequest_EndpointType(t *testing.T) {
	endpointType := "images_generations"
	req := SKUUpdateRequest{
		EndpointType: endpointType,
	}

	if req.EndpointType != "images_generations" {
		t.Errorf("EndpointType = %v, want images_generations", req.EndpointType)
	}
}

func TestSKU_DefaultEndpointType(t *testing.T) {
	sku := SKU{
		ID:      1,
		SPUID:   1,
		SKUCode: "test-sku",
	}

	if sku.EndpointType != "" {
		t.Errorf("EndpointType should be empty by default, got %v", sku.EndpointType)
	}
}
