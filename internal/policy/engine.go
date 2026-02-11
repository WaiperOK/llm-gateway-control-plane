package policy

import "regexp"

// Decision is the result of a policy evaluation.
type Decision struct {
	Allowed bool
	Reason  string
}

// Input carries all context required for policy checks.
type Input struct {
	Model         string
	Prompt        string
	AllowedModels map[string]struct{}
}

// Engine evaluates request policy.
type Engine struct {
	blocked []*regexp.Regexp
}

func NewEngine(patterns []string) *Engine {
	blocked := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		blocked = append(blocked, re)
	}
	return &Engine{blocked: blocked}
}

func (e *Engine) Evaluate(in Input) Decision {
	if _, ok := in.AllowedModels[in.Model]; !ok {
		return Decision{Allowed: false, Reason: "model_not_allowed_for_team"}
	}
	for _, re := range e.blocked {
		if re.MatchString(in.Prompt) {
			return Decision{Allowed: false, Reason: "blocked_pattern_detected"}
		}
	}
	return Decision{Allowed: true}
}
