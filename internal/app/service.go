package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/WaiperOK/llm-gateway-control-plane/internal/audit"
	"github.com/WaiperOK/llm-gateway-control-plane/internal/auth"
	"github.com/WaiperOK/llm-gateway-control-plane/internal/billing"
	"github.com/WaiperOK/llm-gateway-control-plane/internal/config"
	"github.com/WaiperOK/llm-gateway-control-plane/internal/policy"
	"github.com/WaiperOK/llm-gateway-control-plane/internal/ratelimit"
	"github.com/WaiperOK/llm-gateway-control-plane/internal/redaction"
	"github.com/WaiperOK/llm-gateway-control-plane/pkg/contracts"
)

// AppError represents a typed API-level error.
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *AppError) Error() string { return e.Message }

// Service holds business logic for gateway operations.
type Service struct {
	logger       *slog.Logger
	auth         *auth.APIKeyAuth
	policy       *policy.Engine
	limiter      *ratelimit.Limiter
	billing      *billing.Service
	audit        *audit.Store
	metrics      *Metrics
	modelClient  ModelClient
	defaultModel string
}

func NewService(cfg config.Config, logger *slog.Logger, metrics *Metrics, modelClient ModelClient) *Service {
	teamDescriptors := make([]auth.TeamDescriptor, 0, len(cfg.Teams))
	for _, t := range cfg.Teams {
		teamDescriptors = append(teamDescriptors, auth.TeamDescriptor{
			Team:              t.Name,
			APIKey:            t.APIKey,
			AllowedModels:     t.AllowedModels,
			RequestsPerMinute: t.RequestsPerMinute,
			MonthlyBudgetUSD:  t.MonthlyBudgetUSD,
		})
	}

	return &Service{
		logger:       logger,
		auth:         auth.NewAPIKeyAuth(teamDescriptors),
		policy:       policy.NewEngine(cfg.BlockedPatterns),
		limiter:      ratelimit.NewLimiter(),
		billing:      billing.NewService(cfg.PricingPer1KUSD),
		audit:        audit.NewStore(cfg.MaxAuditEvents),
		metrics:      metrics,
		modelClient:  modelClient,
		defaultModel: cfg.DefaultModel,
	}
}

func (s *Service) Authenticate(r *http.Request) (auth.Principal, *AppError) {
	principal, err := s.auth.Authenticate(r)
	if err == nil {
		return principal, nil
	}
	if errors.Is(err, auth.ErrMissingAPIKey) {
		return auth.Principal{}, &AppError{Code: "missing_api_key", Message: err.Error(), HTTPStatus: http.StatusUnauthorized}
	}
	return auth.Principal{}, &AppError{Code: "invalid_api_key", Message: err.Error(), HTTPStatus: http.StatusUnauthorized}
}

