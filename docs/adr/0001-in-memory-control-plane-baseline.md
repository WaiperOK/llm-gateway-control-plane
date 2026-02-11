# ADR-0001: In-memory Control Plane Baseline

## Status
Accepted

## Context

The first milestone requires a demonstrable and testable control plane with safety and cost controls, but without external infrastructure dependencies.

## Decision

Use in-memory implementations for:

- usage ledger
- audit event store
- rate-limiting counters

and a deterministic local model client.

## Consequences

Positive:

- quick bootstrap
- deterministic integration tests
- no secrets required for local demo

Negative:

- no durability
- horizontal scaling requires shared state backend
- not suitable for strict compliance retention requirements

## Follow-up

Replace in-memory modules with PostgreSQL/Redis adapters in v0.2.
