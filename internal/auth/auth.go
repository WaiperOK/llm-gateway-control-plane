package auth

import (
	"errors"
	"net/http"
	"strings"
)

var (
	ErrMissingAPIKey = errors.New("missing api key")
	ErrInvalidAPIKey = errors.New("invalid api key")
)

// Principal is an authenticated gateway caller.
type Principal struct {
	Team              string
	AllowedModels     map[string]struct{}
	RequestsPerMinute int
	MonthlyBudgetUSD  float64
}

// TeamDescriptor is used to construct API key auth map.
type TeamDescriptor struct {
	Team              string
	APIKey            string
	AllowedModels     []string
	RequestsPerMinute int
	MonthlyBudgetUSD  float64
}

// APIKeyAuth authenticates callers by API key.
type APIKeyAuth struct {
	byKey map[string]Principal
}

func NewAPIKeyAuth(teams []TeamDescriptor) *APIKeyAuth {
	byKey := make(map[string]Principal, len(teams))
	for _, t := range teams {
		models := make(map[string]struct{}, len(t.AllowedModels))
		for _, m := range t.AllowedModels {
			models[m] = struct{}{}
		}
		byKey[t.APIKey] = Principal{
			Team:              t.Team,
			AllowedModels:     models,
			RequestsPerMinute: t.RequestsPerMinute,
			MonthlyBudgetUSD:  t.MonthlyBudgetUSD,
		}
	}
	return &APIKeyAuth{byKey: byKey}
}

func (a *APIKeyAuth) Authenticate(r *http.Request) (Principal, error) {
	key := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if key == "" {
		authz := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			key = strings.TrimSpace(authz[len("Bearer "):])
		}
	}
	if key == "" {
		return Principal{}, ErrMissingAPIKey
	}
	principal, ok := a.byKey[key]
	if !ok {
		return Principal{}, ErrInvalidAPIKey
	}
	return principal, nil
}
