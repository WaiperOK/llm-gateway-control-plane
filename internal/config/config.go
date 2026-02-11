package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

// TeamConfig represents tenant-specific gateway limits and permissions.
type TeamConfig struct {
	Name              string   `json:"name"`
	APIKey            string   `json:"api_key"`
	AllowedModels     []string `json:"allowed_models"`
	RequestsPerMinute int      `json:"requests_per_minute"`
	MonthlyBudgetUSD  float64  `json:"monthly_budget_usd"`
}

// Config is runtime gateway configuration.
type Config struct {
	ListenAddr      string             `json:"listen_addr"`
	DefaultModel    string             `json:"default_model"`
	MaxAuditEvents  int                `json:"max_audit_events"`
	BlockedPatterns []string           `json:"blocked_patterns"`
	PricingPer1KUSD map[string]float64 `json:"pricing_per_1k_usd"`
	Teams           []TeamConfig       `json:"teams"`
}

// Default returns a safe local-first configuration.
func Default() Config {
	return Config{
		ListenAddr:     ":8080",
		DefaultModel:   "gpt-4o-mini",
		MaxAuditEvents: 5000,
		BlockedPatterns: []string{
			`(?i)ignore\s+all\s+previous\s+instructions`,
			`(?i)reveal\s+system\s+prompt`,
			`(?i)exfiltrate\s+secrets?`,
			`(?i)bypass\s+policy`,
		},
		PricingPer1KUSD: map[string]float64{
			"gpt-4o-mini":       0.0030,
			"gpt-4.1-mini":      0.0045,
			"claude-3-5-sonnet": 0.0060,
		},
		Teams: []TeamConfig{
			{
				Name:              "red-team",
				APIKey:            "demo-red-key",
				AllowedModels:     []string{"gpt-4o-mini", "gpt-4.1-mini"},
				RequestsPerMinute: 60,
				MonthlyBudgetUSD:  75,
			},
			{
				Name:              "blue-team",
				APIKey:            "demo-blue-key",
				AllowedModels:     []string{"gpt-4o-mini", "claude-3-5-sonnet"},
				RequestsPerMinute: 30,
				MonthlyBudgetUSD:  40,
			},
		},
	}
}

// Load returns env-overridden config. GATEWAY_TEAMS_JSON and GATEWAY_PRICING_JSON
// allow full replacement for teams/pricing.
func Load() Config {
	cfg := Default()

	if v := os.Getenv("GATEWAY_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("GATEWAY_DEFAULT_MODEL"); v != "" {
		cfg.DefaultModel = v
	}
	if v := os.Getenv("GATEWAY_MAX_AUDIT_EVENTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.MaxAuditEvents = n
		}
	}
	if v := os.Getenv("GATEWAY_TEAMS_JSON"); v != "" {
		var teams []TeamConfig
		if err := json.Unmarshal([]byte(v), &teams); err != nil {
			log.Printf("invalid GATEWAY_TEAMS_JSON, using defaults: %v", err)
		} else if len(teams) > 0 {
			cfg.Teams = teams
		}
	}
	if v := os.Getenv("GATEWAY_PRICING_JSON"); v != "" {
		pricing := make(map[string]float64)
		if err := json.Unmarshal([]byte(v), &pricing); err != nil {
			log.Printf("invalid GATEWAY_PRICING_JSON, using defaults: %v", err)
		} else if len(pricing) > 0 {
			cfg.PricingPer1KUSD = pricing
		}
	}

	return cfg
}
