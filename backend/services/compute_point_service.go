package services

import (
	"database/sql"
	"time"

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

func (s *ComputePointService) CreditComputePoints(userID int, points float64, description string) (*models.ComputePointAccount, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var account models.ComputePointAccount
	err = tx.QueryRow(
		"SELECT id, user_id, balance, total_earned, total_used, COALESCE(total_expired, 0) FROM compute_point_accounts WHERE user_id = $1 FOR UPDATE",
		userID,
	).Scan(&account.ID, &account.UserID, &account.Balance, &account.TotalEarned, &account.TotalUsed, &account.TotalExpired)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(
			"INSERT INTO compute_point_accounts (user_id, balance) VALUES ($1, 0)",
			userID,
		)
		if err != nil {
			return nil, err
		}
		account = models.ComputePointAccount{
			UserID:   userID,
			Balance:  0,
		}
	} else if err != nil {
		return nil, err
	}

	newBalance := account.Balance + points
	_, err = tx.Exec(
		"UPDATE compute_point_accounts SET balance = $1, total_earned = total_earned + $2 WHERE user_id = $3",
		newBalance, account.TotalEarned+points, account.UserID,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(
		`INSERT INTO compute_point_transactions (user_id, type, amount, balance_after, description, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, "purchase", points, newBalance, description, time.Now(),
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	account.Balance = newBalance
	account.TotalEarned += points
	return &account, nil
}
