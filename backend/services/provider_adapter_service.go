package services

import (
	"database/sql"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

type ProviderAdapterService struct {
	db *sql.DB
}

func NewProviderAdapterService() *ProviderAdapterService {
	return &ProviderAdapterService{
		db: config.DB,
	}
}

func (s *ProviderAdapterService) GetProviderByCode(code string) (*models.ModelProvider, error) {
	var provider models.ModelProvider
	err := s.db.QueryRow(
		"SELECT id, code, name, api_base_url, api_format, billing_type, cache_enabled, cache_discount, status, sort_order, created_at, updated_at FROM model_providers WHERE code = $1",
		code,
	).Scan(&provider.ID, &provider.Code, &provider.Name, &provider.APIBaseURL, &provider.APIFormat, &provider.BillingType, &provider.CacheEnabled, &provider.CacheDiscount, &provider.Status, &provider.SortOrder, &provider.CreatedAt, &provider.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, err
	}
	return &provider, nil
}
