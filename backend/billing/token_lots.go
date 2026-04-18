package billing

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// DefaultTokenLotValidDays is the default validity for token_pack / subscription bonus credits (加油包 PRD).
const DefaultTokenLotValidDays = 365

func round2(x float64) float64 {
	return math.Round(x*100) / 100
}

// ForfeitExpiredLots zeros expired lots and reduces tokens.balance; records one aggregate token_transactions row per forfeit batch.
func ForfeitExpiredLots(tx *sql.Tx, userID int) error {
	rows, err := tx.Query(`
		SELECT id, remaining_amount
		  FROM token_lots
		 WHERE user_id = $1
		   AND remaining_amount > 0
		   AND expires_at IS NOT NULL
		   AND expires_at < NOW()
		 FOR UPDATE`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("forfeit select lots: %w", err)
	}

	var total float64
	var ids []int
	for rows.Next() {
		var id int
		var rem float64
		if err = rows.Scan(&id, &rem); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan lot: %w", err)
		}
		total += rem
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return fmt.Errorf("close forfeit rows: %w", err)
	}
	if len(ids) == 0 {
		return nil
	}
	total = round2(total)
	if total <= 0 {
		return nil
	}

	for _, id := range ids {
		if _, err = tx.Exec(`UPDATE token_lots SET remaining_amount = 0 WHERE id = $1`, id); err != nil {
			return fmt.Errorf("zero expired lot: %w", err)
		}
	}

	if _, err = tx.Exec(
		`UPDATE tokens SET balance = balance - $1, updated_at = $2 WHERE user_id = $3`,
		total, time.Now(), userID,
	); err != nil {
		return fmt.Errorf("forfeit reduce balance: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO token_transactions (user_id, type, amount, reason) VALUES ($1, 'expired', $2, $3)`,
		userID, -total, "Token lot expired",
	)
	if err != nil {
		return fmt.Errorf("forfeit tx log: %w", err)
	}
	return nil
}

// DebitLotsFIFO removes amount from active lots (earliest expiry first; NULL expiry last). Updates tokens.balance; optionally total_used.
func DebitLotsFIFO(tx *sql.Tx, userID int, amount float64, updateTotalUsed bool) error {
	amount = round2(amount)
	if amount <= 0 {
		return nil
	}

	if err := ForfeitExpiredLots(tx, userID); err != nil {
		return err
	}

	rows, err := tx.Query(`
		SELECT id, remaining_amount
		  FROM token_lots
		 WHERE user_id = $1
		   AND remaining_amount > 0
		   AND (expires_at IS NULL OR expires_at > NOW())
		 ORDER BY expires_at ASC NULLS LAST, id ASC
		 FOR UPDATE`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("debit select lots: %w", err)
	}

	type lotRow struct {
		id  int
		rem float64
	}
	var lots []lotRow
	for rows.Next() {
		var r lotRow
		if err = rows.Scan(&r.id, &r.rem); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan lot: %w", err)
		}
		lots = append(lots, r)
	}
	if err = rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return fmt.Errorf("close debit rows: %w", err)
	}

	need := amount
	for _, r := range lots {
		if need <= 0.000001 {
			break
		}
		id, rem := r.id, r.rem
		take := rem
		if take > need {
			take = need
		}
		take = round2(take)
		newRem := round2(rem - take)
		if _, err = tx.Exec(`UPDATE token_lots SET remaining_amount = $1 WHERE id = $2`, newRem, id); err != nil {
			return fmt.Errorf("update lot: %w", err)
		}
		need = round2(need - take)
	}

	if need > 0.01 {
		return fmt.Errorf("insufficient balance in lots: short by %.2f", need)
	}

	if updateTotalUsed {
		if _, err = tx.Exec(
			`UPDATE tokens SET balance = balance - $1, total_used = total_used + $1, updated_at = $2 WHERE user_id = $3`,
			amount, time.Now(), userID,
		); err != nil {
			return fmt.Errorf("tokens deduct+used: %w", err)
		}
	} else {
		if _, err = tx.Exec(
			`UPDATE tokens SET balance = balance - $1, updated_at = $2 WHERE user_id = $3`,
			amount, time.Now(), userID,
		); err != nil {
			return fmt.Errorf("tokens deduct: %w", err)
		}
	}
	return nil
}

// CreditTokenLot inserts a lot row and increases tokens.balance / total_earned.
func CreditTokenLot(tx *sql.Tx, userID int, amount float64, expiresAt *time.Time, orderItemID int, lotType string) error {
	amount = round2(amount)
	if amount <= 0 {
		return nil
	}
	var exp interface{}
	if expiresAt != nil {
		exp = *expiresAt
	}
	var oi interface{}
	if orderItemID > 0 {
		oi = orderItemID
	}
	if _, err := tx.Exec(`
		INSERT INTO token_lots (user_id, remaining_amount, expires_at, order_item_id, lot_type)
		VALUES ($1, $2, $3, $4, $5)`,
		userID, amount, exp, oi, lotType,
	); err != nil {
		return fmt.Errorf("insert lot: %w", err)
	}

	_, err := tx.Exec(`
		INSERT INTO tokens (user_id, balance, total_used, total_earned)
		VALUES ($1, $2, 0, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			balance = tokens.balance + EXCLUDED.balance,
			total_earned = tokens.total_earned + EXCLUDED.balance,
			updated_at = NOW()`,
		userID, amount,
	)
	if err != nil {
		return fmt.Errorf("credit tokens: %w", err)
	}
	return nil
}

// CreditLegacyLot adds balance as a never-expiring lot (recharge, refunds, pre-deduct cancel).
func CreditLegacyLot(tx *sql.Tx, userID int, amount float64, lotType string) error {
	return CreditTokenLot(tx, userID, amount, nil, 0, lotType)
}
