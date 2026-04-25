package services

type CostEstimation struct {
	APIKeyID      int
	MerchantID    int
	EstimatedCost float64
	InputCost     float64
	OutputCost    float64
	InputPrice    float64
	OutputPrice   float64
}

type CostEstimationService struct {
	tokenService *TokenEstimationService
}

func NewCostEstimationService() *CostEstimationService {
	return &CostEstimationService{
		tokenService: NewTokenEstimationService(),
	}
}

func (s *CostEstimationService) CalculateEstimatedCost(candidate *RoutingCandidate, estimation *TokenEstimation) *CostEstimation {
	inputCost := estimation.EstimatedInputTokens * candidate.InputPrice
	outputCost := estimation.EstimatedOutputTokens * candidate.OutputPrice
	estimatedCost := inputCost + outputCost

	return &CostEstimation{
		APIKeyID:      candidate.APIKeyID,
		MerchantID:    candidate.MerchantID,
		EstimatedCost: estimatedCost,
		InputCost:     inputCost,
		OutputCost:    outputCost,
		InputPrice:    candidate.InputPrice,
		OutputPrice:   candidate.OutputPrice,
	}
}

func (s *CostEstimationService) CalculatePriceScores(candidates []RoutingCandidate, estimation *TokenEstimation) map[int]float64 {
	costs := make(map[int]float64)
	for i := range candidates {
		costEstimation := s.CalculateEstimatedCost(&candidates[i], estimation)
		costs[candidates[i].APIKeyID] = costEstimation.EstimatedCost
	}

	min, max := minMax(costs)

	scores := make(map[int]float64)
	for keyID, cost := range costs {
		if max == min {
			scores[keyID] = 1.0
		} else {
			scores[keyID] = 1.0 - (cost-min)/(max-min)
		}
	}

	return scores
}

func minMax(values map[int]float64) (min, max float64) {
	if len(values) == 0 {
		return 0, 0
	}

	first := true
	for _, v := range values {
		if first {
			min = v
			max = v
			first = false
			continue
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}
