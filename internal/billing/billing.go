package billing

import (
	"math"
	"strings"
	"sync"
)

// TeamUsage is an aggregated billing view.
type TeamUsage struct {
	TotalRequests     int64
	TotalInputTokens  int64
	TotalOutputTokens int64
	TotalCostUSD      float64
	PerModelCostUSD   map[string]float64
}

// Service stores usage counters and pricing metadata.
type Service struct {
	mu      sync.Mutex
	pricing map[string]float64
	usage   map[string]*TeamUsage
}

func NewService(pricing map[string]float64) *Service {
	copyPricing := make(map[string]float64, len(pricing))
	for k, v := range pricing {
		copyPricing[k] = v
	}
	return &Service{
		pricing: copyPricing,
		usage:   make(map[string]*TeamUsage),
	}
}

func ApproxTokens(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// Approximation: 1 token ~= 4 chars for mixed English text.
	n := int(math.Ceil(float64(len([]rune(s))) / 4.0))
	if n < 1 {
		return 1
	}
	return n
}

func (s *Service) UnitPrice(model string) float64 {
	if p, ok := s.pricing[model]; ok {
		return p
	}
	return 0.005
}

func (s *Service) EstimateCost(model string, inputTokens, outputTokens int) float64 {
	pricePer1K := s.UnitPrice(model)
	return float64(inputTokens+outputTokens) / 1000.0 * pricePer1K
}

func (s *Service) RemainingBudget(team string, monthlyBudgetUSD float64) float64 {
	u := s.GetUsage(team)
	remaining := monthlyBudgetUSD - u.TotalCostUSD
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *Service) CanAfford(team string, monthlyBudgetUSD, estimatedCost float64) bool {
	return s.RemainingBudget(team, monthlyBudgetUSD) >= estimatedCost
}

func (s *Service) Record(team, model string, inputTokens, outputTokens int, cost float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u := s.usage[team]
	if u == nil {
		u = &TeamUsage{PerModelCostUSD: make(map[string]float64)}
		s.usage[team] = u
	}

	u.TotalRequests++
	u.TotalInputTokens += int64(inputTokens)
	u.TotalOutputTokens += int64(outputTokens)
	u.TotalCostUSD += cost
	u.PerModelCostUSD[model] += cost
}

func (s *Service) GetUsage(team string) TeamUsage {
	s.mu.Lock()
	defer s.mu.Unlock()

	u := s.usage[team]
	if u == nil {
		return TeamUsage{PerModelCostUSD: map[string]float64{}}
	}
	copyPerModel := make(map[string]float64, len(u.PerModelCostUSD))
	for k, v := range u.PerModelCostUSD {
		copyPerModel[k] = v
	}
	return TeamUsage{
		TotalRequests:     u.TotalRequests,
		TotalInputTokens:  u.TotalInputTokens,
		TotalOutputTokens: u.TotalOutputTokens,
		TotalCostUSD:      u.TotalCostUSD,
		PerModelCostUSD:   copyPerModel,
	}
}
