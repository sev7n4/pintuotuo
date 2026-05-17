package services

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Provider model catalog sync scheduler (env-driven; default off).
//
// PROVIDER_MODEL_SYNC_ENABLED=true
// PROVIDER_MODEL_SYNC_INTERVAL_HOURS=6  (1–168)
type ProviderModelSyncScheduler struct {
	stopChan chan struct{}
	reloadCh chan struct{}
	running  bool
	mu       sync.Mutex
}

var (
	providerModelSyncScheduler     *ProviderModelSyncScheduler
	providerModelSyncSchedulerOnce sync.Once
)

func GetProviderModelSyncScheduler() *ProviderModelSyncScheduler {
	providerModelSyncSchedulerOnce.Do(func() {
		providerModelSyncScheduler = &ProviderModelSyncScheduler{
			stopChan: make(chan struct{}, 1),
			reloadCh: make(chan struct{}, 1),
		}
	})
	return providerModelSyncScheduler
}

func ProviderModelSyncSchedulerEnabled() bool {
	v := strings.TrimSpace(os.Getenv("PROVIDER_MODEL_SYNC_ENABLED"))
	return v == "1" || strings.EqualFold(v, "true")
}

func ProviderModelSyncSchedulerInterval() time.Duration {
	hours := 6
	if raw := strings.TrimSpace(os.Getenv("PROVIDER_MODEL_SYNC_INTERVAL_HOURS")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 && n <= 168 {
			hours = n
		}
	}
	return time.Duration(hours) * time.Hour
}

func (s *ProviderModelSyncScheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return
	}
	s.running = true
	go s.run()
	log.Printf("Provider model sync scheduler started (enabled=%v interval=%s)",
		ProviderModelSyncSchedulerEnabled(), ProviderModelSyncSchedulerInterval())
}

func (s *ProviderModelSyncScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	s.running = false
	s.stopChan <- struct{}{}
	log.Println("Provider model sync scheduler stopped")
}

func (s *ProviderModelSyncScheduler) run() {
	for {
		if ProviderModelSyncSchedulerEnabled() {
			ctx := context.Background()
			svc := NewProviderCatalogGapService()
			if _, err := svc.SyncAllActiveProviderModels(ctx); err != nil {
				log.Printf("Provider model sync scheduler tick error: %v", err)
			}
		}
		interval := ProviderModelSyncSchedulerInterval()
		select {
		case <-s.stopChan:
			return
		case <-s.reloadCh:
			continue
		case <-time.After(interval):
		}
	}
}
