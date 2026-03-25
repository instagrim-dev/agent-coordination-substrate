# Agent Coordination Substrate

A specification for environment-mediated coordination between autonomous agents.

Two layers, one substrate: **advisory signals** shape attention without blocking; **enforcement claims** gate operations with mandatory mortality. Together they provide the governance layer that multi-agent workspaces need — without centralized orchestration, without permanent locks, and without coupling agents to each other.

---

## Why This Exists

Multiple AI agents working in the same workspace need coordination. Today the options are:

- **Direct messaging** (A2A, MCP tool calls) — works for delegation, but creates coupling
- **Centralized orchestrators** — single point of failure, doesn't scale to autonomous agents
- **Nothing** — agents overwrite each other's work

The coordination substrate provides a third path: **indirect coordination through the environment**. Agents deposit signals and claims into shared zones. Other agents observe the environment and react. No agent needs to know about any other agent. Coordination emerges from the substrate, not from wiring.

This pattern draws from [stigmergy](docs/theory.md) — indirect coordination through environmental modification — but uses engineering-native vocabulary and adds an enforcement tier that pure stigmergic systems lack.

---

## Two Layers

| Layer | Contract | Primitives |
|-------|----------|------------|
| **Advisory** | Shapes attention. Never blocks. | Signals, zones, pressure readouts, induction |
| **Enforcement** | Gates operations. Always mortal. | Claims, quarantine zones, evidence gates, capacity reservations |

Both layers share: actor identity, zone addressing, mortality semantics, multi-surface readout.

They differ: advisory primitives never block tool execution. Enforcement primitives may block — but always expire, and operators can always override.

---

## Quick Comparison

| Capability | This spec | SBP v0.2 | MCP | A2A |
|------------|-----------|----------|-----|-----|
| Advisory coordination (signals, zones) | ✓ | ✓ | ✗ | ✗ |
| Enforcement (claims, gates) | ✓ | ✗ | ✗ | ✗ |
| Persistence (crash-safe) | ✓ | ✗ (in-memory) | N/A | N/A |
| Induction (threshold → manifest) | ✓ | ✗ | ✗ | ✗ |
| Governance lifecycle (override, audit) | ✓ | ✗ | ✗ | ✗ |
| Operator cognition readout | ✓ | ✗ | ✗ | ✗ |

See [COMPARISON.md](COMPARISON.md) for the full breakdown with timeline evidence.

---

## Getting Started

### Try the reference implementation

```bash
cd reference/conformance-runner
go run ./cmd/conformance-runner ../../
# → 42 passed, 0 failed, 9 skipped
```

### Build your own implementation

Follow the [Adoption Guide](docs/adoption-guide.md) — start with an advisory signal store (~100 lines), add enforcement when you need workspace claims.

### Validate conformance

The [conformance runner](reference/conformance-runner/) validates any `SignalStore` + `ClaimStore` implementation against the spec's expectations.

---

## Reading Order

1. **[docs/design-principles.md](docs/design-principles.md)** — The six properties every primitive must satisfy
2. **[advisory/SPEC.md](advisory/SPEC.md)** — Advisory layer normative specification
3. **[enforcement/SPEC.md](enforcement/SPEC.md)** — Enforcement layer normative specification
4. **[docs/theory.md](docs/theory.md)** — Intellectual lineage (stigmergy, distributed leases, resilience patterns)
5. **[docs/adoption-guide.md](docs/adoption-guide.md)** — How to implement in your agent framework
6. **[reference/](reference/)** — Go reference implementation and conformance runner

---

## Repository Structure

```
├── advisory/
│   ├── SPEC.md                    # Normative specification
│   ├── schemas/                   # JSON Schema (Draft 2020-12)
│   ├── conformance/               # Test expectations (YAML)
│   └── examples/                  # Example payloads
├── enforcement/
│   ├── SPEC.md                    # Normative specification
│   ├── schemas/                   # JSON Schema (Draft 2020-12)
│   ├── conformance/               # Test expectations (YAML)
│   └── examples/                  # Example payloads
├── docs/
│   ├── adoption-guide.md          # Implementation walkthrough
│   ├── design-principles.md       # Six mandatory properties
│   └── theory.md                  # Intellectual lineage
├── reference/
│   ├── go/                        # Reference implementation (Go, zero deps)
│   └── conformance-runner/        # Executable conformance validator
├── COMPARISON.md                  # vs. SBP, S-MADRL, MCP, A2A
├── CHANGELOG.md                   # Dated development history
├── CONTRIBUTING.md                # Contribution guidelines
└── LICENSE                        # Apache 2.0 + CC BY 4.0
```

---

## Reference Implementation

The [reference implementation](reference/go/) is a minimal, standalone Go package that proves the spec is implementable. Zero external dependencies. Extracted from the [BMO](https://github.com/instagrim-dev/bmo) production runtime — the first and most complete implementation of this specification.

**Key files:**
- `interfaces.go` — `SignalStore` and `ClaimStore` (conformance targets)
- `signal_store.go` — In-memory advisory store with mortality GC
- `claim_store.go` — In-memory enforcement store with conflict detection
- `zone.go` — Zone glob matching

---

## Status

**v0.1.0** — Initial publication. Advisory layer normative. Enforcement layer normative for claims; other enforcement primitives (quarantine zones, evidence gates, capacity reservations) are informative.

---

## License

- **Specification content:** [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- **Reference implementation, schemas, and tooling:** [Apache 2.0](LICENSE)

---

## Contributing

Contributions welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
