package services

import (
	"database/sql"

	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/models"
)

type ComputePointService struct {
	db *sql.DB
}

func NewComputePointService() *ComputePointService {
	return &ComputePointService{
		db: config.DB,
	}
}

// CreditComputePoints credits the unified retail ledger (`tokens`) and records `token_transactions`.
// Legacy compute_point_accounts path removed (046 / IE-2).
func (s *ComputePointService) CreditComputePoints(userID int, points float64, description string) (*models.ComputePointAccount, error) {
	if s.db == nil {
		s.db = config.GetDB()
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := billing.CreditLegacyLot(tx, userID, points, "compute_points"); err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		`INSERT INTO token_transactions (user_id, type, amount, reason) VALUES ($1, 'purchase', $2, $3)`,
		userID, points, description,
	)
	if err != nil {
		return nil, err
	}

	var balance, totalEarned, totalUsed float64
	err = tx.QueryRow(
		`SELECT balance, total_earned, total_used FROM tokens WHERE user_id = $1`,
		userID,
	).Scan(&balance, &totalEarned, &totalUsed)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.ComputePointAccount{
		UserID:      userID,
		Balance:     balance,
		TotalEarned: totalEarned,
		TotalUsed:   totalUsed,
	}, nil
}
