package capabilityprobe

import (
	"testing"

	"github.com/pintuotuo/backend/services"
)

func TestNormalizeAdminNonChatProbes(t *testing.T) {
	out, err := NormalizeAdminNonChatProbes([]string{"Embeddings", "responses"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0] != services.EndpointTypeEmbeddings || out[1] != services.EndpointTypeResponses {
		t.Fatalf("got %#v", out)
	}

	_, err = NormalizeAdminNonChatProbes([]string{"embeddings"}, true)
	if err == nil {
		t.Fatal("expected error when skip_embeddings removes all")
	}

	out, err = NormalizeAdminNonChatProbes([]string{"embeddings", "moderations"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0] != services.EndpointTypeModerations {
		t.Fatalf("got %#v", out)
	}

	_, err = NormalizeAdminNonChatProbes([]string{"chat"}, false)
	if err == nil {
		t.Fatal("expected error for unknown probe")
	}
}
