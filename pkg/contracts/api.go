package contracts

import "time"

// CompletionRequest is a normalized request accepted by the gateway.
type CompletionRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// CompletionResponse is returned for successful requests.
type CompletionResponse struct {
	RequestID      string    `json:"request_id"`
	Team           string    `json:"team"`
	Model          string    `json:"model"`
	Output         string    `json:"output"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	CostUSD        float64   `json:"cost_usd"`
	PolicyDecision string    `json:"policy_decision"`
	ProcessedAt    time.Time `json:"processed_at"`
}

// ErrorResponse is used for policy/rate/budget and validation errors.
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}

// UsageResponse returns current team usage and budget state.
type UsageResponse struct {
	Team               string             `json:"team"`
	TotalRequests      int64              `json:"total_requests"`
	TotalInputTokens   int64              `json:"total_input_tokens"`
	TotalOutputTokens  int64              `json:"total_output_tokens"`
	TotalCostUSD       float64            `json:"total_cost_usd"`
	MonthlyBudgetUSD   float64            `json:"monthly_budget_usd"`
	RemainingBudgetUSD float64            `json:"remaining_budget_usd"`
	PerModel           map[string]float64 `json:"per_model_cost_usd"`
}

// AuditEventView is a scrubbed view returned by audit API.
type AuditEventView struct {
	Timestamp     time.Time `json:"timestamp"`
	RequestID     string    `json:"request_id"`
	Team          string    `json:"team"`
	Model         string    `json:"model"`
	Status        string    `json:"status"`
	DenyReason    string    `json:"deny_reason,omitempty"`
	RedactedInput string    `json:"redacted_input"`
	CostUSD       float64   `json:"cost_usd"`
	LatencyMS     int64     `json:"latency_ms"`
}
