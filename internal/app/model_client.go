package app

import (
	"context"
	"fmt"
	"strings"
)

// ModelClient abstracts LLM provider integrations.
type ModelClient interface {
	Complete(ctx context.Context, model, input string) (string, error)
}

// SimulatedModelClient provides deterministic local responses for demos and tests.
type SimulatedModelClient struct{}

func (SimulatedModelClient) Complete(_ context.Context, model, input string) (string, error) {
	normalized := strings.TrimSpace(input)
	if len(normalized) > 180 {
		normalized = normalized[:180] + "..."
	}
	return fmt.Sprintf("[%s] triage summary: request accepted; key risks extracted from input: %s", model, normalized), nil
}
