# llm-gateway-control-plane

Budget-aware model routing core for multi-provider LLM gateway design.

## Architecture Intent

Go service decides provider/model route by tenant policy, token budget, and criticality class.

## Implemented Vertical Slice

- Deterministic core decision engine
- Small API/CLI surface for integration
- Unit tests for regression safety
- CI workflow for every push/PR

## Quickstart

```bash
go run ./cmd/server
```

## Tests

```bash
go test ./...
```
