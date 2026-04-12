package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
)

type PricingTier struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	InputPrice  float64 `json:"input_price"`  // 元/1K tokens
	OutputPrice float64 `json:"output_price"` // 元/1K tokens
	Currency    string  `json:"currency"`
	EffectiveAt string  `json:"effective_at"`
}

type PreDeductConfig struct {
	Multiplier    int
	MaxMultiplier int
}

type BillingEngine struct {
	pricingCache map[string]PricingTier
	cacheMutex   sync.RWMutex
	db           *sql.DB
	configCache  map[string]*PreDeductConfig
	configMutex  sync.RWMutex
}

var (
	engine     *BillingEngine
	engineOnce sync.Once
)

func GetBillingEngine() *BillingEngine {
	engineOnce.Do(func() {
		engine = &BillingEngine{
			pricingCache: make(map[string]PricingTier),
			db:           config.GetDB(),
			configCache:  make(map[string]*PreDeductConfig),
		}
		engine.loadPricing()
	})
	return engine
}

func (e *BillingEngine) loadPricing() {
	e.cacheMutex.Lock()
	defer e.cacheMutex.Unlock()

	defaultPricing := []PricingTier{
		{Provider: "openai", Model: "gpt-4-turbo-preview", InputPrice: 0.01, OutputPrice: 0.03, Currency: "CNY"},
		{Provider: "openai", Model: "gpt-4", InputPrice: 0.03, OutputPrice: 0.06, Currency: "CNY"},
		{Provider: "openai", Model: "gpt-4o", InputPrice: 0.005, OutputPrice: 0.015, Currency: "CNY"},
		{Provider: "openai", Model: "gpt-4o-mini", InputPrice: 0.00015, OutputPrice: 0.0006, Currency: "CNY"},
		{Provider: "openai", Model: "gpt-3.5-turbo", InputPrice: 0.0005, OutputPrice: 0.0015, Currency: "CNY"},
		{Provider: "anthropic", Model: "claude-3-opus-20240229", InputPrice: 0.015, OutputPrice: 0.075, Currency: "CNY"},
		{Provider: "anthropic", Model: "claude-3-sonnet-20240229", InputPrice: 0.003, OutputPrice: 0.015, Currency: "CNY"},
		{Provider: "anthropic", Model: "claude-3-haiku-20240307", InputPrice: 0.00025, OutputPrice: 0.00125, Currency: "CNY"},
		{Provider: "anthropic", Model: "claude-3-5-sonnet-20241022", InputPrice: 0.003, OutputPrice: 0.015, Currency: "CNY"},
		{Provider: "google", Model: "gemini-pro", InputPrice: 0.0005, OutputPrice: 0.0015, Currency: "CNY"},
		{Provider: "google", Model: "gemini-1.5-pro", InputPrice: 0.0035, OutputPrice: 0.0105, Currency: "CNY"},
		{Provider: "google", Model: "gemini-1.5-flash", InputPrice: 0.000075, OutputPrice: 0.0003, Currency: "CNY"},
	}

	for _, p := range defaultPricing {
		key := fmt.Sprintf("%s:%s", p.Provider, p.Model)
		e.pricingCache[key] = p
	}
}

func (e *BillingEngine) GetPricing(provider, model string) (PricingTier, bool) {
	e.cacheMutex.RLock()
	defer e.cacheMutex.RUnlock()

	key := fmt.Sprintf("%s:%s", provider, model)
	pricing, ok := e.pricingCache[key]
	return pricing, ok
}

func (e *BillingEngine) CalculateCost(provider, model string, inputTokens, outputTokens int) float64 {
	pricing, ok := e.GetPricing(provider, model)
	if !ok {
		pricing = PricingTier{
			InputPrice:  0.001,
			OutputPrice: 0.002,
		}
	}

	inputCost := float64(inputTokens) * pricing.InputPrice / 1_000
	outputCost := float64(outputTokens) * pricing.OutputPrice / 1_000

	return inputCost + outputCost
}

