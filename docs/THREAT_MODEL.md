# Threat Model

## Scope

`llm-gateway-control-plane` request path, tenant auth, policy enforcement, audit data, and observability endpoints.

## Assets

- tenant API keys
- prompt payloads
- audit trail
- usage/cost counters
- policy configuration

## Trust boundaries

- external clients -> gateway API
- gateway -> model provider client
- gateway -> monitoring stack

## Primary threats and mitigations

1. Unauthorized usage
- Threat: leaked/guessed API key
- Mitigations: key-based auth, per-team quotas, audit trail
- Future: key rotation API + HMAC request signing

2. Prompt injection / policy bypass
- Threat: malicious prompts to override instructions
- Mitigations: deny-pattern policy checks before model call
- Future: context-aware policy engine and model-side moderation

3. Cost abuse
- Threat: request floods and expensive model selection
- Mitigations: RPM limiter + per-team budget guard + model allowlist
- Future: adaptive risk scoring and anomaly detection

4. Sensitive data exposure in logs
- Threat: PII in audit/diagnostic events
- Mitigations: redaction pipeline before audit storage
- Future: structured field-level encryption

5. Metrics endpoint information leakage
- Threat: unauthenticated scraping in hostile network
- Mitigations: run behind private network / ingress auth in prod
- Future: mTLS for internal observability traffic

## Residual risk

- Regex policies can miss semantic bypasses.
- In-memory storage has no durability.
- API keys in env are operationally sensitive and require secret management.
