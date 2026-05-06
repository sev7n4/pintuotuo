package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pintuotuo/backend/models"

	"github.com/google/uuid"
)

type ResponseStorageService struct {
	db *sql.DB
}

func NewResponseStorageService(db *sql.DB) *ResponseStorageService {
	return &ResponseStorageService{db: db}
}

func (s *ResponseStorageService) StoreResponse(ctx context.Context, resp *models.StoredResponse) error {
	if resp.ResponseID == "" {
		resp.ResponseID = "resp_" + uuid.New().String()[:24]
	}
	if resp.Status == "" {
		resp.Status = "completed"
	}
	if resp.ExpiresAt.IsZero() {
		resp.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO stored_responses (response_id, user_id, merchant_id, model, input, output, tool_calls, usage, status, background_job_id, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		resp.ResponseID, resp.UserID, resp.MerchantID, resp.Model,
		resp.Input, resp.Output, resp.ToolCalls, resp.Usage,
		resp.Status, resp.BackgroundJobID, resp.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store response: %w", err)
	}
	return nil
}

func (s *ResponseStorageService) GetResponse(ctx context.Context, responseID string) (*models.StoredResponse, error) {
	resp := &models.StoredResponse{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, response_id, user_id, merchant_id, model, input, output, tool_calls, usage, status, background_job_id, created_at, expires_at, deleted_at
		 FROM stored_responses WHERE response_id = $1 AND deleted_at IS NULL`,
		responseID,
	).Scan(&resp.ID, &resp.ResponseID, &resp.UserID, &resp.MerchantID, &resp.Model,
		&resp.Input, &resp.Output, &resp.ToolCalls, &resp.Usage, &resp.Status,
		&resp.BackgroundJobID, &resp.CreatedAt, &resp.ExpiresAt, &resp.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	return resp, nil
}

func (s *ResponseStorageService) DeleteResponse(ctx context.Context, responseID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE stored_responses SET deleted_at = NOW() WHERE response_id = $1 AND deleted_at IS NULL`,
		responseID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete response: %w", err)
	}
	return nil
}

func (s *ResponseStorageService) UpdateStatus(ctx context.Context, responseID, status string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE stored_responses SET status = $1 WHERE response_id = $2`,
		status, responseID,
	)
	if err != nil {
		return fmt.Errorf("failed to update response status: %w", err)
	}
	return nil
}

func (s *ResponseStorageService) UpdateOutput(ctx context.Context, responseID string, output, usage []byte) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE stored_responses SET output = $1, usage = $2, status = 'completed' WHERE response_id = $3`,
		output, usage, responseID,
	)
	if err != nil {
		return fmt.Errorf("failed to update response output: %w", err)
	}
	return nil
}

func (s *ResponseStorageService) GetByBackgroundJobID(ctx context.Context, jobID string) (*models.StoredResponse, error) {
	resp := &models.StoredResponse{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, response_id, user_id, merchant_id, model, input, output, tool_calls, usage, status, background_job_id, created_at, expires_at, deleted_at
		 FROM stored_responses WHERE background_job_id = $1`,
		jobID,
	).Scan(&resp.ID, &resp.ResponseID, &resp.UserID, &resp.MerchantID, &resp.Model,
		&resp.Input, &resp.Output, &resp.ToolCalls, &resp.Usage, &resp.Status,
		&resp.BackgroundJobID, &resp.CreatedAt, &resp.ExpiresAt, &resp.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get response by background job id: %w", err)
	}
	return resp, nil
}

func (s *ResponseStorageService) CleanExpiredResponses(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM stored_responses WHERE expires_at < NOW() AND deleted_at IS NULL`,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to clean expired responses: %w", err)
	}
	return result.RowsAffected()
}
