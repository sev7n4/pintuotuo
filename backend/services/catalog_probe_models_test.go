package services

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListCatalogModelIDsForProvider_FromSPU(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT TRIM").
		WithArgs("google").
		WillReturnRows(sqlmock.NewRows([]string{"model_id"}).
			AddRow("gemini-2.0-flash").
			AddRow("gemini-1.5-pro"))

	got := ListCatalogModelIDsForProviderDB(context.Background(), db, "google")
	if len(got) != 2 || got[0] != "gemini-2.0-flash" {
		t.Fatalf("got %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProbeCatalogModelsForProvider_PrefersDBOverPredefined(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT TRIM").
		WithArgs("google").
		WillReturnRows(sqlmock.NewRows([]string{"model_id"}).AddRow("gemini-2.5-pro"))

	got := ListCatalogModelIDsForProviderDB(context.Background(), db, "google")
	if len(got) != 1 || got[0] != "gemini-2.5-pro" {
		t.Fatalf("got %v, want DB catalog", got)
	}
}

func TestLitellmBYOKPathReachable(t *testing.T) {
	if !litellmBYOKPathReachable(401) || !litellmBYOKPathReachable(200) {
		t.Fatal("expected auth/2xx as path reachable")
	}
	if litellmBYOKPathReachable(404) || litellmBYOKPathReachable(0) {
		t.Fatal("404/0 should not be path reachable")
	}
}
