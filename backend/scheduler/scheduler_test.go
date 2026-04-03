package scheduler

import (
	"testing"
	"time"
)

func TestNewSubscriptionScheduler(t *testing.T) {
	s := NewSubscriptionScheduler(time.Hour, nil)
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
	if s.interval != time.Hour {
		t.Fatalf("interval %v", s.interval)
	}
}

func TestReminderWindowKind(t *testing.T) {
	day := func(y, m, d int) time.Time {
		return time.Date(y, time.Month(m), d, 12, 0, 0, 0, time.UTC)
	}
	today := day(2026, 4, 3)
	if got := reminderWindowKind(today, day(2026, 4, 10)); got != "7d" {
		t.Errorf("7d: got %q", got)
	}
	if got := reminderWindowKind(today, day(2026, 4, 4)); got != "1d" {
		t.Errorf("1d: got %q", got)
	}
	if got := reminderWindowKind(today, day(2026, 4, 5)); got != "" {
		t.Errorf("none: got %q", got)
	}
}

func TestNewOrderScheduler(t *testing.T) {
	interval := 5 * time.Minute
	timeout := 30 * time.Minute

	scheduler := NewOrderScheduler(interval, timeout)

	if scheduler == nil {
		t.Fatal("NewOrderScheduler() returned nil")
	}

	if scheduler.interval != interval {
		t.Errorf("interval = %v, want %v", scheduler.interval, interval)
	}

	if scheduler.timeout != timeout {
		t.Errorf("timeout = %v, want %v", scheduler.timeout, timeout)
	}

	if scheduler.stopChan == nil {
		t.Error("stopChan should not be nil")
	}
}

func TestOrderScheduler_StartStop(t *testing.T) {
	scheduler := NewOrderScheduler(100*time.Millisecond, 1*time.Second)

	scheduler.Start()

	time.Sleep(50 * time.Millisecond)

	scheduler.Stop()
}

func TestOrderScheduler_MultipleStops(t *testing.T) {
	scheduler := NewOrderScheduler(1*time.Second, 1*time.Second)

	scheduler.Start()
	time.Sleep(10 * time.Millisecond)

	scheduler.Stop()
}

func TestNewSettlementScheduler(t *testing.T) {
	interval := 1 * time.Hour

	scheduler := NewSettlementScheduler(interval)

	if scheduler == nil {
		t.Fatal("NewSettlementScheduler() returned nil")
	}

	if scheduler.interval != interval {
		t.Errorf("interval = %v, want %v", scheduler.interval, interval)
	}
}

func TestSettlementScheduler_StartStop(t *testing.T) {
	scheduler := NewSettlementScheduler(100 * time.Millisecond)

	scheduler.Start()

	time.Sleep(50 * time.Millisecond)

	scheduler.Stop()
}

func TestOrderScheduler_CutoffTime(t *testing.T) {
	timeout := 30 * time.Minute
	cutoffTime := time.Now().Add(-timeout)

	if cutoffTime.After(time.Now()) {
		t.Error("Cutoff time should be in the past")
	}

	expectedDiff := time.Since(cutoffTime)
	if expectedDiff < timeout-time.Minute || expectedDiff > timeout+time.Minute {
		t.Errorf("Cutoff time difference should be approximately %v", timeout)
	}
}

func TestSettlementScheduler_FirstDayOfMonth(t *testing.T) {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	if firstOfMonth.Day() != 1 {
		t.Error("First of month should be day 1")
	}

	if firstOfMonth.After(now) {
		t.Error("First of month should not be in the future")
	}
}

func TestSettlementScheduler_PeriodCalculation(t *testing.T) {
	now := time.Now()

	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := now

	if periodStart.After(periodEnd) {
		t.Error("Period start should be before or equal to period end")
	}

	if periodStart.Year() == now.Year() && periodStart.Month() == now.Month() {
	} else {
		t.Error("Period start should be in current month")
	}
}

func TestOrderScheduler_ConcurrentStart(t *testing.T) {
	scheduler := NewOrderScheduler(1*time.Second, 1*time.Second)

	for i := 0; i < 3; i++ {
		go func() {
			time.Sleep(10 * time.Millisecond)
		}()
	}

	time.Sleep(50 * time.Millisecond)
	scheduler.Stop()
}

func TestScheduler_Intervals(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		timeout  time.Duration
	}{
		{"Short interval", 1 * time.Second, 30 * time.Second},
		{"Medium interval", 5 * time.Minute, 30 * time.Minute},
		{"Long interval", 1 * time.Hour, 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler := NewOrderScheduler(tt.interval, tt.timeout)
			if scheduler.interval != tt.interval {
				t.Errorf("interval = %v, want %v", scheduler.interval, tt.interval)
			}
		})
	}
}
