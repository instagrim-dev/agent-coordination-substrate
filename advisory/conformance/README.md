# Advisory Layer Conformance

This directory contains testable behavioral expectations for advisory layer conformance. Any implementation claiming conformance to this spec MUST pass all expectations at its declared conformance level.

## Structure

- `expectations.yaml` — Machine-readable test scenarios organized by conformance level
- See `../../reference/conformance-runner/` for the executable runner

## How to Use

1. Implement the advisory substrate per `../SPEC.md`
2. Write an adapter that exposes your implementation's operations to the conformance runner
3. Run the conformance runner against your adapter
4. Report your conformance level based on which expectation groups pass

## Conformance Levels

| Level | Required Groups |
|-------|----------------|
| Level 1: Signal Store | `deposit`, `mortality`, `zone_addressing`, `pressure_readout`, `persistence` |
| Level 2: Outcome-Aware | Level 1 + `recommendation`, `outcome_ledger`, `calibration` |
| Level 3: Thread-Aware | Level 2 + `causal_threads` |
| Level 4: Induction-Capable | Level 1 + `induction` |
