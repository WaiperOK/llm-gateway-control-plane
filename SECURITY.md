# Security Policy

## Supported versions

This repository is pre-1.0. Security fixes are provided on the `main` branch.

## Reporting a vulnerability

Please report vulnerabilities privately by opening a GitHub Security Advisory draft.

Include:

- impact summary
- reproduction steps
- affected endpoints/files
- suggested fix (if available)

## Security controls in this repository

- API key authentication
- tenant-scoped model allowlist
- per-team rate limiting
- budget enforcement
- PII redaction in audit trail
- metrics and audit records for anomaly investigation

## Hardening checklist for production

- move API keys to managed secrets (Vault/KMS)
- enforce TLS termination and mTLS in service mesh
- protect `/metrics` behind internal auth/network policy
- persist usage/audit in encrypted datastore
- enable key rotation and key expiry automation
