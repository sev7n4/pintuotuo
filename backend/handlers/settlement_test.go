package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"testing"
	"time"

	 "github.com/DATA-DOG/go-sqlmock"
	 "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/pintuotuo/backend/config"
    apperrors "github.com/pintuotuo/backend/errors"
    "github.com/pintuotuo/backend/middleware"
    "github.com/pintuotuo/backend/models"
    "github.com/pintuotuo/backend/services"
)

func TestGetMerchantSettlementByID(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    service := &services.SettlementService{db: db}

    t.Run("settlement not found", func(t *testing.T) {
        mock.ExpectQuery(`SELECT id FROM merchant_settlements WHERE id =`).
            WithArgs(999).
            WillReturnRows(sqlmock.NewRows([]string{"id"}))

        settlement, err := service.GetMerchantSettlementByID(999)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "settlement not found")
    })

    t.Run("merchant not authorized", func(t *testing.T) {
        mock.ExpectQuery(`SELECT id, merchant_id FROM merchant_settlements WHERE id =`).
            WithArgs(1).
            WillReturnRows(sqlmock.NewRows([]string{"id", "merchant_id"}).
                AddRow(2))

        settlement, err := service.GetMerchantSettlementByID(1)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "not authorized")
    }
}

func TestConfirmSettlement(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    service := &services.SettlementService{db: db}

    t.Run("successful confirmation", func(t *testing.T) {
        mock.ExpectQuery(`SELECT merchant_id, status FROM merchant_settlements WHERE id = $1`).
            WithArgs(1).
            WillReturnRows(sqlmock.NewRows([]string{"merchant_id", "status"}).
                AddRow(1, "pending"))

        mock.ExpectExec(`UPDATE merchant_settlements SET merchant_confirmed`).
            WithArgs(true, 1).
            WillReturnResult(sqlmock.NewResult(0, 1))

        err := service.MerchantConfirm(1, 1)
        assert.NoError(t, err)
    })

    t.Run("wrong merchant", func(t *testing.T) {
        mock.ExpectQuery(`SELECT merchant_id, status FROM merchant_settlements WHERE id = $1`).
            WithArgs(1).
            WillReturnRows(sqlmock.NewRows([]string{"merchant_id", "status"}).
                AddRow(2, "pending")

        err := service.MerchantConfirm(1, 2)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "not authorized")
    }
}

func TestSubmitDispute(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()

    service := &services.SettlementService{db: db}

    t.Run("successful submission", func(t *testing.T) {
        mock.ExpectQuery(`SELECT merchant_id FROM merchant_settlements WHERE id = $1`).
            WithArgs(1).
            WillReturnRows(sqlmock.NewRows([]string{"id"}))

        mock.ExpectQuery(`INSERT INTO settlement_disputes`).
            WithArgs(1, 1, "amount_error", "Incorrect order count", 10000.00, 9500.00, "pending").
            WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

        dispute, err := service.SubmitDispute(1, 1, "amount_error", "Incorrect order count", 10000.00, 9500.00, time.Now())
        assert.NoError(t, err)
        assert.NotNil(t, dispute)
        assert.Equal(t, 1, dispute.SettlementID)
        assert.Equal(t, "amount_error", dispute.DisputeType)
        assert.Equal(t, 10000.00, dispute.DisputedAmount)
        assert.Equal(t, "pending", dispute.Status)
        assert.Equal(t, time.Now().Year, dispute.DisputeReason, "Incorrect order count")
    })
}

