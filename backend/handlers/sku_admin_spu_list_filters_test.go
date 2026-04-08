package handlers

import "testing"

func TestAdminSPUListFilters_StatusAllSkipsStatusCond(t *testing.T) {
	where, args := adminSPUListFilters("all", "", "", "")
	if where != "1=1" {
		t.Fatalf("where = %q want 1=1", where)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v want empty", args)
	}
}

func TestAdminSPUListFilters_KeywordTwoPlaceholders(t *testing.T) {
	where, args := adminSPUListFilters("active", "", "", "gpt")
	if want := "1=1 AND p.status = $1 AND (p.spu_code ILIKE $2 OR p.name ILIKE $3)"; where != want {
		t.Fatalf("where = %q\nwant %q", where, want)
	}
	if len(args) != 3 || args[0] != "active" || args[1] != "%gpt%" || args[2] != "%gpt%" {
		t.Fatalf("args = %v", args)
	}
}

func TestAdminSPUListFilters_ProviderTier(t *testing.T) {
	where, args := adminSPUListFilters("", "openai", "pro", "")
	want := "1=1 AND p.model_provider = $1 AND p.model_tier = $2"
	if where != want {
		t.Fatalf("where = %q", where)
	}
	if len(args) != 2 || args[0] != "openai" || args[1] != "pro" {
		t.Fatalf("args = %v", args)
	}
}
