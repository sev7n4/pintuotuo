package handlers

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/services"
)

func TestFilterEntitledAPIKeyIDsByOutboundProvider(t *testing.T) {
	ent := &services.EntitlementRoutingContext{
		AllowedAPIKeyIDs: map[int]struct{}{
			25: {},
			39: {},
			44: {},
		},
		APIKeyToMerchantSKU: map[int]int{
			25: 27,
			44: 28,
		},
	}

	t.Run("empty outbound returns all sorted", func(t *testing.T) {
		got := filterEntitledAPIKeyIDsByOutboundProvider(nil, ent, "")
		require.Equal(t, []int{25, 39, 44}, got)
	})

	t.Run("filters by provider via db", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id"}).AddRow(44)
		mock.ExpectQuery(`SELECT mak\.id`).
			WithArgs(sqlmock.AnyArg(), "alibaba_anthropic").
			WillReturnRows(rows)

		got := filterEntitledAPIKeyIDsByOutboundProvider(db, ent, "alibaba_anthropic")
		require.Equal(t, []int{44}, got)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no match returns empty", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id"})
		mock.ExpectQuery(`SELECT mak\.id`).
			WithArgs(sqlmock.AnyArg(), "alibaba_anthropic").
			WillReturnRows(rows)

		got := filterEntitledAPIKeyIDsByOutboundProvider(db, ent, "alibaba_anthropic")
		require.Empty(t, got)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPickDeterministicEntitledKey_OutboundProvider(t *testing.T) {
	ent := &services.EntitlementRoutingContext{
		AllowedAPIKeyIDs: map[int]struct{}{
			25: {},
			44: {},
		},
		APIKeyToMerchantSKU: map[int]int{
			25: 27,
			44: 28,
		},
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(44)
	mock.ExpectQuery(`SELECT mak\.id`).
		WithArgs(sqlmock.AnyArg(), "alibaba_anthropic").
		WillReturnRows(rows)

	keyID, msID := pickDeterministicEntitledKey(db, ent, "alibaba_anthropic")
	require.Equal(t, 44, keyID)
	require.Equal(t, 28, msID)
	require.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectQuery(`SELECT mak\.id`).
		WithArgs(sqlmock.AnyArg(), "alibaba").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(25))

	keyID, msID = pickDeterministicEntitledKey(db, ent, "alibaba")
	require.Equal(t, 25, keyID)
	require.Equal(t, 27, msID)
	require.NoError(t, mock.ExpectationsWereMet())
}
