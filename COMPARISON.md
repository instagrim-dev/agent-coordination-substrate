# Comparison

How the Agent Coordination Substrate relates to existing and emerging alternatives.

Last updated: 2026-05-24

---

## Capability Matrix

| Dimension | Agent Coordination Substrate | SBP v0.2 | S-MADRL (arxiv 2510.03592) | MCP | A2A |
|-----------|------------------------------|-----------|----------------------------|-----|-----|
| **Advisory signals** | ✓ Named zones, typed signals, structured metadata | ✓ Decaying pheromones, durable traces | ✓ Virtual pheromones in grid | ✗ | ✗ |
| **Signal mortality (TTL)** | ✓ Per-signal mandatory TTL, garbage collection | ✓ Exponential decay | ✓ Fixed decay rate | N/A | N/A |
| **Zone addressing** | ✓ Hierarchical, glob-matchable | ✓ Board-level | ✗ (grid coordinates) | N/A | N/A |
| **Multi-signal induction** | ✓ Threshold triggers → manifested artifacts | ✗ | ✗ | ✗ | ✗ |
| **Enforcement primitives (claims)** | ✓ Mortal, scoped, actor-identified | ✗ | ✗ | ✗ | ✗ |
| **Enforcement gating** | ✓ Operations blocked until claim expires/releases | ✗ | ✗ | ✗ | ✗ |
| **Operator override** | ✓ All enforcement is operator-dismissable | N/A | N/A | N/A | N/A |
| **Persistence (crash-safe)** | ✓ SQLite-backed signal store | ✗ In-memory only | ✗ In-memory only | N/A | N/A |
| **Operator cognition readout** | ✓ Human-readable substrate state | ✗ | ✗ | ✗ | ✗ |
| **HTTP ingress for external signals** | ✓ REST endpoint for depositing signals | ✗ | ✗ | N/A | N/A |
| **Cross-session coordination** | ✓ Signals persist across session boundaries | ✗ Single session | ✗ Single episode | ✗ | ✗ |
| **Actor identity tracking** | ✓ Mandatory on all deposits and claims | Partial (agent_id on deposits) | ✗ Anonymous | N/A | N/A |
| **Composable with A2A/MCP** | ✓ Orthogonal layer | N/A | N/A | ✓ (is the tool layer) | ✓ (is the agent layer) |
| **Saturation semantics** | ✓ Configurable thresholds per zone | ✗ | ✗ | N/A | N/A |
| **Conformance test suite** | ✓ Corpus-backed | ✗ | ✗ | ✗ | ✗ |
| **Programmable capture rules** | ✓ Pattern-matched → actions | ✗ | ✗ | ✗ | ✗ |
| **Production deployment** | ✓ Shipping in multi-agent product | ✗ Prototype | ✗ Research | N/A | N/A |

---

## Timeline

Priority evidence. Post-April-28 dates verifiable via commit SHA. Earlier dates evidenced by plan files with `status: completed` frontmatter (commit history was squashed during a 2026-04-29 repo maintenance event).

