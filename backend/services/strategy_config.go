package services

type StrategyConfig struct {
	Strategy                string  `json:"strategy"`
	MaxRetryCount           int     `json:"max_retry_count"`
	RetryBackoffBase        int     `json:"retry_backoff_base"`
	CircuitBreakerThreshold int     `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   int     `json:"circuit_breaker_timeout"`
	PriceWeight             float64 `json:"price_weight"`
	LatencyWeight           float64 `json:"latency_weight"`
	SuccessWeight           float64 `json:"success_weight"`
	ReliabilityWeight       float64 `json:"reliability_weight"`
}
