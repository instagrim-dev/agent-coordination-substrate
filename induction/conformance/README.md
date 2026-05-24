# Induction Conformance Tests

This directory contains the conformance expectations for the induction layer.

## Structure

- `expectations.yaml` — Scenario definitions for induction conformance

## Running

The induction layer requires an advisory signal store (Layer 1) as its foundation. A conformance runner for induction MUST:

1. Provide a conforming `SignalStore` (advisory Layer 1)
2. Provide an `InductionEngine` implementation under test
3. Execute scenarios from `expectations.yaml` in order

## Actions

| Action | Description |
|--------|-------------|
| `configure_saturation` | Register a saturation specification |
| `deposit` | Deposit a signal into the advisory store (triggers evaluation) |
| `advance_time` | Advance the test clock |
| `list_proposals` | List manifest proposals (optionally filtered by zone) |
| `accept_proposal` | Promote a pending proposal |
| `dismiss_proposal` | Reject a pending proposal |

## Expectations

Each scenario's `expect` block defines the required outcome. Implementations MUST match all fields present in the expectation. Fields not present in the expectation are not validated (allowing implementation-defined extensions).
