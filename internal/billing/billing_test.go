package billing

import "testing"

func TestEstimateAndRecordUsage(t *testing.T) {
	svc := NewService(map[string]float64{"model-a": 0.01})
	cost := svc.EstimateCost("model-a", 100, 50)
	if cost <= 0 {
		t.Fatalf("expected cost > 0, got %f", cost)
	}
	svc.Record("team-a", "model-a", 100, 50, cost)
	u := svc.GetUsage("team-a")
	if u.TotalRequests != 1 {
		t.Fatalf("expected 1 request, got %d", u.TotalRequests)
	}
	if u.TotalInputTokens != 100 || u.TotalOutputTokens != 50 {
		t.Fatalf("unexpected token counters: %+v", u)
	}
}

func TestBudgetCheck(t *testing.T) {
	svc := NewService(map[string]float64{"model-a": 0.01})
	svc.Record("team-a", "model-a", 1000, 0, 0.01)
	if svc.CanAfford("team-a", 0.01, 0.001) {
		t.Fatal("expected budget exceed")
	}
}
