package services

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestListOpenAIModelsFromCatalog(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"code", "model_id", "ts", "provider_input_rate", "provider_output_rate"}).
		AddRow("deepseek", "deepseek-chat", ts, 0.001, 0.002)
	mock.ExpectQuery(`SELECT\s+mp\.code`).WillReturnRows(rows)

	resp, err := ListOpenAIModelsFromCatalog(context.Background(), db)
	require.NoError(t, err)
	require.Len(t, resp.Data, 1)
	require.Equal(t, "deepseek/deepseek-chat", resp.Data[0].ID)
	require.Equal(t, "deepseek", resp.Data[0].OwnedBy)
	require.Equal(t, ts.Unix(), resp.Data[0].Created)
	require.NotNil(t, resp.Data[0].Pricing)
	require.Equal(t, 0.001, resp.Data[0].Pricing.InputPer1KTokens)
	require.Equal(t, 0.002, resp.Data[0].Pricing.OutputPer1KTokens)
	require.Equal(t, "CNY", resp.Data[0].Pricing.Currency)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListOpenAIModelsEntitledForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"code", "model_id", "ts", "provider_input_rate", "provider_output_rate"}).
		AddRow("deepseek", "deepseek-chat", ts, 0.001, 0.002)
	mock.ExpectQuery(`SELECT DISTINCT ON`).WillReturnRows(rows)

	resp, err := ListOpenAIModelsEntitledForUser(context.Background(), db, 13)
	require.NoError(t, err)
	require.Len(t, resp.Data, 1)
	require.Equal(t, "deepseek/deepseek-chat", resp.Data[0].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}
