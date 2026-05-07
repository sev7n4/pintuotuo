package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewResponseCleanupScheduler(t *testing.T) {
	interval := 24 * time.Hour
	s := NewResponseCleanupScheduler(interval)
	assert.NotNil(t, s)
	assert.Equal(t, interval, s.interval)
	assert.False(t, s.isRunning)
}

func TestResponseCleanupScheduler_StartStop(t *testing.T) {
	s := NewResponseCleanupScheduler(100 * time.Millisecond)

	s.Start()
	assert.True(t, s.isRunning)

	time.Sleep(50 * time.Millisecond)

	s.Stop()
	assert.False(t, s.isRunning)
}

func TestResponseCleanupScheduler_DoubleStart(t *testing.T) {
	s := NewResponseCleanupScheduler(1 * time.Second)

	s.Start()
	assert.True(t, s.isRunning)

	s.Start()
	assert.True(t, s.isRunning)

	time.Sleep(10 * time.Millisecond)
	s.Stop()
}

func TestResponseCleanupScheduler_DoubleStop(t *testing.T) {
	s := NewResponseCleanupScheduler(1 * time.Second)

	s.Start()
	time.Sleep(10 * time.Millisecond)

	s.Stop()
	assert.False(t, s.isRunning)

	s.Stop()
	assert.False(t, s.isRunning)
}

func TestResponseCleanupScheduler_StopWithoutStart(t *testing.T) {
	s := NewResponseCleanupScheduler(1 * time.Second)

	s.Stop()
	assert.False(t, s.isRunning)
}

func TestResponseCleanupScheduler_ConcurrentStartStop(t *testing.T) {
	s := NewResponseCleanupScheduler(1 * time.Second)

	done := make(chan struct{})

	go func() {
		for i := 0; i < 10; i++ {
			s.Start()
			time.Sleep(5 * time.Millisecond)
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(7 * time.Millisecond)
			s.Stop()
		}
		close(done)
	}()

	<-done
}

func TestResponseCleanupScheduler_Intervals(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
	}{
		{"Short interval", 1 * time.Second},
		{"Medium interval", 1 * time.Hour},
		{"Long interval", 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewResponseCleanupScheduler(tt.interval)
			assert.Equal(t, tt.interval, s.interval)
		})
	}
}

func TestResponseCleanupScheduler_StartStopCycle(t *testing.T) {
	s := NewResponseCleanupScheduler(100 * time.Millisecond)

	s.Start()
	time.Sleep(20 * time.Millisecond)
	s.Stop()

	time.Sleep(20 * time.Millisecond)

	s.Start()
	time.Sleep(20 * time.Millisecond)
	s.Stop()
}
