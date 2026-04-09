package services

import "testing"

func TestValidateHealthSchedulerPlatformConfig(t *testing.T) {
	if err := ValidateHealthSchedulerPlatformConfig(HealthSchedulerPlatformConfig{
		Enabled: true, IntervalSeconds: 3600, Batch: 2,
	}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHealthSchedulerPlatformConfig(HealthSchedulerPlatformConfig{
		Enabled: true, IntervalSeconds: 30, Batch: 2,
	}); err == nil {
		t.Fatal("expected error for interval too small")
	}
	if err := ValidateHealthSchedulerPlatformConfig(HealthSchedulerPlatformConfig{
		Enabled: true, IntervalSeconds: 3600, Batch: 0,
	}); err == nil {
		t.Fatal("expected error for batch too small")
	}
}
