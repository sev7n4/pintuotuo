package services

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSQLReserveSKUStockForOrder_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(
		`UPDATE skus SET stock = CASE WHEN stock = -1 THEN -1 ELSE stock - \$1 END WHERE id = \$2 AND \(stock = -1 OR stock >= \$1\)`,
	).WithArgs(3, 42).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = db.Exec(SQLReserveSKUStockForOrder, 3, 42)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLRestoreSKUStockFromOrderItems_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(
		`(?s)UPDATE skus s\s+SET stock = CASE WHEN s\.stock = -1 THEN -1 ELSE s\.stock \+ oi\.qty END.+WHERE order_id = \$1`,
	).WithArgs(99).WillReturnResult(sqlmock.NewResult(0, 2))

	_, err = db.Exec(SQLRestoreSKUStockFromOrderItems, 99)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLRestoreFlashSaleStockFromOrderItems_Exec(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(
		`(?s)UPDATE flash_sale_products fsp.+WHERE order_id = \$1`,
	).WithArgs(42).WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = db.Exec(SQLRestoreFlashSaleStockFromOrderItems, 42)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
