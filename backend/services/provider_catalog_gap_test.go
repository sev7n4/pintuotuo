package services

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCompareProviderCatalog_PendingAndStale(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT MAX\(synced_at\)`).
		WithArgs("openai").
		WillReturnRows(sqlmock.NewRows([]string{"max"}).AddRow(nil))

	mock.ExpectQuery(`SELECT model_id, COALESCE\(display_name`).
		WithArgs("openai").
		WillReturnRows(sqlmock.NewRows([]string{"model_id", "display_name"}).
			AddRow("gpt-4o", "").
			AddRow("gpt-4o-mini", "Mini"))

	mock.ExpectQuery(`SELECT sp.id, sp.spu_code`).
		WithArgs("openai").
		WillReturnRows(sqlmock.NewRows([]string{"id", "spu_code", "name", "model_id", "status", "count"}).
			AddRow(1, "SPU-OLD", "Old Model", "gpt-3.5-turbo", "active", 2))

	gap, err := CompareProviderCatalogDB(context.Background(), db, "openai")
	if err != nil {
		t.Fatal(err)
	}
	if len(gap.PendingOnboard) != 2 {
		t.Fatalf("pending: got %d want 2", len(gap.PendingOnboard))
	}
	if len(gap.StaleSPUs) != 1 || gap.StaleSPUs[0].ModelID != "gpt-3.5-turbo" {
		t.Fatalf("stale: %+v", gap.StaleSPUs)
	}
	if gap.ProviderModelCount != 2 || gap.SPUModelCount != 1 {
		t.Fatalf("counts: provider=%d spu=%d", gap.ProviderModelCount, gap.SPUModelCount)
	}
}

func TestCreateSPUDraftsFromProviderModels_Idempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT sp.id, sp.spu_code`).
		WithArgs("google").
		WillReturnRows(sqlmock.NewRows([]string{"id", "spu_code", "name", "model_id", "status", "count"}))

	mock.ExpectQuery(`SELECT COALESCE\(name, code\)`).
		WithArgs("google").
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Google"))

	mock.ExpectQuery(`SELECT 1 FROM spus WHERE spu_code`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(`INSERT INTO spus`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "google", "gemini-2.0-flash", "gemini-2.0-flash").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))

	out, err := CreateSPUDraftsFromProviderModelsDB(context.Background(), db, "google", []string{"gemini-2.0-flash"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].SPUID != 99 {
		t.Fatalf("result: %+v", out)
	}
}

func TestUniqueDraftSPUCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT 1 FROM spus WHERE spu_code = \$1`).
		WithArgs("SPU-OPENAI-GPT-4O").
		WillReturnError(sql.ErrNoRows)

	code, err := uniqueDraftSPUCode(context.Background(), db, "openai", "gpt-4o")
	if err != nil {
		t.Fatal(err)
	}
	if code != "SPU-OPENAI-GPT-4O" {
		t.Fatalf("code=%s", code)
	}
}

func TestNormalizeCatalogModelID(t *testing.T) {
	if normalizeCatalogModelID(" GPT-4o ") != "gpt-4o" {
		t.Fatal("normalize failed")
	}
}

func TestSpuCatalogModelID(t *testing.T) {
	if spuCatalogModelID("claude-3", "") != "claude-3" {
		t.Fatal()
	}
	if spuCatalogModelID("", "deepseek-chat") != "deepseek-chat" {
		t.Fatal()
	}
}

func TestUniqueDraftSPUCode_Sanitizes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM spus WHERE spu_code = $1`)).
		WillReturnError(sql.ErrNoRows)
	code, err := uniqueDraftSPUCode(context.Background(), db, "openrouter", "anthropic/claude-3.5-sonnet")
	if err != nil {
		t.Fatal(err)
	}
	if code == "" {
		t.Fatal("empty code")
	}
}
