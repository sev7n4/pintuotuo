package services

import (
	"time"
)

type RequestIntent string

const (
	IntentChat       RequestIntent = "chat"
	IntentCompletion RequestIntent = "completion"
	IntentEmbedding  RequestIntent = "embedding"
	IntentImage      RequestIntent = "image"
	IntentAudio      RequestIntent = "audio"
	IntentModeration RequestIntent = "moderation"
	IntentResponse   RequestIntent = "response"
	IntentUnknown    RequestIntent = "unknown"
)

type RequestComplexity string

const (
	ComplexitySimple  RequestComplexity = "simple"
	ComplexityMedium  RequestComplexity = "medium"
	ComplexityComplex RequestComplexity = "complex"
)

type RequestAnalysis struct {
	Intent          RequestIntent          `json:"intent"`
	EndpointType    string                 `json:"endpoint_type,omitempty"`
	Complexity      RequestComplexity      `json:"complexity"`
	EstimatedTokens int                    `json:"estimated_tokens"`
	ActualTokens    int                    `json:"actual_tokens,omitempty"`
	Model           string                 `json:"model,omitempty"`
	Provider        string                 `json:"provider,omitempty"`
	PromptLength    int                    `json:"prompt_length"`
	ResponseLength  int                    `json:"response_length,omitempty"`
	Stream          bool                   `json:"stream"`
	Temperature     *float64               `json:"temperature,omitempty"`
	MaxTokens       *int                   `json:"max_tokens,omitempty"`
	UserPreferences map[string]interface{} `json:"user_preferences,omitempty"`
	CostBudget      *float64               `json:"cost_budget,omitempty"`
	ComplianceReqs  []string               `json:"compliance_requirements,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

type StrategyGoal string

const (
	GoalPerformanceFirst StrategyGoal = "performance_first"
	GoalPriceFirst       StrategyGoal = "price_first"
	GoalReliabilityFirst StrategyGoal = "reliability_first"
	GoalBalanced         StrategyGoal = "balanced"
	GoalSecurityFirst    StrategyGoal = "security_first"
	GoalAuto             StrategyGoal = "auto"
)

type StrategyWeightsV2 struct {
	StrategyCode      string  `json:"strategy_code,omitempty"`
	LatencyWeight     float64 `json:"latency_weight"`
	CostWeight        float64 `json:"cost_weight"`
	ReliabilityWeight float64 `json:"reliability_weight"`
	SecurityWeight    float64 `json:"security_weight"`
	LoadBalanceWeight float64 `json:"load_balance_weight"`
}

type StrategyConstraints struct {
	MaxLatencyMs      int      `json:"max_latency_ms,omitempty"`
	MaxCostPerToken   float64  `json:"max_cost_per_token,omitempty"`
	MinSuccessRate    float64  `json:"min_success_rate,omitempty"`
	RequiredRegions   []string `json:"required_regions,omitempty"`
	ExcludedProviders []string `json:"excluded_providers,omitempty"`
	MinSecurityLevel  string   `json:"min_security_level,omitempty"`
}

type StrategyOutput struct {
	Goal        StrategyGoal        `json:"goal"`
	Weights     StrategyWeightsV2   `json:"weights"`
	Constraints StrategyConstraints `json:"constraints"`
	Priority    int                 `json:"priority"`
	Reason      string              `json:"reason"`
}

type RequestContext struct {
	MerchantID      int                    `json:"merchant_id"`
	UserID          int                    `json:"user_id,omitempty"`
	RequestAnalysis *RequestAnalysis       `json:"request_analysis"`
	UserPreferences map[string]interface{} `json:"user_preferences,omitempty"`
	CostBudget      *float64               `json:"cost_budget,omitempty"`
	ComplianceReqs  []string               `json:"compliance_requirements,omitempty"`
	Region          string                 `json:"region,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}
