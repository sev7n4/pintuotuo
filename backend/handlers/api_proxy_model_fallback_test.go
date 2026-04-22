package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/pintuotuo/backend/models"
)

func TestChatCompletionJSONIndicatesProxySuccess(t *testing.T) {
	if !chatCompletionJSONIndicatesProxySuccess(http.StatusOK, []byte(`{"id":"1","choices":[{"message":{"role":"a","content":"x"}}],"usage":{"total_tokens":1}}`)) {
		t.Fatal("expected success without error field")
	}
	if chatCompletionJSONIndicatesProxySuccess(http.StatusOK, []byte(`{"error":{"message":"x"}}`)) {
		t.Fatal("error field should fail")
	}
	if chatCompletionJSONIndicatesProxySuccess(http.StatusBadRequest, []byte(`{}`)) {
		t.Fatal("non-200 should fail")
	}
}

func TestResolveProxyAttemptRuntime_ReuseBaseForSameProvider(t *testing.T) {
	baseKey := models.MerchantAPIKey{ID: 42}
	baseCfg := providerRuntimeConfig{
		Code:       "openai",
		APIBaseURL: "https://api.openai.com/v1",
		APIFormat:  apiFormatOpenAI,
	}
	req := APIProxyRequest{Provider: "openai", Model: "gpt-4o-mini"}
	att := proxyCatalogAttempt{provider: "openai", model: "gpt-4.1-mini"}

	pk, dk, pcfg, skip, fatalErr := resolveProxyAttemptRuntime(
		context.Background(),
		nil,
		1,
		1,
		req,
		att,
		baseKey,
		"plain-key",
		baseCfg,
		nil,
		"req-1",
	)
	if fatalErr != nil {
		t.Fatalf("unexpected fatalErr: %v", fatalErr)
	}
	if skip {
		t.Fatal("expected no skip for same provider")
	}
	if pk.ID != baseKey.ID || dk != "plain-key" || pcfg.Code != baseCfg.Code {
		t.Fatalf("unexpected resolved runtime: key=%d decrypted=%q provider=%q", pk.ID, dk, pcfg.Code)
	}
}
