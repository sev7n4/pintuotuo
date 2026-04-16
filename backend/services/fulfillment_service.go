package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pintuotuo/backend/billing"
)

const (
	skuTypeTokenPack     = "token_pack"
	skuTypeComputePoints = "compute_points"
	skuTypeSubscription  = "subscription"
	skuTypeTrial         = "trial"
	skuTypeConcurrent    = "concurrent"
)

type FulfillmentService struct{}

func NewFulfillmentService() *FulfillmentService {
	return &FulfillmentService{}
}

// FulfillOrderItems fulfills all items in one order transactionally.
func (s *FulfillmentService) FulfillOrderItems(tx *sql.Tx, orderID int) error {
	return s.FulfillOrder(tx, orderID)
}

type skuFulfillmentRow struct {
	SKUType            string
	TokenAmount        sql.NullInt64
	ComputePoints      sql.NullFloat64
	SubscriptionPeriod sql.NullString
	ValidDays          sql.NullInt64
	TrialDurationDays  sql.NullInt64
}

// EffectiveSKUTypeForFulfillment prefers the order snapshot (checkout-time) over live SKU rows.
func EffectiveSKUTypeForFulfillment(orderType sql.NullString, skuType string) string {
	if orderType.Valid {
		s := strings.TrimSpace(orderType.String)
		if s != "" {
			return s
		}
	}
	return skuType
}

// EffectiveTokenAmountForFulfillment prefers orders.token_amount when set so fulfillment matches what was sold.
func EffectiveTokenAmountForFulfillment(orderTA, skuTA sql.NullInt64) sql.NullInt64 {
	if orderTA.Valid && orderTA.Int64 > 0 {
		return orderTA
	}
	return skuTA
}

// EffectiveComputePointsForFulfillment prefers orders.compute_points when set.
func EffectiveComputePointsForFulfillment(orderCP, skuCP sql.NullFloat64) sql.NullFloat64 {
	if orderCP.Valid && orderCP.Float64 > 0 {
		return orderCP
	}
	return skuCP
}

