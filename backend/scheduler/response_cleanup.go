package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services"
)

type ResponseCleanupScheduler struct {
	interval  time.Duration
	stopChan  chan struct{}
	isRunning bool
	mu        sync.Mutex
}

func NewResponseCleanupScheduler(interval time.Duration) *ResponseCleanupScheduler {
	return &ResponseCleanupScheduler{
		interval:  interval,
		stopChan:  make(chan struct{}),
		isRunning: false,
	}
}

func (s *ResponseCleanupScheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning {
		return
	}

	s.isRunning = true
	stopChan := make(chan struct{})
	s.stopChan = stopChan
	ticker := time.NewTicker(s.interval)

	go func(stop <-chan struct{}) {
		for {
			select {
			case <-ticker.C:
				s.cleanExpiredResponses()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}(stopChan)

	log.Println("Response cleanup scheduler started")
}

func (s *ResponseCleanupScheduler) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	stopChan := s.stopChan
	s.isRunning = false
	s.stopChan = nil
	s.mu.Unlock()

	close(stopChan)
	log.Println("Response cleanup scheduler stopped")
}

func (s *ResponseCleanupScheduler) cleanExpiredResponses() {
	db := config.GetDB()
	if db == nil {
		log.Println("Response cleanup scheduler: database not available")
		return
	}

	storageSvc := services.NewResponseStorageService(db)
	deleted, err := storageSvc.CleanExpiredResponses(context.Background())
	if err != nil {
		log.Printf("Response cleanup scheduler: failed to clean expired responses: %v", err)
		return
	}

	if deleted > 0 {
		log.Printf("Response cleanup scheduler: cleaned up %d expired responses", deleted)
	}
}
