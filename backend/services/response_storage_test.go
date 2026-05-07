package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pintuotuo/backend/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResponseStorageService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	assert.NotNil(t, svc)
	assert.Equal(t, db, svc.db)
}

func TestResponseStorageService_StoreResponse(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	t.Run("successful store with defaults", func(t *testing.T) {
		resp := &models.StoredResponse{
			UserID:     1,
			MerchantID: 2,
			Model:      "gpt-4o",
			Input:      json.RawMessage(`"Hello"`),
		}

		mock.ExpectExec(`INSERT INTO stored_responses`).
			WithArgs(
				sqlmock.AnyArg(),
				resp.UserID,
				resp.MerchantID,
				resp.Model,
				resp.Input,
				resp.Output,
				resp.ToolCalls,
				resp.Usage,
				"completed",
				resp.BackgroundJobID,
				resp.ErrorMessage,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := svc.StoreResponse(ctx, resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ResponseID)
		assert.Equal(t, "completed", resp.Status)
		assert.False(t, resp.ExpiresAt.IsZero())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("store with explicit values", func(t *testing.T) {
		resp := &models.StoredResponse{
			ResponseID:      "resp_custom123",
			UserID:          1,
			MerchantID:      2,
			Model:           "gpt-4o",
			Input:           json.RawMessage(`"Hello"`),
			Output:          json.RawMessage(`[{"type":"message"}]`),
			Status:          "queued",
			BackgroundJobID: sql.NullString{String: "job_123", Valid: true},
			ErrorMessage:    sql.NullString{String: "", Valid: false},
			ExpiresAt:       time.Now().Add(24 * time.Hour),
		}

		mock.ExpectExec(`INSERT INTO stored_responses`).
			WithArgs(
				resp.ResponseID,
				resp.UserID,
				resp.MerchantID,
				resp.Model,
				resp.Input,
				resp.Output,
				resp.ToolCalls,
				resp.Usage,
				resp.Status,
				resp.BackgroundJobID,
				resp.ErrorMessage,
				resp.ExpiresAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := svc.StoreResponse(ctx, resp)
		assert.NoError(t, err)
		assert.Equal(t, "resp_custom123", resp.ResponseID)
		assert.Equal(t, "queued", resp.Status)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("store with error message", func(t *testing.T) {
		resp := &models.StoredResponse{
			ResponseID:   "resp_err001",
			UserID:       1,
			MerchantID:   2,
			Model:        "gpt-4o",
			Input:        json.RawMessage(`"Hello"`),
			Status:       "failed",
			ErrorMessage: sql.NullString{String: "upstream timeout", Valid: true},
		}

		mock.ExpectExec(`INSERT INTO stored_responses`).
			WithArgs(
				resp.ResponseID,
				resp.UserID,
				resp.MerchantID,
				resp.Model,
				resp.Input,
				resp.Output,
				resp.ToolCalls,
				resp.Usage,
				resp.Status,
				resp.BackgroundJobID,
				resp.ErrorMessage,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := svc.StoreResponse(ctx, resp)
		assert.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestResponseStorageService_DeleteResponse(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mock.ExpectExec(`UPDATE stored_responses SET deleted_at = NOW`).
			WithArgs("resp_001").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := svc.DeleteResponse(ctx, "resp_001")
		assert.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete non-existent", func(t *testing.T) {
		mock.ExpectExec(`UPDATE stored_responses SET deleted_at = NOW`).
			WithArgs("resp_nonexist").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := svc.DeleteResponse(ctx, "resp_nonexist")
		assert.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestResponseStorageService_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE stored_responses SET status = `).
		WithArgs("running", "resp_001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.UpdateStatus(ctx, "resp_001", "running")
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_UpdateStatusWithError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	t.Run("update with error message", func(t *testing.T) {
		mock.ExpectExec(`UPDATE stored_responses SET status = .+, error_message = `).
			WithArgs("failed", "upstream timeout", "resp_001").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := svc.UpdateStatusWithError(ctx, "resp_001", "failed", "upstream timeout")
		assert.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update with internal error", func(t *testing.T) {
		mock.ExpectExec(`UPDATE stored_responses SET status = .+, error_message = `).
			WithArgs("failed", "internal error: nil pointer dereference", "resp_002").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := svc.UpdateStatusWithError(ctx, "resp_002", "failed", "internal error: nil pointer dereference")
		assert.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestResponseStorageService_UpdateOutput(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	output := json.RawMessage(`[{"type":"message","content":"Hello!"}]`)
	usage := json.RawMessage(`{"input_tokens":10,"output_tokens":5,"total_tokens":15}`)

	mock.ExpectExec(`UPDATE stored_responses SET output = .+, usage = .+, status = 'completed', error_message = NULL`).
		WithArgs(output, usage, "resp_001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = svc.UpdateOutput(ctx, "resp_001", output, usage)
	assert.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_CleanExpiredResponses(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	t.Run("clean expired responses", func(t *testing.T) {
		mock.ExpectExec(`UPDATE stored_responses SET deleted_at = NOW`).
			WillReturnResult(sqlmock.NewResult(0, 5))

		deleted, err := svc.CleanExpiredResponses(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), deleted)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no expired responses", func(t *testing.T) {
		mock.ExpectExec(`UPDATE stored_responses SET deleted_at = NOW`).
			WillReturnResult(sqlmock.NewResult(0, 0))

		deleted, err := svc.CleanExpiredResponses(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), deleted)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestResponseStorageService_StoreResponse_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	resp := &models.StoredResponse{
		UserID: 1,
		Model:  "gpt-4o",
		Input:  json.RawMessage(`"Hello"`),
	}

	mock.ExpectExec(`INSERT INTO stored_responses`).
		WillReturnError(sql.ErrConnDone)

	err = svc.StoreResponse(ctx, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store response")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_GetResponse_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM stored_responses`).
		WithArgs("resp_001").
		WillReturnError(sql.ErrConnDone)

	resp, err := svc.GetResponse(ctx, "resp_001")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get response")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_GetResponse_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM stored_responses WHERE response_id = .+ AND deleted_at IS NULL`).
		WithArgs("resp_nonexist").
		WillReturnError(sql.ErrNoRows)

	resp, err := svc.GetResponse(ctx, "resp_nonexist")
	assert.NoError(t, err)
	assert.Nil(t, resp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_GetByBackgroundJobID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM stored_responses WHERE background_job_id = `).
		WithArgs("job_nonexist").
		WillReturnError(sql.ErrNoRows)

	resp, err := svc.GetByBackgroundJobID(ctx, "job_nonexist")
	assert.NoError(t, err)
	assert.Nil(t, resp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_GetByBackgroundJobID_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT .+ FROM stored_responses WHERE background_job_id = `).
		WithArgs("job_001").
		WillReturnError(sql.ErrConnDone)

	resp, err := svc.GetByBackgroundJobID(ctx, "job_001")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get response by background job id")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_DeleteResponse_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE stored_responses SET deleted_at = NOW`).
		WithArgs("resp_001").
		WillReturnError(sql.ErrConnDone)

	err = svc.DeleteResponse(ctx, "resp_001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete response")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_UpdateStatus_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	mock.ExpectExec(`UPDATE stored_responses SET status = `).
		WithArgs("running", "resp_001").
		WillReturnError(sql.ErrConnDone)

	err = svc.UpdateStatus(ctx, "resp_001", "running")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update response status")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResponseStorageService_UpdateOutput_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	svc := NewResponseStorageService(db)
	ctx := context.Background()

	output := json.RawMessage(`[{"type":"message"}]`)
	usage := json.RawMessage(`{"total_tokens":15}`)

	mock.ExpectExec(`UPDATE stored_responses SET output`).
		WithArgs(output, usage, "resp_001").
		WillReturnError(sql.ErrConnDone)

	err = svc.UpdateOutput(ctx, "resp_001", output, usage)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update response output")
	require.NoError(t, mock.ExpectationsWereMet())
}