func (e *BillingEngine) DeductBalance(userID int, amount float64, reason string, requestID string) error {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil {
		return fmt.Errorf("database not available")
	}

	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM tokens WHERE user_id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient balance: current=%.2f, required=%.2f", currentBalance, amount)
	}

	_, err = tx.Exec(
		"UPDATE tokens SET balance = balance - $1, total_used = total_used + $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to deduct balance: %w", err)
	}

	_, err = tx.Exec(
		"INSERT INTO token_transactions (user_id, type, amount, reason, request_id) VALUES ($1, $2, $3, $4, $5)",
		userID, "usage", -amount, reason, requestID,
	)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.TokenBalanceKey(userID))

	return nil
}

// DeductBalanceTx deducts token balance inside an existing transaction (e.g. order payment + fulfillment).
// Caller must Commit and invalidate cache.TokenBalanceKey(userID) after success.
func (e *BillingEngine) DeductBalanceTx(tx *sql.Tx, userID int, amount float64, reason string, requestID string) error {
	var currentBalance float64
	err := tx.QueryRow("SELECT balance FROM tokens WHERE user_id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("insufficient balance: no token account for user %d", userID)
		}
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient balance: current=%.2f, required=%.2f", currentBalance, amount)
	}

	_, err = tx.Exec(
		"UPDATE tokens SET balance = balance - $1, total_used = total_used + $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to deduct balance: %w", err)
	}

	reqID := interface{}(nil)
	if requestID != "" {
		reqID = requestID
	}
	_, err = tx.Exec(
		"INSERT INTO token_transactions (user_id, type, amount, reason, request_id) VALUES ($1, $2, $3, $4, $5)",
		userID, "usage", -amount, reason, reqID,
	)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	return nil
}

func (e *BillingEngine) AddBalance(userID int, amount float64, reason string, orderID int) error {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil {
		return fmt.Errorf("database not available")
	}

	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM tokens WHERE user_id = $1)", userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check token existence: %w", err)
	}

	if !exists {
		_, err = tx.Exec(
			"INSERT INTO tokens (user_id, balance, total_used, total_earned) VALUES ($1, $2, 0, $2)",
			userID, amount,
		)
	} else {
		_, err = tx.Exec(
			"UPDATE tokens SET balance = balance + $1, total_earned = total_earned + $1, updated_at = $2 WHERE user_id = $3",
			amount, time.Now(), userID,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to add balance: %w", err)
	}

	var orderIDVal interface{} = nil
	if orderID > 0 {
		orderIDVal = orderID
	}

	_, err = tx.Exec(
		"INSERT INTO token_transactions (user_id, type, amount, reason, order_id) VALUES ($1, $2, $3, $4, $5)",
		userID, "purchase", amount, reason, orderIDVal,
	)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.TokenBalanceKey(userID))

	return nil
}

func (e *BillingEngine) GetBalance(userID int) (float64, error) {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil {
		return 0, fmt.Errorf("database not available")
	}

	ctx := context.Background()
	cacheKey := cache.TokenBalanceKey(userID)

	if cachedBalance, err := cache.Get(ctx, cacheKey); err == nil {
		var balance float64
		if err := json.Unmarshal([]byte(cachedBalance), &balance); err == nil {
			return balance, nil
		}
	}

	var balance float64
	err := e.db.QueryRow("SELECT balance FROM tokens WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	if balanceJSON, err := json.Marshal(balance); err == nil {
		cache.Set(ctx, cacheKey, string(balanceJSON), cache.TokenBalanceTTL)
	}

	return balance, nil
}

