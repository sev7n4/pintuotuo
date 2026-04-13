// Command reconcile runs full-database usage ledger checks (billable tokens in api_usage_logs vs
// token_transactions usage deductions, same Token 口径 as /admin/reconciliation/ledger).
// Intended for cron: exit 0 when matched, 1 when drift exceeds epsilon or on error.
//
// Usage: DATABASE_URL=... go run ./cmd/reconcile (from backend module root)
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/pintuotuo/backend/config"
	"github.com/pintuotuo/backend/services"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := config.InitDB(); err != nil {
		log.Fatalf("db: %v", err)
	}
	defer func() {
		_ = config.CloseDB()
	}()

	db := config.GetDB()
	logSum, txSum, err := services.GlobalUsageLedgerMatch(db)
	if err != nil {
		log.Fatalf("reconcile: %v", err)
	}
	delta := logSum - txSum
	ok := services.UsageReconcileOK(logSum, txSum)

	out := map[string]interface{}{
		"usage_log_total": logSum,
		"usage_tx_total":  txSum,
		"delta":           delta,
		"matched":         ok,
		"unit":            "tokens",
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		log.Fatalf("encode: %v", err)
	}
	if !ok {
		os.Exit(1)
	}
}
