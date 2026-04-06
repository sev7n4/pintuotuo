package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pintuotuo/backend/logger"

	"github.com/robfig/cron/v3"
)

type SettlementScheduler struct {
	service     *SettlementService
	cron        *cron.Cron
	running     bool
	mu          sync.RWMutex
	cronExpr    string
}

func NewSettlementScheduler(service *SettlementService) *SettlementScheduler {
	return &SettlementScheduler{
		service:  service,
		cron:     cron.New(),
		running:  false,
		cronExpr: "0 2 1 * *", // 每月1日凌晨2点执行
	}
}

func (s *SettlementScheduler) Start() error {
	ctx := context.Background()

	logger.LogInfo(ctx, "settlement_scheduler", "Starting settlement scheduler", map[string]interface{}{
		"cron_expression": s.cronExpr,
	})

	_, err := s.cron.AddFunc(s.cronExpr, func() {
		s.ScheduleMonthlySettlement()
	})
	if err != nil {
		logger.LogError(ctx, "settlement_scheduler", "Failed to add cron job", err, nil)
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	s.cron.Start()

	logger.LogInfo(ctx, "settlement_scheduler", "Settlement scheduler started successfully", nil)

	return nil
}

func (s *SettlementScheduler) Stop() {
	ctx := context.Background()

	logger.LogInfo(ctx, "settlement_scheduler", "Stopping settlement scheduler", nil)

	s.cron.Stop()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	logger.LogInfo(ctx, "settlement_scheduler", "Settlement scheduler stopped successfully", nil)
}

func (s *SettlementScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *SettlementScheduler) GetCronExpression() string {
	return s.cronExpr
}

func (s *SettlementScheduler) ScheduleMonthlySettlement() error {
	ctx := context.Background()

	now := time.Now()
	periodStart, periodEnd := calculateSettlementPeriod(now)

	logger.LogInfo(ctx, "settlement_scheduler", "Starting monthly settlement generation", map[string]interface{}{
		"period_start": periodStart.Format("2006-01-02"),
		"period_end":   periodEnd.Format("2006-01-02"),
		"triggered_at": now.Format("2006-01-02 15:04:05"),
	})

	settlements, err := s.service.GenerateMonthlySettlements(periodStart, periodEnd)
	if err != nil {
		logger.LogError(ctx, "settlement_scheduler", "Failed to generate monthly settlements", err, nil)
		return fmt.Errorf("failed to generate monthly settlements: %w", err)
	}

	logger.LogInfo(ctx, "settlement_scheduler", "Monthly settlement generation completed", map[string]interface{}{
		"total_settlements": len(settlements),
	})

	return nil
}

func (s *SettlementScheduler) TriggerManualSettlement(year, month int) error {
	ctx := context.Background()

	logger.LogInfo(ctx, "settlement_scheduler", "Manual settlement triggered", map[string]interface{}{
		"year":  year,
		"month": month,
	})

	periodStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	settlements, err := s.service.GenerateMonthlySettlements(periodStart, periodEnd)
	if err != nil {
		logger.LogError(ctx, "settlement_scheduler", "Manual settlement generation failed", err, nil)
		return fmt.Errorf("failed to trigger manual settlement: %w", err)
	}

	logger.LogInfo(ctx, "settlement_scheduler", "Manual settlement generation completed", map[string]interface{}{
		"total_settlements": len(settlements),
	})

	return nil
}

func calculateSettlementPeriod(now time.Time) (time.Time, time.Time) {
	year := now.Year()
	month := now.Month()

	if month == 1 {
		year--
		month = 12
	} else {
		month--
	}

	periodStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	return periodStart, periodEnd
}

func InitializeScheduler() (*SettlementScheduler, error) {
	service := GetSettlementService()
	if service == nil {
		return nil, fmt.Errorf("settlement service is nil")
	}

	scheduler := NewSettlementScheduler(service)

	return scheduler, nil
}
