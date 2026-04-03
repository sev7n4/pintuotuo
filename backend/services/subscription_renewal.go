package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/cache"
)

// ProcessDueTokenAutoRenewals attempts one billing period extension for each due subscription (Token debit).
// Only rows with sku_type = subscription, status active, auto_renew true, end_date <= CURRENT_DATE.
func ProcessDueTokenAutoRenewals(ctx context.Context, db *sql.DB, engine *billing.BillingEngine) (attempted int, renewed int, disabled int, err error) {
	if db == nil || engine == nil {
		return 0, 0, 0, fmt.Errorf("db or billing engine nil")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT us.id FROM user_subscriptions us
		JOIN skus s ON s.id = us.sku_id
		WHERE us.status = 'active' AND us.auto_renew = TRUE
		  AND us.end_date <= CURRENT_DATE
		  AND s.sku_type = 'subscription'`)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return attempted, renewed, disabled, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return attempted, renewed, disabled, err
	}

	for _, id := range ids {
		attempted++
		r, d, e := renewSubscriptionWithToken(ctx, db, engine, id)
		if e != nil {
			return attempted, renewed, disabled, e
		}
		if r {
			renewed++
		}
		if d {
			disabled++
		}
	}
	return attempted, renewed, disabled, nil
}

// renewSubscriptionWithToken returns (renewed, autoRenewDisabled, err).
func renewSubscriptionWithToken(ctx context.Context, db *sql.DB, engine *billing.BillingEngine, subscriptionID int) (renewed bool, autoRenewDisabled bool, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, false, err
	}
	defer tx.Rollback()

	var userID, skuID int
	var endDate time.Time
	var autoRenew bool
	var status string
	var skuType string
	var retailPrice float64
	var subPeriod sql.NullString
	var validDays sql.NullInt64

	err = tx.QueryRowContext(ctx, `
		SELECT us.user_id, us.sku_id, us.end_date, us.auto_renew, us.status,
		       s.sku_type, s.retail_price, s.subscription_period, s.valid_days
		FROM user_subscriptions us
		JOIN skus s ON s.id = us.sku_id
		WHERE us.id = $1 FOR UPDATE`,
		subscriptionID,
	).Scan(&userID, &skuID, &endDate, &autoRenew, &status, &skuType, &retailPrice, &subPeriod, &validDays)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		return false, false, err
	}

	if status != "active" || !autoRenew || skuType != skuTypeSubscription {
		return false, false, tx.Commit()
	}

	if !subscriptionEndOnOrBeforeToday(endDate) {
		return false, false, tx.Commit()
	}

	period := ""
	if subPeriod.Valid {
		period = subPeriod.String
	}
	vd := 0
	if validDays.Valid {
		vd = int(validDays.Int64)
	}

	base := dateUTC(endDate)
	newEnd := StackSubscriptionPeriods(base, period, vd, 1)
	newEndStr := newEnd.Format("2006-01-02")

	reason := fmt.Sprintf("订阅自动续费 #%d", subscriptionID)
	if err = engine.DeductBalanceTx(tx, userID, retailPrice, reason, ""); err != nil {
		_, uerr := tx.ExecContext(ctx, `
			UPDATE user_subscriptions SET auto_renew = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
			subscriptionID,
		)
		if uerr != nil {
			return false, false, uerr
		}
		if cerr := tx.Commit(); cerr != nil {
			return false, false, cerr
		}
		cache.Delete(ctx, cache.TokenBalanceKey(userID))
		return false, true, nil
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE user_subscriptions SET end_date = $1::date, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		newEndStr, subscriptionID,
	)
	if err != nil {
		return false, false, err
	}
	if cerr := tx.Commit(); cerr != nil {
		return false, false, cerr
	}
	cache.Delete(ctx, cache.TokenBalanceKey(userID))
	return true, false, nil
}

func dateUTC(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

// subscriptionEndOnOrBeforeToday is true when the subscription's end_date (calendar) is on or before today (UTC calendar).
func subscriptionEndOnOrBeforeToday(endDate time.Time) bool {
	endD := dateUTC(endDate)
	today := dateUTC(subscriptionRenewalNow())
	return !endD.After(today)
}

// subscriptionRenewalNow is time.Now in production; tests may override via subscriptionRenewalNowForTest.
func subscriptionRenewalNow() time.Time {
	if subscriptionRenewalNowForTest != nil {
		return subscriptionRenewalNowForTest()
	}
	return time.Now()
}

// subscriptionRenewalNowForTest when non-nil replaces the clock in subscriptionEndOnOrBeforeToday (tests only).
var subscriptionRenewalNowForTest func() time.Time
