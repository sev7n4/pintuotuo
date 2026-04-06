package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSettlementScheduler_Start(t *testing.T) {
	t.Run("TC-SETTLEMENT-SCHEDULER-001: should start scheduler successfully", func(t *testing.T) {
		t.Skip("Requires database setup - will be implemented in integration tests")
		service := &SettlementService{}
		scheduler := NewSettlementScheduler(service)

		err := scheduler.Start()
		assert.NoError(t, err)
		assert.True(t, scheduler.IsRunning())

		scheduler.Stop()
		assert.False(t, scheduler.IsRunning())
	})
}

func TestSettlementScheduler_ScheduleMonthlySettlement(t *testing.T) {
	t.Run("TC-SETTLEMENT-SCHEDULER-002: should generate settlements for previous month", func(t *testing.T) {
		t.Skip("Requires database setup - will be implemented in integration tests")
		service := &SettlementService{}
		scheduler := NewSettlementScheduler(service)

		err := scheduler.ScheduleMonthlySettlement()
		assert.NoError(t, err)
	})
}

func TestSettlementScheduler_TriggerManualSettlement(t *testing.T) {
	t.Run("TC-SETTLEMENT-SCHEDULER-003: should trigger manual settlement for specific month", func(t *testing.T) {
		t.Skip("Requires database setup - will be implemented in integration tests")
		service := &SettlementService{}
		scheduler := NewSettlementScheduler(service)

		err := scheduler.TriggerManualSettlement(2026, 3)
		assert.NoError(t, err)
	})
}

func TestSettlementScheduler_CronExpression(t *testing.T) {
	t.Run("TC-SETTLEMENT-SCHEDULER-004: should use correct cron expression", func(t *testing.T) {
		service := &SettlementService{}
		scheduler := NewSettlementScheduler(service)

		expectedCron := "0 2 1 * *"
		assert.Equal(t, expectedCron, scheduler.GetCronExpression())
	})
}

func TestSettlementScheduler_PeriodCalculation(t *testing.T) {
	t.Run("TC-SETTLEMENT-SCHEDULER-005: should calculate correct period for settlement", func(t *testing.T) {
		now := time.Date(2026, 4, 6, 10, 0, 0, 0, time.UTC)

		periodStart, periodEnd := calculateSettlementPeriod(now)

		expectedStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		expectedEnd := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

		assert.Equal(t, expectedStart, periodStart)
		assert.Equal(t, expectedEnd, periodEnd)
	})
}
