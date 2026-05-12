package services

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPromoteFlashSaleStatuses_ExecBothUpdates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec(`UPDATE flash_sales SET status = 'active'`).
		WithArgs(now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE flash_sales SET status = 'ended'`).
		WithArgs(now).
		WillReturnResult(sqlmock.NewResult(0, 0))

	changed, err := PromoteFlashSaleStatuses(db, now)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true when first update affects rows")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPromoteFlashSaleStatuses_NoChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectExec(`UPDATE flash_sales SET status = 'active'`).
		WithArgs(now).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`UPDATE flash_sales SET status = 'ended'`).
		WithArgs(now).
		WillReturnResult(sqlmock.NewResult(0, 0))

	changed, err := PromoteFlashSaleStatuses(db, now)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected changed=false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
