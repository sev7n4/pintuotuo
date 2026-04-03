package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type FulfillmentService struct{}

func NewFulfillmentService() *FulfillmentService {
	return &FulfillmentService{}
}

type skuFulfillmentRow struct {
	SKUType            string
	TokenAmount        sql.NullInt64
	ComputePoints      sql.NullFloat64
	SubscriptionPeriod sql.NullString
	ValidDays          sql.NullInt64
	TrialDurationDays  sql.NullInt64
}

// FulfillOrder delivers digital goods for a paid order. Idempotent via orders.fulfilled_at.
func (s *FulfillmentService) FulfillOrder(tx *sql.Tx, orderID int) error {
	var userID int
	var skuID sql.NullInt64
	var qty int
	var status string
	var fulfilledAt sql.NullTime

	err := tx.QueryRow(`
		SELECT user_id, sku_id, quantity, status, fulfilled_at
		FROM orders WHERE id = $1 FOR UPDATE`,
		orderID,
	).Scan(&userID, &skuID, &qty, &status, &fulfilledAt)
	if err != nil {
		return fmt.Errorf("load order %d: %w", orderID, err)
	}

	if fulfilledAt.Valid {
		return nil
	}
	if status != "paid" {
		return fmt.Errorf("order %d not paid (status=%s)", orderID, status)
	}
	if !skuID.Valid || skuID.Int64 <= 0 {
		return fmt.Errorf("order %d has no sku_id", orderID)
	}
	if qty < 1 {
		qty = 1
	}

	var skuRow skuFulfillmentRow
	err = tx.QueryRow(`
		SELECT sku_type, token_amount, compute_points, subscription_period, valid_days, trial_duration_days
		FROM skus WHERE id = $1`,
		skuID.Int64,
	).Scan(
		&skuRow.SKUType,
		&skuRow.TokenAmount,
		&skuRow.ComputePoints,
		&skuRow.SubscriptionPeriod,
		&skuRow.ValidDays,
		&skuRow.TrialDurationDays,
	)
	if err != nil {
		return fmt.Errorf("load sku %d: %w", skuID.Int64, err)
	}

	st := strings.ToLower(strings.TrimSpace(skuRow.SKUType))
	switch st {
	case "token_pack":
		if err := s.fulfillTokenPack(tx, userID, int(skuID.Int64), orderID, qty, skuRow.TokenAmount); err != nil {
			return err
		}
	case "compute_points":
		if err := s.fulfillComputePoints(tx, userID, int(skuID.Int64), orderID, qty, skuRow.ComputePoints); err != nil {
			return err
		}
	case "subscription":
		if err := s.fulfillSubscription(tx, userID, int(skuID.Int64), orderID, qty, skuRow); err != nil {
			return err
		}
	case "trial":
		if err := s.fulfillTrial(tx, userID, int(skuID.Int64), orderID, qty, skuRow); err != nil {
			return err
		}
	case "concurrent":
		if skuRow.TokenAmount.Valid && skuRow.TokenAmount.Int64 > 0 {
			if err := s.fulfillTokenPack(tx, userID, int(skuID.Int64), orderID, qty, skuRow.TokenAmount); err != nil {
				return err
			}
		} else if skuRow.ComputePoints.Valid && skuRow.ComputePoints.Float64 > 0 {
			if err := s.fulfillComputePoints(tx, userID, int(skuID.Int64), orderID, qty, skuRow.ComputePoints); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown sku_type %q", skuRow.SKUType)
	}

	_, err = tx.Exec(`UPDATE orders SET fulfilled_at = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, time.Now(), orderID)
	return err
}

func (s *FulfillmentService) fulfillTokenPack(tx *sql.Tx, userID, skuID, orderID, qty int, tokenAmount sql.NullInt64) error {
	if !tokenAmount.Valid || tokenAmount.Int64 <= 0 {
		return fmt.Errorf("token_pack sku %d has no token_amount", skuID)
	}
	add := float64(tokenAmount.Int64) * float64(qty)
	_, err := tx.Exec(`
		INSERT INTO tokens (user_id, balance, total_used, total_earned)
		VALUES ($1, $2, 0, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			balance = tokens.balance + EXCLUDED.balance,
			total_earned = tokens.total_earned + EXCLUDED.balance,
			updated_at = NOW()`,
		userID, add,
	)
	if err != nil {
		return fmt.Errorf("token_pack credit: %w", err)
	}
	_, err = tx.Exec(
		`INSERT INTO token_transactions (user_id, type, amount, reason, order_id) VALUES ($1, 'purchase', $2, $3, $4)`,
		userID, add, fmt.Sprintf("订单商品 #%d", orderID), orderID,
	)
	return err
}

func (s *FulfillmentService) fulfillComputePoints(tx *sql.Tx, userID, skuID, orderID, qty int, computePoints sql.NullFloat64) error {
	if !computePoints.Valid || computePoints.Float64 <= 0 {
		return fmt.Errorf("compute_points sku %d has no compute_points", skuID)
	}
	points := computePoints.Float64 * float64(qty)

	var balance float64
	err := tx.QueryRow(
		`SELECT balance FROM compute_point_accounts WHERE user_id = $1 FOR UPDATE`,
		userID,
	).Scan(&balance)
	if err == sql.ErrNoRows {
		_, err = tx.Exec(`INSERT INTO compute_point_accounts (user_id, balance) VALUES ($1, 0)`, userID)
		if err != nil {
			return fmt.Errorf("compute account create: %w", err)
		}
		balance = 0
	} else if err != nil {
		return fmt.Errorf("compute account load: %w", err)
	}

	newBal := balance + points
	_, err = tx.Exec(
		`UPDATE compute_point_accounts SET balance = $1, total_earned = total_earned + $2, updated_at = NOW() WHERE user_id = $3`,
		newBal, points, userID,
	)
	if err != nil {
		return fmt.Errorf("compute credit: %w", err)
	}
	_, err = tx.Exec(
		`INSERT INTO compute_point_transactions (user_id, type, amount, balance_after, order_id, sku_id, description)
		 VALUES ($1, 'purchase', $2, $3, $4, $5, $6)`,
		userID, points, newBal, orderID, skuID, fmt.Sprintf("订单 #%d", orderID),
	)
	return err
}

func (s *FulfillmentService) fulfillSubscription(tx *sql.Tx, userID, skuID, orderID, qty int, skuRow skuFulfillmentRow) error {
	period := ""
	if skuRow.SubscriptionPeriod.Valid {
		period = skuRow.SubscriptionPeriod.String
	}
	validDays := 0
	if skuRow.ValidDays.Valid {
		validDays = int(skuRow.ValidDays.Int64)
	}
	return s.upsertSubscription(tx, userID, skuID, orderID, qty, period, validDays, false)
}

func (s *FulfillmentService) fulfillTrial(tx *sql.Tx, userID, skuID, orderID, qty int, skuRow skuFulfillmentRow) error {
	days := 7
	if skuRow.TrialDurationDays.Valid && skuRow.TrialDurationDays.Int64 > 0 {
		days = int(skuRow.TrialDurationDays.Int64)
	} else if skuRow.ValidDays.Valid && skuRow.ValidDays.Int64 > 0 {
		days = int(skuRow.ValidDays.Int64)
	}
	_ = orderID
	today := time.Now().UTC().Truncate(24 * time.Hour)
	end := today.AddDate(0, 0, days*qty)
	var subID int
	err := tx.QueryRow(
		`SELECT id FROM user_subscriptions WHERE user_id = $1 AND sku_id = $2 AND status = 'active' FOR UPDATE`,
		userID, skuID,
	).Scan(&subID)
	if err == sql.ErrNoRows {
		return s.insertSubscription(tx, userID, skuID, today, end, false)
	}
	if err != nil {
		return err
	}
	var endDate time.Time
	err = tx.QueryRow(`SELECT end_date FROM user_subscriptions WHERE id = $1`, subID).Scan(&endDate)
	if err != nil {
		return err
	}
	base := endDate.UTC().Truncate(24 * time.Hour)
	if base.Before(today) {
		base = today
	}
	newEnd := base.AddDate(0, 0, days*qty)
	_, err = tx.Exec(`UPDATE user_subscriptions SET end_date = $1::date, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, newEnd.Format("2006-01-02"), subID)
	return err
}

func (s *FulfillmentService) upsertSubscription(tx *sql.Tx, userID, skuID, orderID, qty int, period string, validDays int, autoRenew bool) error {
	_ = orderID
	var subID int
	var endDate time.Time
	err := tx.QueryRow(
		`SELECT id, end_date FROM user_subscriptions WHERE user_id = $1 AND sku_id = $2 AND status = 'active' FOR UPDATE`,
		userID, skuID,
	).Scan(&subID, &endDate)
	if err == sql.ErrNoRows {
		start := time.Now().UTC().Truncate(24 * time.Hour)
		end := stackSubscriptionPeriods(start, period, validDays, qty)
		return s.insertSubscription(tx, userID, skuID, start, end, autoRenew)
	}
	if err != nil {
		return err
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	base := endDate.UTC().Truncate(24 * time.Hour)
	if base.Before(today) {
		base = today
	}
	newEnd := stackSubscriptionPeriods(base, period, validDays, qty)
	_, err = tx.Exec(
		`UPDATE user_subscriptions SET end_date = $1::date, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		newEnd.Format("2006-01-02"), subID,
	)
	return err
}

func stackSubscriptionPeriods(base time.Time, period string, validDays int, qty int) time.Time {
	t := base
	for i := 0; i < qty; i++ {
		t = CalculateSubscriptionEndFrom(t, period, validDays, 0, "subscription")
	}
	return t
}

func (s *FulfillmentService) insertSubscription(tx *sql.Tx, userID, skuID int, start, end time.Time, autoRenew bool) error {
	_, err := tx.Exec(
		`INSERT INTO user_subscriptions (user_id, sku_id, start_date, end_date, status, auto_renew)
		 VALUES ($1, $2, $3::date, $4::date, 'active', $5)`,
		userID, skuID, start.Format("2006-01-02"), end.Format("2006-01-02"), autoRenew,
	)
	return err
}

// CalculateSubscriptionEndFrom computes subscription end date (one period) from a base date.
func CalculateSubscriptionEndFrom(base time.Time, period string, validDays int, trialDays int, skuType string) time.Time {
	base = base.UTC().Truncate(24 * time.Hour)
	st := strings.ToLower(strings.TrimSpace(skuType))
	if st == "trial" {
		d := trialDays
		if d <= 0 {
			d = validDays
		}
		if d <= 0 {
			d = 7
		}
		return base.AddDate(0, 0, d)
	}
	p := strings.ToLower(strings.TrimSpace(period))
	switch p {
	case "monthly":
		return base.AddDate(0, 1, 0)
	case "quarterly":
		return base.AddDate(0, 3, 0)
	case "yearly":
		return base.AddDate(1, 0, 0)
	case "":
		if validDays > 0 {
			return base.AddDate(0, 0, validDays)
		}
		return base.AddDate(0, 1, 0)
	default:
		if validDays > 0 {
			return base.AddDate(0, 0, validDays)
		}
		return base.AddDate(0, 1, 0)
	}
}

// ValidateSKUForOrder checks SKU fields required for checkout (create order).
func ValidateSKUForOrder(skuType string, tokenAmount int64, computePoints float64, subscriptionPeriod string, validDays int, trialDurationDays int) error {
	st := strings.ToLower(strings.TrimSpace(skuType))
	switch st {
	case "token_pack":
		if tokenAmount <= 0 {
			return fmt.Errorf("token_pack requires token_amount > 0")
		}
	case "compute_points":
		if computePoints <= 0 {
			return fmt.Errorf("compute_points requires compute_points > 0")
		}
	case "subscription":
		if strings.TrimSpace(subscriptionPeriod) != "" {
			p := strings.ToLower(subscriptionPeriod)
			if p != "monthly" && p != "quarterly" && p != "yearly" {
				return fmt.Errorf("subscription has invalid subscription_period")
			}
			return nil
		}
		if validDays <= 0 {
			return fmt.Errorf("subscription requires subscription_period or valid_days > 0")
		}
	case "trial":
		if trialDurationDays <= 0 && validDays <= 0 {
			return fmt.Errorf("trial requires trial_duration_days or valid_days > 0")
		}
	case "concurrent":
		if tokenAmount <= 0 && computePoints <= 0 {
			return fmt.Errorf("concurrent requires token_amount or compute_points > 0")
		}
	default:
		return fmt.Errorf("unknown sku_type %q", skuType)
	}
	return nil
}
