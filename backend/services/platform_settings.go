package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/pintuotuo/backend/config"
)

const (
	PlatformKeyHealthSchedulerEnabled         = "health_scheduler_enabled"
	PlatformKeyHealthSchedulerIntervalSeconds = "health_scheduler_interval_seconds"
	PlatformKeyHealthSchedulerBatch           = "health_scheduler_batch"
)

// 默认：方案 B — 开启主动探测，长间隔 + 小批量，成本可控。
const (
	defaultHealthSchedulerEnabled         = true
	defaultHealthSchedulerIntervalSeconds = 3600
	defaultHealthSchedulerBatch           = 2
)

const (
	minHealthSchedulerIntervalSeconds = 60
	maxHealthSchedulerIntervalSeconds = 86400
	minHealthSchedulerBatch           = 1
	maxHealthSchedulerBatch           = 50
)

var (
	platformMu sync.RWMutex
	// 内存缓存，由 ReloadPlatformSettingsCache 与 UPDATE 后刷新。
	healthSchedulerEnabledCache         = defaultHealthSchedulerEnabled
	healthSchedulerIntervalSecondsCache = defaultHealthSchedulerIntervalSeconds
	healthSchedulerBatchCache           = defaultHealthSchedulerBatch
)

// HealthSchedulerPlatformConfig 主动探测调度（与 per-key health_check_level 无关）。
type HealthSchedulerPlatformConfig struct {
	Enabled         bool
	IntervalSeconds int
	Batch           int
}

func GetHealthSchedulerPlatformConfig() HealthSchedulerPlatformConfig {
	platformMu.RLock()
	defer platformMu.RUnlock()
	return HealthSchedulerPlatformConfig{
		Enabled:         healthSchedulerEnabledCache,
		IntervalSeconds: healthSchedulerIntervalSecondsCache,
		Batch:           healthSchedulerBatchCache,
	}
}

func parseBoolString(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "true") || strings.TrimSpace(s) == "1"
}

func ReloadPlatformSettingsCache(ctx context.Context) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	rows, err := db.QueryContext(ctx, `
		SELECT key, value FROM platform_settings
		WHERE key IN ($1, $2, $3)`,
		PlatformKeyHealthSchedulerEnabled,
		PlatformKeyHealthSchedulerIntervalSeconds,
		PlatformKeyHealthSchedulerBatch,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	enabled := defaultHealthSchedulerEnabled
	intervalSec := defaultHealthSchedulerIntervalSeconds
	batch := defaultHealthSchedulerBatch

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			continue
		}
		switch k {
		case PlatformKeyHealthSchedulerEnabled:
			enabled = parseBoolString(v)
		case PlatformKeyHealthSchedulerIntervalSeconds:
			if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				intervalSec = clampInt(n, minHealthSchedulerIntervalSeconds, maxHealthSchedulerIntervalSeconds)
			}
		case PlatformKeyHealthSchedulerBatch:
			if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				batch = clampInt(n, minHealthSchedulerBatch, maxHealthSchedulerBatch)
			}
		}
	}

	platformMu.Lock()
	healthSchedulerEnabledCache = enabled
	healthSchedulerIntervalSecondsCache = intervalSec
	healthSchedulerBatchCache = batch
	platformMu.Unlock()

	return nil
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// UpdateHealthSchedulerPlatformSettings 写入 DB；触发器会 NOTIFY，监听方会 Reload。
func UpdateHealthSchedulerPlatformSettings(ctx context.Context, cfg HealthSchedulerPlatformConfig) error {
	if err := ValidateHealthSchedulerPlatformConfig(cfg); err != nil {
		return err
	}

	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	enabledStr := "false"
	if cfg.Enabled {
		enabledStr = "true"
	}

	upserts := []struct {
		key   string
		value string
	}{
		{PlatformKeyHealthSchedulerEnabled, enabledStr},
		{PlatformKeyHealthSchedulerIntervalSeconds, strconv.Itoa(cfg.IntervalSeconds)},
		{PlatformKeyHealthSchedulerBatch, strconv.Itoa(cfg.Batch)},
	}

	for _, row := range upserts {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO platform_settings (key, value, updated_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP`,
			row.key, row.value,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return ReloadPlatformSettingsCache(ctx)
}

// ValidateHealthSchedulerPlatformConfig 用于 API 校验（非法则返回错误，不写库）。
func ValidateHealthSchedulerPlatformConfig(cfg HealthSchedulerPlatformConfig) error {
	if cfg.IntervalSeconds < minHealthSchedulerIntervalSeconds || cfg.IntervalSeconds > maxHealthSchedulerIntervalSeconds {
		return fmt.Errorf("health_scheduler_interval_seconds must be between %d and %d",
			minHealthSchedulerIntervalSeconds, maxHealthSchedulerIntervalSeconds)
	}
	if cfg.Batch < minHealthSchedulerBatch || cfg.Batch > maxHealthSchedulerBatch {
		return fmt.Errorf("health_scheduler_batch must be between %d and %d", minHealthSchedulerBatch, maxHealthSchedulerBatch)
	}
	return nil
}

// HealthSchedulerPlatformLimits 供管理端展示校验范围。
func HealthSchedulerPlatformLimits() map[string]int {
	return map[string]int{
		"interval_seconds_min": minHealthSchedulerIntervalSeconds,
		"interval_seconds_max": maxHealthSchedulerIntervalSeconds,
		"batch_min":            minHealthSchedulerBatch,
		"batch_max":            maxHealthSchedulerBatch,
	}
}
