package router

type Request struct {
	Tenant         string `json:"tenant"`
	TaskClass      string `json:"task_class"` // "latency" | "quality"
	EstimatedTokens int   `json:"estimated_tokens"`
}

type TenantBudget struct {
	MonthlyTokenLimit int
	UsedTokens        int
}

type RouteDecision struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Allowed  bool   `json:"allowed"`
	Reason   string `json:"reason"`
}

func Decide(req Request, budget TenantBudget) RouteDecision {
	if budget.UsedTokens+req.EstimatedTokens > budget.MonthlyTokenLimit {
		return RouteDecision{
			Provider: "",
			Model:    "",
			Allowed:  false,
			Reason:   "budget_exceeded",
		}
	}
	if req.TaskClass == "quality" {
		return RouteDecision{Provider: "openai", Model: "gpt-5", Allowed: true, Reason: "quality_path"}
	}
	return RouteDecision{Provider: "openai", Model: "gpt-5-mini", Allowed: true, Reason: "latency_path"}
}
