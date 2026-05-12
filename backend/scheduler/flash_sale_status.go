package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services"
)

const flashSaleActiveCacheKey = "flash_sales:active"

// FlashSaleStatusScheduler periodically promotes flash sale statuses so fields update without relying on public traffic.
type FlashSaleStatusScheduler struct {
	interval time.Duration
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewFlashSaleStatusScheduler(interval time.Duration) *FlashSaleStatusScheduler {
	return &FlashSaleStatusScheduler{
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

func (s *FlashSaleStatusScheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Printf("Flash sale status scheduler started: interval %v", s.interval)
}

func (s *FlashSaleStatusScheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Flash sale status scheduler stopped")
}

func (s *FlashSaleStatusScheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *FlashSaleStatusScheduler) tick() {
	db := config.GetDB()
	if db == nil {
		log.Printf("Flash sale status scheduler: database not available")
		return
	}
	ctx := context.Background()
	changed, err := services.PromoteFlashSaleStatuses(db, time.Now())
	if err != nil {
		log.Printf("Flash sale status scheduler: promote failed: %v", err)
		return
	}
	if changed {
		cache.Delete(ctx, flashSaleActiveCacheKey)
	}
}
