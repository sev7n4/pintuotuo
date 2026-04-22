package handlers

import (
	"net/http"
	"testing"
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