func (e *BillingEngine) GetTransactionHistory(userID int, limit, offset int) ([]map[string]interface{}, error) {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := e.db.Query(
		"SELECT id, type, amount, reason, order_id, request_id, created_at FROM token_transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var id int
		var txType, reason string
		var amount float64
		var orderID, requestID sql.NullString
		var createdAt time.Time

		err := rows.Scan(&id, &txType, &amount, &reason, &orderID, &requestID, &createdAt)
		if err != nil {
			log.Printf("Failed to scan transaction row: %v", err)
			continue
		}

		transactions = append(transactions, map[string]interface{}{
			"id":         id,
			"type":       txType,
			"amount":     amount,
			"reason":     reason,
			"order_id":   orderID.String,
			"request_id": requestID.String,
			"created_at": createdAt,
		})
	}

	return transactions, nil
}

func (e *BillingEngine) RefundTokens(userID int, amount float64, reason string, orderID int) error {
	return e.AddBalance(userID, amount, reason, orderID)
}

func (e *BillingEngine) GetUsageStats(userID int, startDate, endDate time.Time) (map[string]interface{}, error) {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var stats struct {
		TotalUsage     float64 `json:"total_usage"`
		TotalRequests  int     `json:"total_requests"`
		TotalInputTok  int64   `json:"total_input_tokens"`
		TotalOutputTok int64   `json:"total_output_tokens"`
	}

	err := e.db.QueryRow(
		`SELECT COALESCE(SUM(cost), 0), COUNT(*), COALESCE(SUM(input_tokens), 0), COALESCE(SUM(output_tokens), 0) 
		 FROM api_usage_logs 
		 WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3`,
		userID, startDate, endDate,
	).Scan(&stats.TotalUsage, &stats.TotalRequests, &stats.TotalInputTok, &stats.TotalOutputTok)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	rows, err := e.db.Query(
		`SELECT provider, COUNT(*) as count, SUM(cost) as cost 
		 FROM api_usage_logs 
		 WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 
		 GROUP BY provider 
		 ORDER BY cost DESC`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}
	defer rows.Close()

	byProvider := make([]map[string]interface{}, 0)
	for rows.Next() {
		var provider string
		var count int
		var cost float64
		if err := rows.Scan(&provider, &count, &cost); err != nil {
			continue
		}
		byProvider = append(byProvider, map[string]interface{}{
			"provider": provider,
			"count":    count,
			"cost":     cost,
		})
	}

	return map[string]interface{}{
		"summary":     stats,
		"by_provider": byProvider,
		"period": map[string]string{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
		},
	}, nil
}

func (e *BillingEngine) CalculateTokenUsage(inputTokens, outputTokens int) int64 {
	return int64(inputTokens) + int64(outputTokens)
}

func (e *BillingEngine) EstimateTokenUsage(inputTokens int, config *PreDeductConfig) int64 {
	if config == nil {
		config = &PreDeductConfig{Multiplier: 2, MaxMultiplier: 10}
	}

	multiplier := config.Multiplier
	if multiplier <= 0 {
		multiplier = 2
	}

	estimate := int64(inputTokens * multiplier)

	maxMultiplier := config.MaxMultiplier
	if maxMultiplier <= 0 {
		maxMultiplier = 10
	}
	maxEstimate := int64(inputTokens * maxMultiplier)

	if estimate > maxEstimate && maxEstimate > 0 {
		estimate = maxEstimate
	}

	return estimate
}

func (e *BillingEngine) GetPreDeductConfig(skuID, spuID int, providerCode string) *PreDeductConfig {
	cacheKey := fmt.Sprintf("pre_deduct_config:%d:%d:%s", skuID, spuID, providerCode)

	e.configMutex.RLock()
	if cached, ok := e.configCache[cacheKey]; ok {
		e.configMutex.RUnlock()
		return cached
	}
	e.configMutex.RUnlock()

	config := &PreDeductConfig{
		Multiplier:    2,
		MaxMultiplier: 10,
	}

	providerConfig := e.getProviderPreDeductConfig(providerCode)
	if providerConfig != nil {
		config = providerConfig
	}

	spuConfig := e.getSPUPreDeductConfig(spuID)
	if spuConfig != nil {
		config = spuConfig
	}

	skuConfig := e.getSKUPreDeductConfig(skuID)
	if skuConfig != nil {
		config = skuConfig
	}

	e.configMutex.Lock()
	e.configCache[cacheKey] = config
	e.configMutex.Unlock()

	return config
}

