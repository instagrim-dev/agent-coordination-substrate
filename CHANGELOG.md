# Changelog

Timeline of the Agent Coordination Substrate implementation. Plan files with dated filenames and `status: completed` frontmatter serve as the primary evidence trail. Post-April-28 commits are verifiable by SHA in the [BMO](https://github.com/instagrim-dev/bmo) repository (the reference implementation); earlier history was squashed during a repository maintenance event on 2026-04-29.

This changelog establishes implementation priority: the coordination substrate architecture was designed and prototyped in **March 2026** — nearly two months before the first alternative published a commit.

---

## 2026-05-24 — Enforcement Tier Formalization

- Formalize enforcement tier architecture (advisory + enforcement two-layer model)
- Define six mandatory properties for all substrate primitives
- Publish landscape transfer brief and comparison with alternatives

## 2026-05-23 — Operator Cognition MCP

- Add MCP `stigmergy_explain` and `stigmergy_triage` tools
- Operator cognition readout over recommendation ledger
- Read parity surfaces: HTTP, MCP, and agent tool access to same state

## 2026-05-22 — Conformance and Classification

- Tension manifest with conformance validation and classification
- Enhance tension manifest with conformance and evidence

## 2026-05-21 — Read Parity

- Signal fabric read parity surfaces across all access modalities
- Capability-fit readout documentation
- Signal fabric store extension with recommendation readouts

## 2026-05-20 — Recommendations Engine

- Signal fabric store extended with zone-scoped recommendations
- Evidence-backed capability-fit recommendations
- Threshold evidence deduplication

## 2026-05-19 — Programmable Capture Rules

- Pattern-matched capture rules on contract preflight surfaces
- Rules evaluate signal patterns and trigger actions
- First production use of induction: accumulated signals → manifested obligations

## 2026-05-18 — Saturation and Cross-Workflow

- **Signal saturation evaluation** — configurable thresholds per zone
- COMPOUND cross-session synthesis observer (threshold → materialized artifact)
- Cross-workflow convergence (plan 020): multi-session coordination
- Workspace leases and event fanout
- Rule manifest broker with TUI notification
- Zone+window signal count and distinct session queries

## 2026-05-16 — Programmable Signal Fabric

- **Programmable signal fabric**: typed signals, named zones, TTL, structured metadata
- First complete advisory layer implementation

## 2026-05-13 — Morphogen Decision Evidence

- Signal decision evidence contract (morphogen)
- Nanite decision evidence persistence
- Evaluation report command

## 2026-05-12 — Cross-Workflow Coordination

- Cross workflow executor with workspace stigmergy integration
- Strategy router integration

## 2026-05-08 — Workspace Stigmergy Foundation

- **Workspace stigmergy cell package**: persistent signal store (types, state, tests)
- Cell-based storage model for advisory signals

## 2026-05-06 — Claim Hardening

- Harden claim coordination boundaries
- Boundary enforcement between concurrent agents

## 2026-05-04 — Coordination Layer

- Shared-workspace coordination layer foundation
- Coordination visibility and contention diagnostics
- Diagnostic readouts for active claims

## 2026-05-03 — Claim Negotiation

- **Claim negotiation surfaces**: multi-agent contention resolution
- Structured negotiation when multiple agents claim overlapping zones

## 2026-05-01 — Enforcement Primitive (Claims)

- **Workspace claim registry** (W1): first enforcement primitive
- Tests, shutdown cleanup, TUI visibility
- Claims with actor identity, zone scope, TTL, structured metadata

---

## Pre-May Implementation (plan-file evidence; commit history squashed by 2026-04-29 repo maintenance)

### 2026-04-25 — Shared Workspace Concurrency Architecture

- `shared-workspace-concurrency-architecture-plan.md` — Master plan for multi-writer coordination
- Concurrency MVP design and workstream breakdown

### 2026-04-17 — Multi-Session Trail Persistence

- `bounded-workspace-stigmergy-multi-session-trail-plan.md` — Signals surviving session boundaries
- `workspace-trail-prometheus-sse-observability-plan.md` — External observability for trails
- `workspace-trail-second-writer-contract-validation-plan.md` — Multi-writer correctness

### 2026-04-06 — Gradient Fields and Territorial Leases

- **`workspace-environment-gradient-fields-plan.md`** (status: completed) — Continuous pheromone-style sensing with decay
- **`environment-fields-half-life-decay-plan.md`** (status: completed) — TTL/half-life mortality semantics
- **`territorial-workspace-leases-plan.md`** (status: active) — Scoped exclusive/shared claims with TTL and explicit release. Plan explicitly states: "territorial leases are optional hard gates — closer to signposts + fences than pheromones alone"

### 2026-03-31 — Environmental Substrate MVP Shipped

- **`environmental-heatmap-stigmergy-plan.md`** (status: completed) — Full heat-map + stigmergy substrate
- **`bounded-environment-mvp-shipped-tests-plan.md`** (status: completed) — **MVP with tests shipped**
- `ai-organism-constitution-checklist-plan.md` (status: completed) — Governance framework

### 2026-03-28 — Workspace Observer + Stigmergy

- **`workspace-observer-heatmap-stigmergy-plan.md`** (status: completed) — Heat-map and stigmergic hints design
- Explicit "Colony/Flock taxonomy gap" acknowledgment in plan title

### 2026-03-25 — Architecture Research (Earliest Formal Design)

- **`learned-niche-health-spawn-workspace-observer-research-plan.md`** (status: completed) — Research program
- **`docs/design/2026-03-25-learned-niche-health-spawn-workspace-observer-architecture.md`** — Architecture document satisfying research acceptance criteria
- Origin: `2026-03-25-learned-niche-workspace-observer-taxonomy-spike.md` brainstorm

### 2026-03-19 — Shared Fleet Memory (Coordination Precursor)

- **`shared-fleet-memory-plan.md`** (status: completed) — Fleet-wide shared knowledge propagation
- Infrastructure that later evolved into the coordination substrate

---

## External Timeline (For Context)

| Date | Project | Event | Notes |
|------|---------|-------|-------|
| 2025-10-06 | S-MADRL | Arxiv submission | Research paper, no implementation artifact |
| **2026-03-25** | **This spec** | **Architecture research doc** | **Formal design for workspace observer + stigmergic coordination** |
| **2026-03-31** | **This spec** | **Environmental substrate MVP shipped with tests** | **53 days before SBP's first commit** |
| **2026-04-06** | **This spec** | **Gradient fields + territorial leases designed** | **Decay semantics and enforcement (claims) designed** |
| **2026-05-01** | **This spec** | **Enforcement claims in production** | **16 days before SBP's first commit** |
| **2026-05-08** | **This spec** | **Advisory signal store in production** | **9 days before SBP's first commit** |
| 2026-05-17 | SBP | First commit | Advisory-only, in-memory, no enforcement |
| 2026-05-19 | SBP | v0.2 release | Added blackboard reads, still no enforcement |

---

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 0.1.0 | 2026-05-24 | Initial public specification |
