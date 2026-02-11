# Architecture

## Context

The gateway acts as a safety and governance layer between product teams and LLM providers.

## Components

1. `transport/httpapi`
- request parsing
- auth enforcement
- JSON response shaping

2. `auth`
- API-key authentication
- principal resolution (team, allowlist, limits)

3. `policy`
- deny patterns for prompt-injection and secret exfiltration intents
- model allowlist check per team

4. `ratelimit`
- per-team requests-per-minute windowed limiter

5. `billing`
- token approximation
- model pricing table
- budget gate + usage ledger

6. `audit`
- bounded in-memory event store
- redacted payload capture for compliance and incident review

7. `redaction`
- PII scrubbing (email/phone/IP)

8. `app`
- orchestration of request lifecycle
- unified decision pipeline: auth -> policy -> quota -> budget -> model -> accounting

## Request lifecycle

1. Auth principal from API key.
2. Validate input payload.
3. Evaluate policy.
4. Apply team RPM limit.
5. Budget pre-check with estimated output tokens.
6. Call model client.
7. Final cost accounting.
8. Audit event write and metrics emission.

## Data model

- Team config: API key, allowed models, RPM, monthly budget.
- Usage ledger: requests/tokens/cost totals + per-model cost.
- Audit event: request id, team, model, status, deny reason, redacted input, cost, latency.

## Tradeoffs

- In-memory audit and usage simplify deployment; production extension should persist in PostgreSQL.
- Token estimation is intentionally approximate to keep integration provider-agnostic.
- Simulated model client keeps deterministic tests and no secret requirement for local runs.
