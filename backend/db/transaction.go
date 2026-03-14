package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Transaction wraps database transactions with automatic rollback
type Transaction struct {
	tx *sql.Tx
	ctx context.Context
}

// BeginTx starts a new database transaction
func BeginTx(db *sql.DB, ctx context.Context) (*Transaction, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{
		tx:  tx,
		ctx: ctx,
	}, nil
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(t.ctx, query, args...)
}

// Query queries rows within the transaction
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, query, args...)
}

// QueryRow queries a single row within the transaction
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(t.ctx, query, args...)
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

// DoInTransaction wraps code execution in a transaction with automatic rollback
func DoInTransaction(db *sql.DB, ctx context.Context, fn func(*Transaction) error) error {
	tx, err := BeginTx(db, ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		// Attempt rollback, but don't override the original error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w (rollback error: %w)", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// PaymentTransaction handles payment processing with proper transaction semantics
func PaymentTransaction(db *sql.DB, ctx context.Context,
	paymentID int,
	orderID int,
	userID int,
	amount float64,
	fn func(*Transaction) error) error {

	return DoInTransaction(db, ctx, func(tx *Transaction) error {
		// Update payment status
		_, err := tx.Exec(
			"UPDATE payments SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
			"success", paymentID,
		)
		if err != nil {
			return fmt.Errorf("failed to update payment: %w", err)
		}

		// Update order status
		_, err = tx.Exec(
			"UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
			"paid", orderID,
		)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		// Update token balance if needed
		_, err = tx.Exec(
			"UPDATE tokens SET total_used = total_used + $1 WHERE user_id = $2",
			amount, userID,
		)
		if err != nil {
			return fmt.Errorf("failed to update token balance: %w", err)
		}

		// Log transaction
		_, err = tx.Exec(
			"INSERT INTO token_transactions (user_id, type, amount, reason, order_id) VALUES ($1, $2, $3, $4, $5)",
			userID, "use", amount, "Payment for order", orderID,
		)
		if err != nil {
			return fmt.Errorf("failed to log transaction: %w", err)
		}

		// Execute custom logic if provided
		if fn != nil {
			if err := fn(tx); err != nil {
				return err
			}
		}

		return nil
	})
}

// GroupPurchaseTransaction handles group purchase completion
func GroupPurchaseTransaction(db *sql.DB, ctx context.Context,
	groupID int,
	fn func(*Transaction) error) error {

	return DoInTransaction(db, ctx, func(tx *Transaction) error {
		// Update group status
		_, err := tx.Exec(
			"UPDATE groups SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
			"completed", groupID,
		)
		if err != nil {
			return fmt.Errorf("failed to update group: %w", err)
		}

		// Execute custom logic if provided
		if fn != nil {
			if err := fn(tx); err != nil {
				return err
			}
		}

		return nil
	})
}
