package router

import "testing"

func TestQualityRoute(t *testing.T) {
	req := Request{Tenant: "t1", TaskClass: "quality", EstimatedTokens: 1000}
	budget := TenantBudget{MonthlyTokenLimit: 10_000, UsedTokens: 100}
	d := Decide(req, budget)
	if !d.Allowed || d.Model != "gpt-5" {
		t.Fatalf("unexpected decision: %+v", d)
	}
}

func TestBudgetExceeded(t *testing.T) {
	req := Request{Tenant: "t1", TaskClass: "latency", EstimatedTokens: 5000}
	budget := TenantBudget{MonthlyTokenLimit: 5000, UsedTokens: 100}
	d := Decide(req, budget)
	if d.Allowed {
		t.Fatalf("expected denied by budget")
	}
}
