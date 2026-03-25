# v0.1.0 — Initial Publication

The first public release of the Agent Coordination Substrate specification.

## What's Included

### Specification

- **Advisory Layer** (`advisory/SPEC.md`) — normative specification for signals,
  zones, pressure readouts, mortality, and induction
- **Enforcement Layer** (`enforcement/SPEC.md`) — normative specification for claims
  (hard/soft modes, conflict detection, operator override); informative appendix for
  future primitives (quarantine zones, evidence gates, capacity reservations)

### Schemas (JSON Schema Draft 2020-12)

- Signal envelope, zone pressure, saturation spec, manifest proposal (advisory)
- Claim, actor identity, enforcement primitive interface (enforcement)

### Conformance Suite

- 24 advisory expectations across 6 groups
- 27 enforcement expectations across 9 groups
- YAML-driven, implementation-agnostic

### Reference Implementation (Go)

- Zero external dependencies
- `SignalStore` and `ClaimStore` interfaces as conformance targets
- In-memory stores with full spec compliance
- Conformance runner CLI: `42 passed, 0 failed, 9 skipped`

### Documentation

- Design principles (6 mandatory properties)
- Theoretical lineage (stigmergy, distributed leases, resilience engineering)
- Adoption guide (step-by-step implementation walkthrough)
- Comparison vs. SBP, S-MADRL, MCP, A2A
- Changelog with commit-verifiable dates

## Conformance Levels

| Level | Layer | Groups |
|-------|-------|--------|
| 1 — Signal Store | Advisory | deposit, mortality, zone_addressing, pressure_readout |
| 2 — Persistent | Advisory | + persistence |
| 3 — Induction-Capable | Advisory | + induction |
| 4 — Basic Claims | Enforcement | acquire, release, ttl_expiry, hard_conflict, soft_contention, override, persistence |
| 5 — Contention-Aware | Enforcement | + mixed_mode |
| 6 — Remote-Capable | Enforcement | + remote_ttl_validation |

## Getting Started

```bash
cd reference/conformance-runner
go run ./cmd/conformance-runner ../../
```

Or follow the [Adoption Guide](docs/adoption-guide.md) to build your own implementation.

## Origin

Extracted from the [BMO](https://github.com/instagrim-dev/bmo) production runtime,
which has implemented this substrate since January 2025 (advisory) and March 2025
(enforcement). See [CHANGELOG.md](CHANGELOG.md) for the full development timeline.
