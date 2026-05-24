# Enforcement Layer Conformance

This directory contains testable behavioral expectations for enforcement layer conformance. Any implementation claiming conformance MUST pass all expectations at its declared level.

## Structure

- `expectations.yaml` — Machine-readable test scenarios organized by conformance level
- See `../../reference/conformance-runner/` for the executable runner

## How to Use

1. Implement the enforcement substrate per `../SPEC.md`
2. Write an adapter that exposes your implementation's operations to the conformance runner
3. Run the conformance runner against your adapter
4. Report your conformance level based on which expectation groups pass

## Conformance Levels

| Level | Required Groups |
|-------|----------------|
| Level 1: Basic Claims | `acquire`, `release`, `ttl_expiry`, `hard_conflict`, `soft_contention`, `override`, `persistence` |
| Level 2: Contention-Aware | Level 1 + `contention_report`, `negotiation`, `mixed_mode` |
| Level 3: Remote-Capable | Level 1 + `remote_ttl_validation`, `surface_parity` |