| Date | Milestone | Evidence |
|------|-----------|----------|
| **2026-03-19** | **Shared fleet memory** — coordination precursor (fleet-wide knowledge propagation) | Plan file: `2026-03-19-002-feat-shared-fleet-memory-plan.md` (status: completed) |
| **2026-03-25** | **Architecture research: workspace observer + stigmergic coordination** | Architecture doc: `docs/design/2026-03-25-learned-niche-health-spawn-workspace-observer-architecture.md` |
| **2026-03-28** | **Workspace observer heat-map + stigmergy** design (Colony/Flock taxonomy) | Plan file: `2026-03-28-002-feat-workspace-observer-heatmap-stigmergy-plan.md` (status: completed) |
| **2026-03-31** | **Environmental substrate MVP shipped with tests** | Plan file: `2026-03-31-012-feat-bounded-environment-mvp-shipped-tests-plan.md` (status: completed) |
| **2026-04-06** | **Gradient fields** (continuous decay) + **territorial leases** (scoped claims) designed | Plan files: `2026-04-06-001`, `2026-04-06-008`, `2026-04-06-014` (gradient completed, leases active) |
| **2026-04-17** | Multi-session trail persistence + observability | Plan files: `2026-04-17-003`, `2026-04-17-004` |
| **2026-04-25** | Shared workspace concurrency master architecture | Plan files: `2026-04-25-shared-workspace-concurrency-*` |
| 2026-05-01 | **Enforcement: workspace claim registry** — first claim primitive with tests, shutdown cleanup, TUI visibility | Commit: `feat(workspaceclaim): W1 claim registry` |
| 2026-05-03 | Claim negotiation surfaces: multi-agent contention | Commit: `feat(workspace): add claim negotiation surfaces` |
| 2026-05-04 | Coordination visibility and contention diagnostics | Commit: `feat(workspacecoord): add shared-workspace coordination layer foundation` |
| 2026-05-08 | **Advisory: workspace stigmergy cell** — first persistent signal store | Commit: `workspace-stigmergy: create cell package with types and state` |
| 2026-05-12 | Cross-workflow executor with workspace stigmergy integration | Commit: `feat(workflow): add cross workflow executor` |
| 2026-05-16 | **Advisory: programmable signal fabric** — zones, TTL, typed signals | Commit: `feat(signals): add programmable signal fabric` |
| 2026-05-18 | Signal saturation, compound synthesis, cross-workflow convergence | Multiple commits |
| 2026-05-19 | Programmable capture rules (induction: signals → manifested actions) | Commit: `feat: add programmable capture rules` |
| 2026-05-21 | Read parity surfaces (HTTP, MCP, agent tool) | Commit: `feat(stigmergy): add read parity surfaces` |
| 2026-05-23 | MCP explain/triage tools, operator cognition | Commit: `feat: add MCP stigmergy tools` |
| 2026-05-24 | Full enforcement tier formalization | Internal meta-design document |

### External Timeline (For Context)

| Date | Project | Event |
|------|---------|-------|
| 2025-10-06 | S-MADRL | Arxiv submission (2510.03592) |
| **2026-03-25** | **This spec** | Architecture research: workspace observer + stigmergic coordination |
| **2026-03-31** | **This spec** | Environmental substrate MVP shipped with tests |
| 2026-05-17 | SBP | First commit on `AdviceNXT/sbp` (GitHub) |
| 2026-05-19 | SBP | v0.2 release |

---

## Key Differentiators

### 1. Enforcement Tier

No alternative provides an enforcement layer. SBP and S-MADRL are advisory-only: agents observe environmental state and decide whether to respect it. This spec includes claims that **block tool execution** — with mandatory mortality (always expire), actor identity (attribution), and operator override (human can dismiss).

### 2. Induction (Threshold → Artifact)

When multiple independent signals accumulate in a zone past a configurable threshold, the substrate can **induce** a new artifact — materializing an obligation, recommendation, or constraint. No alternative implements this emergent-from-convergence mechanism.

### 3. Persistence

Signals survive process restart. This is critical for multi-session coordination where agents may not share process lifetimes. SBP is explicitly in-memory; S-MADRL is episodic.

### 4. Operator Cognition

The substrate provides structured readouts designed for human operators to understand coordination state. This bridges the "observability gap" that purely agent-to-agent coordination creates.

### 5. Production Maturity

The reference implementation has been in active development since March 2026 (architecture research and environmental substrate MVP shipped March 25–31), with enforcement claims in production since May 1 and continuous hardening thereafter: race condition fixes, conformance corpus expansion, performance optimization, and API evolution. Alternatives are prototype or research-stage.

---

## What This Spec Is Not

- **Not a replacement for MCP** — MCP provides tool execution. This spec provides coordination between tool-using agents.
- **Not a replacement for A2A** — A2A provides direct agent-to-agent communication. This spec provides indirect coordination through the environment.
- **Not an orchestrator** — This spec does not schedule, assign, or route work. It provides the coordination substrate that orchestrators and autonomous agents alike can use.

---

## Relationship to Stigmergy

The intellectual heritage of this work includes stigmergy (Grassé 1959, Theraulaz & Bonabeau 1999) — coordination through environmental modification rather than direct communication. We acknowledge this lineage in [docs/theory.md](docs/theory.md).

However, this specification uses **engineering-native vocabulary** (signals, zones, claims, enforcement, mortality) rather than biological metaphor (pheromones, scent, trail). The engineering vocabulary maps precisely to distributed systems concepts, enables clear normative specification, and avoids the misleading implication that agents are simple reactive entities.