func (s *Service) HandleCompletion(
	ctx context.Context,
	requestID string,
	principal auth.Principal,
	req contracts.CompletionRequest,
) (contracts.CompletionResponse, *AppError) {
	start := time.Now()
	model := req.Model
	if model == "" {
		model = s.defaultModel
	}
	status := "ok"

	track := func(inputTokens, outputTokens int, cost float64) {
		s.metrics.RequestsTotal.WithLabelValues(principal.Team, model, status).Inc()
		s.metrics.LatencySec.WithLabelValues(principal.Team, model, status).Observe(time.Since(start).Seconds())
		if inputTokens > 0 {
			s.metrics.TokensTotal.WithLabelValues(principal.Team, model, "input").Add(float64(inputTokens))
		}
		if outputTokens > 0 {
			s.metrics.TokensTotal.WithLabelValues(principal.Team, model, "output").Add(float64(outputTokens))
		}
		if cost > 0 {
			s.metrics.CostTotalUSD.WithLabelValues(principal.Team, model).Add(cost)
		}
	}

	if req.Input == "" {
		status = "bad_request"
		track(0, 0, 0)
		return contracts.CompletionResponse{}, &AppError{Code: "invalid_input", Message: "input is required", HTTPStatus: http.StatusBadRequest}
	}
	if len([]rune(req.Input)) > 32000 {
		status = "bad_request"
		track(0, 0, 0)
		return contracts.CompletionResponse{}, &AppError{Code: "input_too_large", Message: "input exceeds 32000 characters", HTTPStatus: http.StatusBadRequest}
	}

	redacted := redaction.Scrub(req.Input)
	decision := s.policy.Evaluate(policy.Input{Model: model, Prompt: req.Input, AllowedModels: principal.AllowedModels})
	if !decision.Allowed {
		status = "denied_policy"
		s.audit.Add(audit.Event{
			Timestamp:     time.Now().UTC(),
			RequestID:     requestID,
			Team:          principal.Team,
			Model:         model,
			Status:        status,
			DenyReason:    decision.Reason,
			RedactedInput: redacted.Text,
			LatencyMS:     time.Since(start).Milliseconds(),
		})
		track(0, 0, 0)
		return contracts.CompletionResponse{}, &AppError{Code: "policy_denied", Message: decision.Reason, HTTPStatus: http.StatusForbidden}
	}

	if allowed := s.limiter.Allow(principal.Team, principal.RequestsPerMinute, time.Now()); !allowed {
		status = "rate_limited"
		s.audit.Add(audit.Event{
			Timestamp:     time.Now().UTC(),
			RequestID:     requestID,
			Team:          principal.Team,
			Model:         model,
			Status:        status,
			DenyReason:    "requests_per_minute_exceeded",
			RedactedInput: redacted.Text,
			LatencyMS:     time.Since(start).Milliseconds(),
		})
		track(0, 0, 0)
		return contracts.CompletionResponse{}, &AppError{Code: "rate_limited", Message: "requests_per_minute_exceeded", HTTPStatus: http.StatusTooManyRequests}
	}

	inputTokens := billing.ApproxTokens(req.Input)
	estimatedCost := s.billing.EstimateCost(model, inputTokens, 120)
	if !s.billing.CanAfford(principal.Team, principal.MonthlyBudgetUSD, estimatedCost) {
		status = "budget_exceeded"
		s.audit.Add(audit.Event{
			Timestamp:     time.Now().UTC(),
			RequestID:     requestID,
			Team:          principal.Team,
			Model:         model,
			Status:        status,
			DenyReason:    "estimated_cost_exceeds_budget",
			RedactedInput: redacted.Text,
			LatencyMS:     time.Since(start).Milliseconds(),
		})
		track(0, 0, 0)
		return contracts.CompletionResponse{}, &AppError{Code: "budget_exceeded", Message: "estimated_cost_exceeds_budget", HTTPStatus: http.StatusPaymentRequired}
	}

	output, err := s.modelClient.Complete(ctx, model, req.Input)
	if err != nil {
		status = "upstream_error"
		s.logger.Error("model completion failed", "request_id", requestID, "team", principal.Team, "err", err)
		s.audit.Add(audit.Event{
			Timestamp:     time.Now().UTC(),
			RequestID:     requestID,
			Team:          principal.Team,
			Model:         model,
			Status:        status,
			DenyReason:    "upstream_completion_failed",
			RedactedInput: redacted.Text,
			LatencyMS:     time.Since(start).Milliseconds(),
		})
		track(inputTokens, 0, 0)
		return contracts.CompletionResponse{}, &AppError{Code: "upstream_error", Message: "upstream_completion_failed", HTTPStatus: http.StatusBadGateway}
	}

	outputTokens := billing.ApproxTokens(output)
	cost := s.billing.EstimateCost(model, inputTokens, outputTokens)
	if !s.billing.CanAfford(principal.Team, principal.MonthlyBudgetUSD, cost) {
		status = "budget_exceeded"
		s.audit.Add(audit.Event{
			Timestamp:     time.Now().UTC(),
			RequestID:     requestID,
			Team:          principal.Team,
			Model:         model,
			Status:        status,
			DenyReason:    "actual_cost_exceeds_budget",
			RedactedInput: redacted.Text,
			LatencyMS:     time.Since(start).Milliseconds(),
		})
		track(inputTokens, outputTokens, 0)
		return contracts.CompletionResponse{}, &AppError{Code: "budget_exceeded", Message: "actual_cost_exceeds_budget", HTTPStatus: http.StatusPaymentRequired}
	}

	s.billing.Record(principal.Team, model, inputTokens, outputTokens, cost)
	s.audit.Add(audit.Event{
		Timestamp:     time.Now().UTC(),
		RequestID:     requestID,
		Team:          principal.Team,
		Model:         model,
		Status:        status,
		RedactedInput: redacted.Text,
		CostUSD:       cost,
		LatencyMS:     time.Since(start).Milliseconds(),
	})
	track(inputTokens, outputTokens, cost)

	return contracts.CompletionResponse{
		RequestID:      requestID,
		Team:           principal.Team,
		Model:          model,
		Output:         output,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		CostUSD:        cost,
		PolicyDecision: "allow",
		ProcessedAt:    time.Now().UTC(),
	}, nil
}

func (s *Service) Usage(principal auth.Principal) contracts.UsageResponse {
	u := s.billing.GetUsage(principal.Team)
	remaining := s.billing.RemainingBudget(principal.Team, principal.MonthlyBudgetUSD)
	return contracts.UsageResponse{
		Team:               principal.Team,
		TotalRequests:      u.TotalRequests,
		TotalInputTokens:   u.TotalInputTokens,
		TotalOutputTokens:  u.TotalOutputTokens,
		TotalCostUSD:       u.TotalCostUSD,
		MonthlyBudgetUSD:   principal.MonthlyBudgetUSD,
		RemainingBudgetUSD: remaining,
		PerModel:           u.PerModelCostUSD,
	}
}

func (s *Service) AuditEvents(principal auth.Principal, limit int) []contracts.AuditEventView {
	events := s.audit.List(principal.Team, limit)
	out := make([]contracts.AuditEventView, 0, len(events))
	for _, ev := range events {
		out = append(out, contracts.AuditEventView{
			Timestamp:     ev.Timestamp,
			RequestID:     ev.RequestID,
			Team:          ev.Team,
			Model:         ev.Model,
			Status:        ev.Status,
			DenyReason:    ev.DenyReason,
			RedactedInput: ev.RedactedInput,
			CostUSD:       ev.CostUSD,
			LatencyMS:     ev.LatencyMS,
		})
	}
	return out
}

func (e *AppError) WithRequestID(requestID string) contracts.ErrorResponse {
	return contracts.ErrorResponse{Error: e.Message, Code: e.Code, RequestID: requestID}
}

func NewInternalError(err error) *AppError {
	return &AppError{
		Code:       "internal_error",
		Message:    fmt.Sprintf("internal_error: %v", err),
		HTTPStatus: http.StatusInternalServerError,
	}
}
