package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pintuotuo/backend/billing"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/notification"
	"github.com/pintuotuo/backend/services"
)

// SubscriptionScheduler sends expiry reminders, runs Token auto-renewal, then marks expired subscriptions.
type SubscriptionScheduler struct {
	interval time.Duration
	stopChan chan struct{}
	wg       sync.WaitGroup
	notify   *notification.NotificationService
}

func NewSubscriptionScheduler(interval time.Duration, notify *notification.NotificationService) *SubscriptionScheduler {
	return &SubscriptionScheduler{
		interval: interval,
		stopChan: make(chan struct{}),
		notify:   notify,
	}
}

func (s *SubscriptionScheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Printf("Subscription scheduler started: interval %v", s.interval)
}

func (s *SubscriptionScheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Subscription scheduler stopped")
}

func (s *SubscriptionScheduler) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	s.tick()
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *SubscriptionScheduler) tick() {
	db := config.GetDB()
	if db == nil {
		log.Printf("Subscription scheduler: database not available")
		return
	}
	ctx := context.Background()
	s.processReminders(ctx, db)
	s.processTokenAutoRenew(ctx, db)
	s.markExpired(ctx, db)
}

func dateUTC(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

// reminderWindowKind returns "7d", "1d", or "" when endDate (UTC calendar) is exactly 7 or 1 day after todayUTC.
func reminderWindowKind(todayUTC, endDate time.Time) string {
	t0 := dateUTC(todayUTC)
	endD := dateUTC(endDate)
	switch {
	case endD.Equal(t0.AddDate(0, 0, 7)):
		return "7d"
	case endD.Equal(t0.AddDate(0, 0, 1)):
		return "1d"
	default:
		return ""
	}
}

func (s *SubscriptionScheduler) processReminders(ctx context.Context, db *sql.DB) {
	rows, err := db.QueryContext(ctx, `
		SELECT us.id, us.user_id, us.end_date, us.auto_renew, u.email, u.name, sp.name
		FROM user_subscriptions us
		JOIN users u ON u.id = us.user_id
		JOIN skus s ON s.id = us.sku_id
		JOIN spus sp ON sp.id = s.spu_id
		WHERE us.status = 'active'
		  AND s.sku_type IN ('subscription', 'trial')
		  AND (us.end_date = CURRENT_DATE + INTERVAL '7 days'
		       OR us.end_date = CURRENT_DATE + INTERVAL '1 day')`)
	if err != nil {
		log.Printf("Subscription scheduler: reminder query failed: %v", err)
		return
	}
	defer rows.Close()

	now := time.Now()

	for rows.Next() {
		var subID, userID int
		var endDate time.Time
		var autoRenew bool
		var email, userName, spuName string
		if err := rows.Scan(&subID, &userID, &endDate, &autoRenew, &email, &userName, &spuName); err != nil {
			log.Printf("Subscription scheduler: reminder scan: %v", err)
			continue
		}

		kind := reminderWindowKind(now, endDate)
		if kind == "" {
			continue
		}

		res, err := db.ExecContext(ctx, `
			INSERT INTO subscription_reminders (subscription_id, kind, channel)
			VALUES ($1, $2, 'in_app+email')
			ON CONFLICT (subscription_id, kind) DO NOTHING`,
			subID, kind)
		if err != nil {
			log.Printf("Subscription scheduler: reminder insert %d %s: %v", subID, kind, err)
			continue
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			continue
		}

		endStr := endDate.Format("2006-01-02")
		arText := "否"
		if autoRenew {
			arText = "是（将尝试使用 Token 余额自动续费）"
		}
		title := "订阅即将到期"
		body := fmt.Sprintf("您的订阅「%s」将于 %s 到期。已开启自动续费：%s。", spuName, endStr, arText)
		if kind == "1d" {
			title = "订阅明日到期"
			body = fmt.Sprintf("您的订阅「%s」将于明日（%s）到期。已开启自动续费：%s。", spuName, endStr, arText)
		}

		data, _ := json.Marshal(map[string]interface{}{
			"subscription_id": subID,
			"end_date":        endStr,
			"kind":            kind,
			"spu_name":        spuName,
		})

		_, err = db.ExecContext(ctx, `
			INSERT INTO notifications (user_id, type, title, content, data)
			VALUES ($1, 'subscription_expiring', $2, $3, $4::jsonb)`,
			userID, title, body, string(data))
		if err != nil {
			log.Printf("Subscription scheduler: in-app notification %d: %v", subID, err)
		}

		if s.notify != nil {
			if err := s.notify.TrySendSubscriptionExpiringEmail(email, userName, map[string]interface{}{
				"SPUName":      spuName,
				"EndDate":      endStr,
				"Kind":         kind,
				"AutoRenewTxt": arText,
			}); err != nil {
				log.Printf("Subscription scheduler: email %d: %v", subID, err)
			}
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Subscription scheduler: reminder rows: %v", err)
	}
}

func (s *SubscriptionScheduler) processTokenAutoRenew(ctx context.Context, db *sql.DB) {
	engine := billing.GetBillingEngine()
	attempted, renewed, disabled, err := services.ProcessDueTokenAutoRenewals(ctx, db, engine)
	if err != nil {
		log.Printf("Subscription scheduler: token auto-renew failed: %v", err)
		return
	}
	if attempted > 0 {
		log.Printf("Subscription scheduler: token renew attempted=%d renewed=%d auto_renew_disabled=%d", attempted, renewed, disabled)
	}
}

func (s *SubscriptionScheduler) markExpired(ctx context.Context, db *sql.DB) {
	res, err := db.ExecContext(ctx, `
		UPDATE user_subscriptions
		SET status = 'expired', updated_at = CURRENT_TIMESTAMP
		WHERE status = 'active' AND end_date < CURRENT_DATE`)
	if err != nil {
		log.Printf("Subscription scheduler: expire update failed: %v", err)
		return
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		log.Printf("Subscription scheduler: marked %d subscriptions expired", n)
	}
}