// FulfillOrder delivers digital goods for a paid order. Idempotent via orders.fulfilled_at.
func (s *FulfillmentService) FulfillOrder(tx *sql.Tx, orderID int) error {
	var userID int
	var status string
	var fulfilledAt sql.NullTime

	err := tx.QueryRow(`
		SELECT user_id, status, fulfilled_at
		FROM orders WHERE id = $1 FOR UPDATE`,
		orderID,
	).Scan(&userID, &status, &fulfilledAt)
	if err != nil {
		return fmt.Errorf("load order %d: %w", orderID, err)
	}

	if fulfilledAt.Valid {
		return nil
	}
	if status != "paid" {
		return fmt.Errorf("order %d not paid (status=%s)", orderID, status)
	}

	rows, err := tx.Query(`
		SELECT id, sku_id, quantity, sku_type, token_amount, compute_points, pricing_version_id, fulfilled_at
		  FROM order_items
		 WHERE order_id = $1
		 ORDER BY id ASC
		 FOR UPDATE`,
		orderID,
	)
	if err != nil {
		return fmt.Errorf("load order_items for order %d: %w", orderID, err)
	}
	defer rows.Close()

	type orderItemRow struct {
		ID               int
		SKUID            int
		Quantity         int
		SKUType          sql.NullString
		TokenAmount      sql.NullInt64
		ComputePoints    sql.NullFloat64
		PricingVersionID sql.NullInt64
		FulfilledAt      sql.NullTime
	}
	items := make([]orderItemRow, 0)
	for rows.Next() {
		var item orderItemRow
		if scanErr := rows.Scan(
			&item.ID, &item.SKUID, &item.Quantity, &item.SKUType, &item.TokenAmount, &item.ComputePoints, &item.PricingVersionID, &item.FulfilledAt,
		); scanErr != nil {
			return fmt.Errorf("scan order_item for order %d: %w", orderID, scanErr)
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return fmt.Errorf("order %d has no order_items", orderID)
	}

	for _, item := range items {
		if item.FulfilledAt.Valid {
			continue
		}
		if item.Quantity < 1 {
			item.Quantity = 1
		}

		var skuRow skuFulfillmentRow
		err = tx.QueryRow(`
			SELECT sku_type, token_amount, compute_points, subscription_period, valid_days, trial_duration_days
			FROM skus WHERE id = $1`,
			item.SKUID,
		).Scan(
			&skuRow.SKUType,
			&skuRow.TokenAmount,
			&skuRow.ComputePoints,
			&skuRow.SubscriptionPeriod,
			&skuRow.ValidDays,
			&skuRow.TrialDurationDays,
		)
		if err != nil {
			return fmt.Errorf("load sku %d: %w", item.SKUID, err)
		}

		effectiveType := EffectiveSKUTypeForFulfillment(item.SKUType, skuRow.SKUType)
		tokenAmt := EffectiveTokenAmountForFulfillment(item.TokenAmount, skuRow.TokenAmount)
		computeAmt := EffectiveComputePointsForFulfillment(item.ComputePoints, skuRow.ComputePoints)

		st := strings.ToLower(strings.TrimSpace(effectiveType))
		switch st {
		case skuTypeTokenPack:
			if err = s.fulfillTokenPack(tx, userID, item.SKUID, orderID, item.ID, item.Quantity, tokenAmt, "token_pack"); err != nil {
				return err
			}
		case skuTypeComputePoints:
			if err = s.fulfillComputePoints(tx, userID, item.SKUID, orderID, item.ID, item.Quantity, computeAmt); err != nil {
				return err
			}
		case skuTypeSubscription:
			if err = s.fulfillSubscription(tx, userID, item.SKUID, orderID, item.ID, item.Quantity, skuRow, tokenAmt); err != nil {
				return err
			}
		case skuTypeTrial:
			if err = s.fulfillTrial(tx, userID, item.SKUID, orderID, item.Quantity, skuRow); err != nil {
				return err
			}
		case skuTypeConcurrent:
			if tokenAmt.Valid && tokenAmt.Int64 > 0 {
				if err = s.fulfillTokenPack(tx, userID, item.SKUID, orderID, item.ID, item.Quantity, tokenAmt, "token_pack"); err != nil {
					return err
				}
			} else if computeAmt.Valid && computeAmt.Float64 > 0 {
				if err = s.fulfillComputePoints(tx, userID, item.SKUID, orderID, item.ID, item.Quantity, computeAmt); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unknown sku_type %q", effectiveType)
		}

		if _, err = tx.Exec(`UPDATE order_items SET fulfilled_at = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, time.Now(), item.ID); err != nil {
			return fmt.Errorf("mark order_item fulfilled: %w", err)
		}
	}

	_, err = tx.Exec(`UPDATE orders SET fulfilled_at = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, time.Now(), orderID)
	return err
}

// fulfillTokenPack credits a dated token lot (default 365d). lotType examples: token_pack, subscription_bonus.
func (s *FulfillmentService) fulfillTokenPack(tx *sql.Tx, userID, skuID, orderID, orderItemID, qty int, tokenAmount sql.NullInt64, lotType string) error {
	if !tokenAmount.Valid || tokenAmount.Int64 <= 0 {
		return fmt.Errorf("token_pack sku %d has no token_amount", skuID)
	}
	if lotType == "" {
		lotType = "token_pack"
	}
	add := float64(tokenAmount.Int64) * float64(qty)
	exp := time.Now().UTC().AddDate(0, 0, billing.DefaultTokenLotValidDays)
	if err := billing.CreditTokenLot(tx, userID, add, &exp, orderItemID, lotType); err != nil {
		return fmt.Errorf("token_pack credit: %w", err)
	}
	_, err := tx.Exec(
		`INSERT INTO token_transactions (user_id, type, amount, reason, order_id) VALUES ($1, 'purchase', $2, $3, $4)`,
		userID, add, fmt.Sprintf("订单商品 #%d", orderID), orderID,
	)
	return err
}

func (s *FulfillmentService) fulfillComputePoints(tx *sql.Tx, userID, skuID, orderID, orderItemID, qty int, computePoints sql.NullFloat64) error {
	if !computePoints.Valid || computePoints.Float64 <= 0 {
		return fmt.Errorf("compute_points sku %d has no compute_points", skuID)
	}
	add := computePoints.Float64 * float64(qty)
	// No expiry: same pool semantics as legacy recharge (可后续改为与 token_pack 一致).
	if err := billing.CreditTokenLot(tx, userID, add, nil, orderItemID, "compute_points"); err != nil {
		return fmt.Errorf("compute_points credit (tokens ledger): %w", err)
	}
	_, err := tx.Exec(
		`INSERT INTO token_transactions (user_id, type, amount, reason, order_id) VALUES ($1, 'purchase', $2, $3, $4)`,
		userID, add, fmt.Sprintf("订单商品 #%d (compute_points SKU)", orderID), orderID,
	)
	return err
}

func (s *FulfillmentService) fulfillSubscription(tx *sql.Tx, userID, skuID, orderID, orderItemID, qty int, skuRow skuFulfillmentRow, tokenAmount sql.NullInt64) error {
	period := ""
	if skuRow.SubscriptionPeriod.Valid {
		period = skuRow.SubscriptionPeriod.String
	}
	validDays := 0
	if skuRow.ValidDays.Valid {
		validDays = int(skuRow.ValidDays.Int64)
	}
	if err := s.upsertSubscription(tx, userID, skuID, orderID, qty, period, validDays, false); err != nil {
		return err
	}

	// Scheme 2: subscription SKU can also grant token balance at activation time.
	if tokenAmount.Valid && tokenAmount.Int64 > 0 {
		return s.fulfillTokenPack(tx, userID, skuID, orderID, orderItemID, qty, tokenAmount, "subscription_bonus")
	}
	return nil
}

// subscriptionPVAndAnchor binds pricing snapshot from the paid order (baseline if unset).
func subscriptionPVAndAnchor(tx *sql.Tx, orderID int) (sql.NullInt64, time.Time, error) {
	var pv sql.NullInt64
	err := tx.QueryRow(`SELECT pricing_version_id FROM orders WHERE id = $1`, orderID).Scan(&pv)
	if err != nil {
		return sql.NullInt64{}, time.Time{}, err
	}
	if !pv.Valid {
		pv = BaselinePricingVersionID(tx)
	}
	return pv, time.Now().UTC(), nil
}

func (s *FulfillmentService) fulfillTrial(tx *sql.Tx, userID, skuID, orderID, qty int, skuRow skuFulfillmentRow) error {
	days := 7
	if skuRow.TrialDurationDays.Valid && skuRow.TrialDurationDays.Int64 > 0 {
		days = int(skuRow.TrialDurationDays.Int64)
	} else if skuRow.ValidDays.Valid && skuRow.ValidDays.Int64 > 0 {
		days = int(skuRow.ValidDays.Int64)
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	end := today.AddDate(0, 0, days*qty)
	var subID int
	err := tx.QueryRow(
		`SELECT id FROM user_subscriptions WHERE user_id = $1 AND sku_id = $2 AND status = 'active' FOR UPDATE`,
		userID, skuID,
	).Scan(&subID)
	if err == sql.ErrNoRows {
		pv, anchor, perr := subscriptionPVAndAnchor(tx, orderID)
		if perr != nil {
			return perr
		}
		return s.insertSubscription(tx, userID, skuID, today, end, false, pv, anchor)
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
	pv, anchor, perr := subscriptionPVAndAnchor(tx, orderID)
	if perr != nil {
		return perr
	}
	var pvArg interface{}
	if pv.Valid {
		pvArg = pv.Int64
	}
	_, err = tx.Exec(
		`UPDATE user_subscriptions SET end_date = $1::date, pricing_version_id = $2, entitlement_anchor_at = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4`,
		newEnd.Format("2006-01-02"), pvArg, anchor, subID,
	)
	return err
}

func (s *FulfillmentService) upsertSubscription(tx *sql.Tx, userID, skuID, orderID, qty int, period string, validDays int, autoRenew bool) error {
	pv, anchor, err := subscriptionPVAndAnchor(tx, orderID)
	if err != nil {
		return err
	}
	var subID int
	var endDate time.Time
	err = tx.QueryRow(
		`SELECT id, end_date FROM user_subscriptions WHERE user_id = $1 AND sku_id = $2 AND status = 'active' FOR UPDATE`,
		userID, skuID,
	).Scan(&subID, &endDate)
	if err == sql.ErrNoRows {
		start := time.Now().UTC().Truncate(24 * time.Hour)
		end := StackSubscriptionPeriods(start, period, validDays, qty)
		return s.insertSubscription(tx, userID, skuID, start, end, autoRenew, pv, anchor)
	}
	if err != nil {
		return err
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	base := endDate.UTC().Truncate(24 * time.Hour)
	if base.Before(today) {
		base = today
	}
	newEnd := StackSubscriptionPeriods(base, period, validDays, qty)
	var pvArg interface{}
	if pv.Valid {
		pvArg = pv.Int64
	}
	_, err = tx.Exec(
		`UPDATE user_subscriptions SET end_date = $1::date, pricing_version_id = $2, entitlement_anchor_at = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4`,
		newEnd.Format("2006-01-02"), pvArg, anchor, subID,
	)
	return err
}

// StackSubscriptionPeriods extends end date by qty billing periods from base (subscription SKUs).
func StackSubscriptionPeriods(base time.Time, period string, validDays int, qty int) time.Time {
	t := base
	for i := 0; i < qty; i++ {
		t = CalculateSubscriptionEndFrom(t, period, validDays, 0, skuTypeSubscription)
	}
	return t
}

func (s *FulfillmentService) insertSubscription(tx *sql.Tx, userID, skuID int, start, end time.Time, autoRenew bool, pv sql.NullInt64, anchor time.Time) error {
	var pvArg interface{}
	if pv.Valid {
		pvArg = pv.Int64
	}
	_, err := tx.Exec(
		`INSERT INTO user_subscriptions (user_id, sku_id, start_date, end_date, status, auto_renew, pricing_version_id, entitlement_anchor_at)
		 VALUES ($1, $2, $3::date, $4::date, 'active', $5, $6, $7)`,
		userID, skuID, start.Format("2006-01-02"), end.Format("2006-01-02"), autoRenew, pvArg, anchor,
	)
	return err
}

// CalculateSubscriptionEndFrom computes subscription end date (one period) from a base date.
func CalculateSubscriptionEndFrom(base time.Time, period string, validDays int, trialDays int, skuType string) time.Time {
	base = base.UTC().Truncate(24 * time.Hour)
	st := strings.ToLower(strings.TrimSpace(skuType))
	if st == skuTypeTrial {
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
	case skuTypeTokenPack:
		if tokenAmount <= 0 {
			return fmt.Errorf("token_pack requires token_amount > 0")
		}
	case skuTypeComputePoints:
		if computePoints <= 0 {
			return fmt.Errorf("compute_points requires compute_points > 0")
		}
	case skuTypeSubscription:
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
	case skuTypeTrial:
		if trialDurationDays <= 0 && validDays <= 0 {
			return fmt.Errorf("trial requires trial_duration_days or valid_days > 0")
		}
	case skuTypeConcurrent:
		if tokenAmount <= 0 && computePoints <= 0 {
			return fmt.Errorf("concurrent requires token_amount or compute_points > 0")
		}
	default:
		return fmt.Errorf("unknown sku_type %q", skuType)
	}
	return nil
}
