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
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	InputPrice   float64 `json:"input_price"`   // per 1M tokens
	OutputPrice  float64 `json:"output_price"`  // per 1M tokens
	Currency     string  `json:"currency"`
	EffectiveAt  string  `json:"effective_at"`
}

type BillingEngine struct {
	pricingCache map[string]PricingTier
	cacheMutex   sync.RWMutex
	db           *sql.DB
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
		}
		engine.loadPricing()
	})
	return engine
}

func (e *BillingEngine) loadPricing() {
	e.cacheMutex.Lock()
	defer e.cacheMutex.Unlock()

	defaultPricing := []PricingTier{
		{Provider: "openai", Model: "gpt-4-turbo-preview", InputPrice: 10, OutputPrice: 30, Currency: "USD"},
		{Provider: "openai", Model: "gpt-4", InputPrice: 30, OutputPrice: 60, Currency: "USD"},
		{Provider: "openai", Model: "gpt-4o", InputPrice: 5, OutputPrice: 15, Currency: "USD"},
		{Provider: "openai", Model: "gpt-4o-mini", InputPrice: 0.15, OutputPrice: 0.6, Currency: "USD"},
		{Provider: "openai", Model: "gpt-3.5-turbo", InputPrice: 0.5, OutputPrice: 1.5, Currency: "USD"},
		{Provider: "anthropic", Model: "claude-3-opus-20240229", InputPrice: 15, OutputPrice: 75, Currency: "USD"},
		{Provider: "anthropic", Model: "claude-3-sonnet-20240229", InputPrice: 3, OutputPrice: 15, Currency: "USD"},
		{Provider: "anthropic", Model: "claude-3-haiku-20240307", InputPrice: 0.25, OutputPrice: 1.25, Currency: "USD"},
		{Provider: "anthropic", Model: "claude-3-5-sonnet-20241022", InputPrice: 3, OutputPrice: 15, Currency: "USD"},
		{Provider: "google", Model: "gemini-pro", InputPrice: 0.5, OutputPrice: 1.5, Currency: "USD"},
		{Provider: "google", Model: "gemini-1.5-pro", InputPrice: 3.5, OutputPrice: 10.5, Currency: "USD"},
		{Provider: "google", Model: "gemini-1.5-flash", InputPrice: 0.075, OutputPrice: 0.3, Currency: "USD"},
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
			InputPrice:  1,
			OutputPrice: 2,
		}
	}

	inputCost := float64(inputTokens) * pricing.InputPrice / 1_000_000
	outputCost := float64(outputTokens) * pricing.OutputPrice / 1_000_000

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
		TotalUsage      float64 `json:"total_usage"`
		TotalRequests   int     `json:"total_requests"`
		TotalInputTok   int64   `json:"total_input_tokens"`
		TotalOutputTok  int64   `json:"total_output_tokens"`
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
