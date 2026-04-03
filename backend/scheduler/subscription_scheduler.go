package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
)

// SubscriptionScheduler marks active subscriptions past end_date as expired.
type SubscriptionScheduler struct {
	interval time.Duration
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewSubscriptionScheduler(interval time.Duration) *SubscriptionScheduler {
	return &SubscriptionScheduler{
		interval: interval,
		stopChan: make(chan struct{}),
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
	s.markExpired()
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.markExpired()
		}
	}
}

func (s *SubscriptionScheduler) markExpired() {
	db := config.GetDB()
	if db == nil {
		log.Printf("Subscription scheduler: database not available")
		return
	}
	ctx := context.Background()
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