func (e *BillingEngine) getProviderPreDeductConfig(providerCode string) *PreDeductConfig {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil || providerCode == "" {
		return nil
	}

	var segmentConfig []byte
	err := e.db.QueryRow(
		"SELECT segment_config FROM model_providers WHERE code = $1 AND status = 'active'",
		providerCode,
	).Scan(&segmentConfig)
	if err != nil {
		return nil
	}

	if len(segmentConfig) == 0 {
		return nil
	}

	var cfg struct {
		PreDeductMultiplier    int `json:"pre_deduct_multiplier"`
		PreDeductMaxMultiplier int `json:"pre_deduct_max_multiplier"`
	}
	if err := json.Unmarshal(segmentConfig, &cfg); err != nil {
		return nil
	}

	if cfg.PreDeductMultiplier > 0 {
		return &PreDeductConfig{
			Multiplier:    cfg.PreDeductMultiplier,
			MaxMultiplier: cfg.PreDeductMaxMultiplier,
		}
	}

	return nil
}

func (e *BillingEngine) getSPUPreDeductConfig(spuID int) *PreDeductConfig {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil || spuID <= 0 {
		return nil
	}

	var multiplier, maxMultiplier sql.NullInt64
	err := e.db.QueryRow(
		"SELECT pre_deduct_multiplier, pre_deduct_max_multiplier FROM spus WHERE id = $1 AND pre_deduct_multiplier IS NOT NULL",
		spuID,
	).Scan(&multiplier, &maxMultiplier)
	if err != nil {
		return nil
	}

	if multiplier.Valid && multiplier.Int64 > 0 {
		maxVal := 10
		if maxMultiplier.Valid && maxMultiplier.Int64 > 0 {
			maxVal = int(maxMultiplier.Int64)
		}
		return &PreDeductConfig{
			Multiplier:    int(multiplier.Int64),
			MaxMultiplier: maxVal,
		}
	}

	return nil
}

func (e *BillingEngine) getSKUPreDeductConfig(skuID int) *PreDeductConfig {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil || skuID <= 0 {
		return nil
	}

	var multiplier, maxMultiplier sql.NullInt64
	err := e.db.QueryRow(
		"SELECT pre_deduct_multiplier, pre_deduct_max_multiplier FROM skus WHERE id = $1 AND pre_deduct_multiplier IS NOT NULL",
		skuID,
	).Scan(&multiplier, &maxMultiplier)
	if err != nil {
		return nil
	}

	if multiplier.Valid && multiplier.Int64 > 0 {
		maxVal := 10
		if maxMultiplier.Valid && maxMultiplier.Int64 > 0 {
			maxVal = int(maxMultiplier.Int64)
		}
		return &PreDeductConfig{
			Multiplier:    int(multiplier.Int64),
			MaxMultiplier: maxVal,
		}
	}

	return nil
}

func (e *BillingEngine) InvalidateConfigCache(skuID, spuID int, providerCode string) {
	cacheKey := fmt.Sprintf("pre_deduct_config:%d:%d:%s", skuID, spuID, providerCode)
	e.configMutex.Lock()
	delete(e.configCache, cacheKey)
	e.configMutex.Unlock()
}

