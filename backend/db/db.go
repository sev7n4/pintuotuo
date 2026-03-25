package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var dbConn *sql.DB

type PoolStats struct {
	MaxOpenConnections int           `json:"max_open_connections"`
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
}

func Init() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		//nolint:gosec,G101
		databaseURL = "postgresql://pintuotuo:dev_password_123@localhost:5432/pintuotuo_db?sslmode=disable"
	}

	var err error
	dbConn, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 50)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetime := getEnvAsDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute)
	connMaxIdleTime := getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute)

	dbConn.SetMaxOpenConns(maxOpenConns)
	dbConn.SetMaxIdleConns(maxIdleConns)
	dbConn.SetConnMaxLifetime(connMaxLifetime)
	dbConn.SetConnMaxIdleTime(connMaxIdleTime)

	if err := dbConn.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connection pool initialized: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		maxOpenConns, maxIdleConns, connMaxLifetime)

	return nil
}

func GetDB() *sql.DB {
	return dbConn
}

func Close() error {
	if dbConn != nil {
		return dbConn.Close()
	}
	return nil
}

func GetPoolStats() *PoolStats {
	if dbConn == nil {
		return nil
	}
	stats := dbConn.Stats()
	return &PoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}

func HealthCheck(ctx context.Context) error {
	if dbConn == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := dbConn.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
