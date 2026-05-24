# Advisory Substrate Specification

**Version:** 0.1.0  
**Status:** Normative  
**Layer:** Advisory (Layer 1)

This document specifies the behavioral contract for the advisory layer of the Agent Coordination Substrate. The advisory layer provides indirect coordination between autonomous agents through environment-mediated signals. Advisory primitives shape attention without blocking operations.

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [RFC 2119](https://www.ietf.org/rfc/rfc2119.txt).

---

## 1. Overview

The advisory layer enables agents to:

1. **Deposit** typed, mortal signals into named zones
2. **Observe** accumulated signal pressure in zones
3. **React** to pressure patterns (voluntarily — advisory never blocks)
4. **Induce** artifacts when independent signals converge past a threshold

Advisory primitives satisfy all six design properties defined in [design-principles.md](../docs/design-principles.md): mortality, actor identity, zone addressing, inspectability, overridability (N/A — advisory never blocks, so nothing to override), and composability.

---

## 2. Signal Envelope

A signal is the fundamental deposit unit. Every signal MUST conform to the following envelope:

### 2.1 Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Globally unique identifier for this signal |
| `zone` | string | Target zone path (see §3) |
| `kind` | string | Signal type classifier (implementation-defined vocabulary) |
| `actor_id` | string | Identity of the depositing actor (see §4) |
| `created_at_unix` | integer | Deposit timestamp (Unix seconds) |
| `expires_at_unix` | integer | Mandatory expiry timestamp (Unix seconds) |

### 2.2 Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `strength_milli` | integer | Signal strength in milli-units (0–1000). Default: 1000 |
| `properties` | object | Structured metadata (implementation-defined schema) |
| `source` | string | Origin surface identifier (e.g., "agent_tool", "http", "mcp") |
| `reason` | string | Human-readable deposit rationale |
| `session_id` | string | Originating session (when actor operates within sessions) |
| `workflow_id` | string | Originating workflow (when actor operates within workflows) |

### 2.3 Mortality Contract

Every signal MUST have a finite TTL. The substrate MUST enforce this:

- `expires_at_unix` MUST be greater than `created_at_unix`
- The substrate MUST NOT accept a signal without `expires_at_unix`
- The substrate MUST garbage-collect signals after `expires_at_unix` passes
- There is no maximum TTL defined by this spec, but implementations SHOULD document their maximum

### 2.4 Strength Semantics

Signal strength (`strength_milli`) represents the depositor's confidence or intensity:

- Range: 0–1000 (milli-units, where 1000 = full strength)
- Default when omitted: 1000
- The substrate MAY use strength in pressure calculations
- A signal with `strength_milli: 0` is valid (represents a deliberate "zero-weight" deposit for tracking purposes)

### 2.5 Deposit Validation

A conforming implementation MUST reject deposits that:

- Omit any required field
- Have `expires_at_unix <= created_at_unix`
- Have an empty `zone` string
- Have an empty `actor_id` string

A conforming implementation SHOULD reject deposits that:

- Have `strength_milli` outside the range 0–1000

### 2.6 Capacity Protection

Implementations MAY reject deposits when internal capacity limits are reached. When rejecting for capacity, the implementation MUST return error reason `capacity_exceeded`. Implementations SHOULD document their capacity limits.

---

## 3. Zone Addressing

Zones provide locality for signals. All signals exist within a named zone.

### 3.1 Zone Name Format

- Zone names MUST be non-empty strings
- Zone names SHOULD use path-like hierarchical format (e.g., `backend/auth`, `internal/db/migrations`)
- Forward slash (`/`) is the RECOMMENDED hierarchy separator
- Leading and trailing slashes SHOULD be normalized away (e.g., `/backend/auth/` → `backend/auth`)

### 3.2 Zone Glob Matching

Readout queries MUST support glob-pattern matching:

- `*` matches any single path segment
- `**` matches zero or more path segments
- Exact match: `backend/auth` matches only `backend/auth`
- Wildcard: `backend/*` matches `backend/auth`, `backend/api`, but not `backend/auth/tokens`
- Recursive: `backend/**` matches `backend/auth`, `backend/auth/tokens`, `backend`

### 3.3 Zone Lifecycle

- Zones are implicitly created when the first signal targets them
- Zones MAY be implicitly removed when all signals in them expire (implementation choice)
- There is no "create zone" operation — zones exist as long as signals reference them

---

## 4. Actor Identity

Every deposit MUST identify its actor.

### 4.1 Actor Model

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `actor_id` | string | Yes | Stable identifier for the actor within a coordination scope |
| `actor_kind` | string | No | Category: `agent`, `operator`, `remote_agent`, `mcp_client`, `http_client`, `system` |

### 4.2 Identity Semantics

- `actor_id` MUST be stable for the duration of an actor's participation in coordination (minimally, a session or connection lifetime)
- The substrate MUST include `actor_id` in all readouts containing signal data
- The substrate MUST support querying signals by `actor_id`
- Two signals from the same `actor_id` in the same zone within the same time window represent reinforcement, not separate observations (relevant for lineage counting)

---

## 5. Zone Pressure Readout

The primary read surface. Returns the current coordination state of a zone.

### 5.1 Readout Fields

A zone pressure readout MUST include:

| Field | Type | Description |
|-------|------|-------------|
| `zone` | string | The zone this readout describes |
| `pressure` | object | Current pressure state (see §5.2) |
| `trend` | object | Temporal trend (see §5.3) |

A zone pressure readout SHOULD include:

| Field | Type | Description |
|-------|------|-------------|
| `deposit_count` | integer | Active (non-expired) signal count |
| `session_count` | integer | Distinct sessions that deposited |
| `independent_lineages` | integer | Distinct actors that deposited independently |
| `spec_zone_glob` | string | The glob pattern that matched this zone |
| `contributing_signals` | array | Bounded list of active signals |
| `evidence_truncated` | boolean | Whether the signal list was capped |

### 5.2 Pressure States

The pressure state vocabulary:

| State | Meaning |
|-------|---------|
| `quiet` | No significant signal activity |
| `active` | Signals present, below concern thresholds |
| `warming` | Accumulating toward a threshold |
| `convergent` | Multiple independent actors depositing same classification |
| `saturated` | Threshold reached — induction candidate |
| `contested` | Conflicting signals from different actors |
| `avoidable` | Active avoidance signals (negative coordination) |
| `boundary_sensitive` | Near a governance boundary, elevated attention warranted |

Implementations MUST preserve the meaning of these states when emitting them. Implementations MAY add additional states prefixed with `x_` (extension states).

### Minimum Classification Thresholds

Implementations MUST classify zone pressure using at least these boundaries:

| Condition | Minimum State |
|-----------|--------------|
| Zero live signals in zone | `quiet` |
| 1 independent lineage (actor) | `active` |
| 2 independent lineages | `warming` |
| 3+ independent lineages | `convergent` |

Implementations MAY use finer-grained classifications or additional states beyond this minimum. The `score` field is implementation-defined but MUST be monotonically non-decreasing with deposit count and lineage count.

### 5.3 Trend States

Temporal context comparing recent activity to a baseline:

| State | Meaning |
|-------|---------|
| `current` | Activity within the short window |
| `recent` | Activity within the medium window but not the short window |
| `cooling` | Activity declining from prior peak |
| `surging` | Activity increasing rapidly |
| `stale` | No recent activity; signals aging toward expiry |

Trend is advisory temporal context. Trend states MUST NOT themselves trigger mutation or enforcement.

A trend object SHOULD include:

| Field | Type | Description |
|-------|------|-------------|
| `state` | string | One of the trend states above |
| `current_deposit_count` | integer | Deposits in the short window |
| `recent_deposit_count` | integer | Deposits in the medium window |
| `short_window_seconds` | integer | Short window duration |
| `medium_window_seconds` | integer | Medium window duration |
| `newest_signal_created_at_unix` | integer | Most recent deposit time |
| `oldest_signal_created_at_unix` | integer | Oldest active deposit time |

### 5.4 Pressure Score

Implementations SHOULD compute a numeric pressure score:

- Range: 0.0–1.0
- Incorporates: deposit count, lineage count, signal strength, recency
- The formula is implementation-defined
- Score is advisory — it MUST NOT trigger enforcement

---

## 6. Induction

Induction is the mechanism by which accumulated advisory signals produce a new artifact when convergence criteria are met.

### 6.1 Saturation Specification

A saturation spec defines the threshold for induction:

| Field | Type | Description |
|-------|------|-------------|
| `zone_glob` | string | Target zone(s) for this saturation rule |
| `threshold_deposits` | integer | Minimum deposit count to trigger |
| `threshold_lineages` | integer | Minimum independent actor count to trigger |
| `window_seconds` | integer | Time window within which deposits must accumulate |
| `kind_filter` | string | Optional: only count signals of this kind |

### 6.2 Manifest Proposal

When saturation criteria are met, the substrate SHOULD generate a manifest proposal:

| Field | Type | Description |
|-------|------|-------------|
| `manifest_id` | string | Unique identifier for this proposal |
| `zone` | string | The zone where saturation was detected |
| `signal_ids` | array | The signals that triggered saturation |
| `confidence` | number | Derived from signal strength and lineage diversity |
| `proposed_at_unix` | integer | When the manifest was generated |
| `expires_at_unix` | integer | When the proposal expires if not acted upon |

### 6.3 Proposal Lifecycle

A manifest proposal transitions through states:

```
pending → accepted → canary → active → retired
                 ↘ dismissed
```

- **pending**: Generated by induction, awaiting operator or agent action
- **accepted**: Explicitly promoted by an authorized actor
- **canary**: Under observation period after acceptance
- **active**: Fully promoted and in effect
- **retired**: Expired or explicitly decommissioned
- **dismissed**: Explicitly rejected by an authorized actor

Transitions between states MUST be explicit actions (not implicit side effects of readout). The substrate MUST NOT automatically promote proposals without explicit authorization.

### 6.4 Induction Invariants

- Induction MUST NOT block any operation (advisory layer guarantee)
- Induction MUST require convergence from multiple independent actors (never from a single actor's repeated deposits)
- Generated proposals MUST be mortal (have `expires_at_unix`)
- Proposal acceptance MUST be an explicit command, never automatic

---

## 7. Readout Surface Requirements

A conforming implementation MUST expose the following read operations:

### 7.1 Zone List

List zones matching a glob pattern. Returns zone pressure readouts for all matching zones.

**Input:** zone glob pattern (or empty for all zones)  
**Output:** array of zone pressure readouts (§5)

### 7.2 Zone Explain

Detailed readout for a single zone including all available context.

**Input:** zone name  
**Output:** zone pressure readout + contributing signals + active manifest proposals + causal context

### 7.3 Zone Triage

Priority-ordered list of zones requiring attention.

**Input:** optional limit, sort criteria, zone filter  
**Output:** ordered list of triage items with priority, zone, recommended action, and evidence

### 7.4 Readout Invariants

- Read operations MUST NOT mutate substrate state
- Read operations MUST NOT trigger induction as a side effect
- Read operations MUST include actor identity for all signals returned
- Read operations MUST indicate when output is truncated (`evidence_truncated`, `truncated`)

---

## 8. Deposit Operations

### 8.1 Deposit Signal

Create a new signal in the substrate.

**Input:** signal envelope (§2)  
**Output:** success confirmation with assigned `id` (if not provided by caller)

### 8.2 Reinforce Signal

Extend the TTL of an existing signal (by the same actor).

**Input:** `signal_id`, `actor_id`, new `expires_at_unix`  
**Output:** success or error (signal not found, actor mismatch)

An actor MAY only reinforce their own signals.

### 8.3 Kill Signal

Immediately expire a signal (by the same actor or an authorized operator).

**Input:** `signal_id`, `actor_id`  
**Output:** success or error

- An actor MAY kill their own signals
- An operator MAY kill any signal (operator override)

---

## 9. Persistence

### 9.1 Persistence Guarantee

A conforming implementation at Level 1 or higher MUST persist signals across process restart:

- Signals deposited before a crash MUST be readable after restart (if not expired)
- Expired signals MAY be garbage-collected during restart
- Active manifest proposals MUST survive restart

### 9.2 Crash Recovery

After crash recovery, the substrate MUST:

1. Load persisted signals
2. Garbage-collect signals where `expires_at_unix < now`
3. Restore zone pressure state from surviving signals
4. Resume induction evaluation

---

## 10. Multi-Surface Access

The advisory substrate MUST be accessible from multiple surfaces:

### 10.1 Required Surfaces

A conforming implementation MUST provide at least ONE of:

- HTTP REST API
- Agent tool (callable within an agent loop)
- CLI command

### 10.2 Recommended Surfaces

A conforming implementation SHOULD provide:

- HTTP REST API for remote deposits and readouts
- Agent tool for in-session deposits and readouts
- CLI for operator inspection
- MCP tool for IDE/editor integration
- Programmatic library API for embedding

### 10.3 Surface Parity

All surfaces MUST expose the same signal data. A signal deposited via HTTP MUST be visible in agent tool readouts, and vice versa. Parity applies to:

- Signal visibility (deposit from any surface, read from any surface)
- Zone pressure (same pressure state regardless of which surface reads)
- Induction (saturation considers signals from all surfaces together)

---

## 11. Conformance Levels

### Level 1: Signal Store

The implementation provides signal deposit, zone pressure readout, and persistence.

**Requirements:**
- Signal envelope validation (§2.5)
- Zone addressing with glob matching (§3)
- Actor identity on all deposits (§4)
- Zone pressure readout with pressure and trend states (§5)
- Signal mortality enforcement (garbage collection)
- Persistence across restart (§9)

### Level 2: Outcome-Aware

The implementation persists recommendation snapshots and command events, then exposes recommendation outcomes without writing from read paths.

**Additional requirements beyond Level 1:**
- Recommendation packet generation (evidence-bound suggestions)
- Outcome ledger (snapshot + command event correlation)
- Calibration metadata (empirical trust from outcome history)
- Read operations MUST NOT append to the outcome ledger

### Level 3: Thread-Aware

The implementation exposes bounded causal threads connecting related zones.

**Additional requirements beyond Level 2:**
- Causal thread derivation from shared evidence
- Thread summary with freshness, stale/orphan flags
- Thread membership tracking across zones

### Level 4: Induction-Capable

The implementation supports saturation-triggered manifest generation.

**Additional requirements beyond Level 1:**
- Saturation specification evaluation (§6.1)
- Manifest proposal generation (§6.2)
- Proposal lifecycle state machine (§6.3)
- Multi-actor convergence requirement (§6.4)

---

## 12. Non-Goals

The advisory layer explicitly does NOT:

- Block, gate, or constrain any operation (that is enforcement — see the enforcement spec)
- Provide direct agent-to-agent messaging (that is A2A)
- Provide tool execution (that is MCP)
- Schedule, assign, or route work (that is orchestration)
- Replace persistence layers (it coordinates, not stores)

---

## Appendix A: Naming Conventions

| Normative term | NOT used in this spec |
|----------------|----------------------|
| signal | pheromone |
| zone | trail, namespace |
| deposit | emit |
| readout | sniff |
| pressure | intensity |
| mortality | decay (except when describing half-life math) |
| actor | holder, owner |
| induction | emergence (in normative text) |

Biology-metaphor terms are acknowledged as intellectual lineage in [docs/theory.md](../docs/theory.md) but MUST NOT appear in normative specification text.