func (e *BillingEngine) PreDeductBalance(userID int, amount int64, reason string, requestID string) error {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil {
		return fmt.Errorf("database not available")
	}

	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var currentBalance float64
	err = tx.QueryRow("SELECT balance FROM tokens WHERE user_id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	if currentBalance < float64(amount) {
		return fmt.Errorf("insufficient balance: current=%.0f, required=%d", currentBalance, amount)
	}

	_, err = tx.Exec(
		"UPDATE tokens SET balance = balance - $1, updated_at = $2 WHERE user_id = $3",
		amount, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to pre-deduct balance: %w", err)
	}

	_, err = tx.Exec(
		"INSERT INTO pre_deductions (user_id, request_id, pre_deduct_amount, status) VALUES ($1, $2, $3, 'pending')",
		userID, requestID, amount,
	)
	if err != nil {
		return fmt.Errorf("failed to record pre-deduction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.TokenBalanceKey(userID))

	return nil
}

func (e *BillingEngine) SettlePreDeduct(userID int, requestID string, actualUsage int64) error {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil {
		return fmt.Errorf("database not available")
	}

	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var preDeductAmount int64
	var status string
	err = tx.QueryRow(
		"SELECT pre_deduct_amount, status FROM pre_deductions WHERE user_id = $1 AND request_id = $2 FOR UPDATE",
		userID, requestID,
	).Scan(&preDeductAmount, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to get pre-deduction record: %w", err)
	}

	if status != "pending" {
		return nil
	}

	diff := preDeductAmount - actualUsage

	if diff > 0 {
		_, err = tx.Exec(
			"UPDATE tokens SET balance = balance + $1, updated_at = $2 WHERE user_id = $3",
			diff, time.Now(), userID,
		)
		if err != nil {
			return fmt.Errorf("failed to refund balance: %w", err)
		}
	} else if diff < 0 {
		extraNeeded := -diff
		_, err = tx.Exec(
			"UPDATE tokens SET balance = balance - $1, total_used = total_used + $1, updated_at = $2 WHERE user_id = $3",
			extraNeeded, time.Now(), userID,
		)
		if err != nil {
			return fmt.Errorf("failed to deduct extra balance: %w", err)
		}
	}

	_, err = tx.Exec(
		"UPDATE pre_deductions SET actual_amount = $1, status = 'settled', settled_at = $2 WHERE user_id = $3 AND request_id = $4",
		actualUsage, time.Now(), userID, requestID,
	)
	if err != nil {
		return fmt.Errorf("failed to update pre-deduction status: %w", err)
	}

	_, err = tx.Exec(
		"INSERT INTO token_transactions (user_id, type, amount, reason, request_id) VALUES ($1, 'usage', $2, $3, $4)",
		userID, -actualUsage, "API usage settled", requestID,
	)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.TokenBalanceKey(userID))

	return nil
}

func (e *BillingEngine) CancelPreDeduct(userID int, requestID string) error {
	if e.db == nil {
		e.db = config.GetDB()
	}
	if e.db == nil {
		return fmt.Errorf("database not available")
	}

	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var preDeductAmount int64
	var status string
	err = tx.QueryRow(
		"SELECT pre_deduct_amount, status FROM pre_deductions WHERE user_id = $1 AND request_id = $2 FOR UPDATE",
		userID, requestID,
	).Scan(&preDeductAmount, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("failed to get pre-deduction record: %w", err)
	}

	if status != "pending" {
		return nil
	}

	_, err = tx.Exec(
		"UPDATE tokens SET balance = balance + $1, updated_at = $2 WHERE user_id = $3",
		preDeductAmount, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to refund balance: %w", err)
	}

	_, err = tx.Exec(
		"UPDATE pre_deductions SET status = 'cancelled', settled_at = $1 WHERE user_id = $2 AND request_id = $3",
		time.Now(), userID, requestID,
	)
	if err != nil {
		return fmt.Errorf("failed to update pre-deduction status: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ctx := context.Background()
	cache.Delete(ctx, cache.TokenBalanceKey(userID))

	return nil
}

func (e *BillingEngine) CalculateCostForSettlement(provider, model string, inputTokens, outputTokens int) float64 {
	return e.CalculateCost(provider, model, inputTokens, outputTokens)
}
