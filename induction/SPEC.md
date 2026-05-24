# Induction Substrate Specification

**Version:** 0.1.0  
**Status:** Normative  
**Layer:** Induction (Layer 3)  
**Depends on:** Advisory Substrate (Layer 1)

This document specifies the behavioral contract for the induction layer of the Agent Coordination Substrate. Induction is the mechanism by which accumulated advisory signals autonomously produce new artifacts when multi-actor convergence criteria are met — closing the bootstrapping gap for systems that cannot pre-configure rules for situations they have not yet encountered.

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [RFC 2119](https://www.ietf.org/rfc/rfc2119.txt).

---

## 1. Overview

The induction layer enables:

1. **Observation** — monitoring zone pressure for threshold convergence
2. **Saturation detection** — recognizing when independent actors converge past configured criteria
3. **Manifest generation** — producing artifact proposals from detected convergence
4. **Lifecycle governance** — explicit acceptance, dismissal, and retirement of proposals

Induction operates on the advisory signal store (Layer 1). It reads signals and zone pressure but never blocks operations. The induction layer satisfies all six design properties: mortality (proposals expire), actor identity (generation traced to triggering actors), zone addressing (proposals reference zones), inspectability (saturation state readable), overridability (operators dismiss proposals), and composability (proposals are signals themselves).

---

## 2. Dependency on Advisory Layer

The induction layer REQUIRES a conforming advisory signal store (Layer 1, Level 1 minimum). Induction:

- Reads signals from the advisory store to evaluate saturation
- Counts independent lineages (actors) from the advisory store
- Deposits manifest signals back into the advisory store (at a staging zone)
- Uses zone glob matching from the advisory spec (§3)

An implementation MAY provide induction without enforcement (Layer 2). Induction and enforcement are independent extensions of the advisory base.

---

## 3. Saturation Specification

A saturation spec defines the threshold conditions for induction.

### 3.1 Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone_glob` | string | Target zone(s) for this saturation rule. Supports glob patterns per advisory §3.2. |
| `threshold_deposits` | integer | Minimum active deposit count to trigger (>= 1) |
| `threshold_lineages` | integer | Minimum independent actor count to trigger (>= 2, enforced) |
| `window_seconds` | integer | Time window within which deposits must accumulate (>= 1) |

### 3.2 Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier for this saturation rule |
| `kind_filter` | string | Only count signals of this kind toward saturation |
| `strength_minimum_milli` | integer | Only count signals with strength at or above this threshold (0–1000) |
| `enabled` | boolean | Whether this rule is active (default: true) |
| `manifest_ttl_seconds` | integer | TTL for generated proposals (default: implementation-defined) |
| `staging_zone` | string | Zone where generated manifest signals are deposited |
| `min_sessions` | integer | Minimum distinct sessions (when session tracking is available) |

### 3.3 Multi-Actor Requirement

A conforming implementation MUST enforce `threshold_lineages >= 2`. A single actor's repeated deposits MUST NEVER trigger induction regardless of deposit count. This is the fundamental invariant that distinguishes induction from simple threshold alerting.

### 3.4 Evaluation Semantics

- The implementation MUST evaluate saturation after each deposit into a matching zone
- Evaluation considers only signals within the `window_seconds` time window ending at the current time
- Signals that have expired (past `expires_at_unix`) MUST NOT count toward saturation
- If `kind_filter` is set, only signals matching that kind are counted
- If `strength_minimum_milli` is set, only signals at or above that threshold are counted
- Two signals from the same `actor_id` count as one lineage (per advisory §4.2)

### 3.5 Default-Off

Induction MUST be default-off. If no saturation specs are configured, no induction evaluation runs. This ensures zero overhead for deployments that do not use induction.

---

## 4. Manifest Proposal

When saturation criteria are met, the implementation generates a manifest proposal.

### 4.1 Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `manifest_id` | string | Globally unique identifier for this proposal |
| `zone` | string | The zone where saturation was detected |
| `state` | string | Current lifecycle state (see §5) |
| `signal_ids` | array[string] | The signals that triggered saturation |
| `confidence` | number | Derived from signal strength and lineage diversity (0.0–1.0) |
| `proposed_at_unix` | integer | When the manifest was generated (Unix seconds) |
| `expires_at_unix` | integer | When the proposal expires if not acted upon (Unix seconds) |

### 4.2 Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `saturation_spec_id` | string | The saturation rule that generated this proposal |
| `actor_count` | integer | Number of independent actors whose signals contributed |
| `accepted_at_unix` | integer | When accepted (if applicable) |
| `accepted_by` | string | Actor who accepted (if applicable) |
| `dismissed_at_unix` | integer | When dismissed (if applicable) |
| `dismissed_by` | string | Actor who dismissed (if applicable) |
| `dismiss_reason` | string | Reason for dismissal (if applicable) |
| `retired_at_unix` | integer | When retired (if applicable) |
| `payload` | object | Implementation-defined proposal content (e.g., suggested rule template) |

### 4.3 Mortality Contract

Every manifest proposal MUST have a finite TTL:

- `expires_at_unix` MUST be greater than `proposed_at_unix`
- The implementation MUST garbage-collect proposals after `expires_at_unix` passes
- Operator-ignored proposals evaporate — this is correct behavior, not an error

### 4.4 Confidence Calculation

The `confidence` field is implementation-defined but SHOULD incorporate:

- Lineage count relative to `threshold_lineages`
- Signal strength (average `strength_milli` of contributing signals)
- Temporal density (signals clustered vs. spread across the window)

Higher confidence means stronger convergence evidence. Implementations MUST keep confidence in the range 0.0–1.0.

---

## 5. Proposal Lifecycle

A manifest proposal transitions through states:

```
pending → accepted → canary → active → retired
                 ↘ dismissed
```

### 5.1 State Definitions

| State | Meaning | Terminal |
|-------|---------|---------|
| `pending` | Generated by induction, awaiting action | No |
| `accepted` | Explicitly promoted by an authorized actor | No |
| `canary` | Under observation period after acceptance | No |
| `dismissed` | Explicitly rejected by an authorized actor | Yes |
| `active` | Fully promoted and in effect | No |
| `retired` | Expired or explicitly decommissioned | Yes |

### 5.2 Transition Rules

- `pending → accepted`: Requires explicit `accept_proposal` command from an authorized actor
- `pending → dismissed`: Requires explicit `dismiss_proposal` command
- `accepted → canary`: Implementation-defined (MAY be automatic with configured observation period)
- `accepted → active`: If no canary period configured, MAY transition directly
- `canary → active`: After observation period passes without issues
- `canary → dismissed`: If issues detected during canary
- `active → retired`: Explicit decommission or TTL expiry
- `pending → retired`: TTL expiry without action (implicit)

### 5.3 Invariants

- The substrate MUST NOT automatically promote proposals from `pending` to `accepted` without explicit authorization
- All state transitions (except TTL-based retirement) MUST be explicit commands
- The substrate MUST record the actor identity for all explicit transitions (audit trail)
- A `dismissed` or `retired` proposal MUST NOT transition to any other state

---

## 6. Operations

### 6.1 Configure Saturation

Register or update a saturation specification.

**Input:** saturation spec (§3)  
**Output:** success confirmation

### 6.2 List Proposals

List manifest proposals, optionally filtered by zone or state.

**Input:** optional zone filter, optional state filter  
**Output:** array of manifest proposals

### 6.3 Accept Proposal

Promote a pending proposal.

**Input:** `manifest_id`, `actor_id`  
**Output:** success (new state: `accepted`) or error (not found, wrong state, unauthorized)

### 6.4 Dismiss Proposal

Reject a pending proposal.

**Input:** `manifest_id`, `actor_id`, optional `reason`  
**Output:** success (new state: `dismissed`) or error (not found, wrong state, unauthorized)

### 6.5 Retire Proposal

Decommission an active proposal.

**Input:** `manifest_id`, `actor_id`  
**Output:** success (new state: `retired`) or error

---

## 7. Compound Synthesis (OPTIONAL)

An implementation MAY provide cross-session synthesis as an induction accelerator.

### 7.1 Semantics

When N or more distinct agent sessions deposit signals in the same zone within a configured window, the implementation MAY generate a compound signal that contributes to saturation evaluation. This elevates zones that attract independent session attention.

### 7.2 Requirements (when provided)

- Compound signals MUST be deposited back into the advisory store (they are regular signals with a distinct `source` marker)
- Compound signals MUST NOT trigger self-reinforcing loops (the implementation MUST filter signals with the compound `source` from compound evaluation)
- Session boundaries MUST be respected (compound synthesis MUST NOT cross workspace or partition boundaries)
- The session threshold MUST be configurable (default: 2)

---

## 8. Persistence

### 8.1 Proposal Persistence

Active manifest proposals (states: `pending`, `accepted`, `canary`, `active`) MUST survive process restart:

- Proposals present before a crash MUST be recoverable after restart
- Expired proposals MAY be garbage-collected during restart
- Saturation specifications MUST persist across restart

### 8.2 Crash Recovery

After crash recovery, the implementation MUST:

1. Load persisted proposals
2. Garbage-collect proposals where `expires_at_unix < now`
3. Resume saturation evaluation with existing signal state

---

## 9. Induction Invariants (Summary)

1. Induction MUST NOT block any operation (advisory layer guarantee inherited)
2. Induction MUST require convergence from multiple independent actors (§3.3)
3. Generated proposals MUST be mortal (§4.3)
4. Proposal acceptance MUST be an explicit command, never automatic (§5.3)
5. Induction MUST be default-off (§3.5)
6. All proposal state transitions MUST be auditable (§5.3)
7. A single actor MUST NOT be able to force induction regardless of deposit volume (§3.3)

---

## 10. Conformance

A conforming induction implementation MUST:

- Implement all operations in §6
- Satisfy all invariants in §9
- Pass all scenarios in [conformance/expectations.yaml](conformance/expectations.yaml)
- Depend on a Level 1 conforming advisory signal store

An implementation MAY additionally provide compound synthesis (§7) — this is not required for induction conformance.

---

## 11. Non-Goals

The induction layer explicitly does NOT:

- Block, gate, or constrain any operation (that is enforcement)
- Execute or apply generated proposals (that is the consumer's responsibility)
- Replace human/operator judgment (proposals require explicit acceptance)
- Provide scheduling or cron-like evaluation (evaluation is deposit-triggered)
- Define the content format of proposal payloads (that is implementation-defined)

---

## Appendix A: Naming Conventions

| Normative term | NOT used in this spec |
|----------------|----------------------|
| induction | emergence (in normative text) |
| saturation | convergence threshold |
| manifest | recommendation, suggestion |
| proposal | draft, candidate |
| confidence | certainty, probability |

Biology-metaphor terms are acknowledged as intellectual lineage in [docs/theory.md](../docs/theory.md) but MUST NOT appear in normative specification text.

---

## Appendix B: Relationship to FNI

This specification formalizes the "Fabric-Native Induction" (FNI) pattern first implemented in BMO. Key implementation decisions from FNI that informed this spec:

- Saturation evaluation runs in an app-layer observer, not inside the signal store — keeping the store thin
- Manifest signals are deposited into a staging zone (e.g., `_fabric_staging/rules/`) — they are signals, subject to the same mortality contract
- Compound synthesis uses a `source` discriminator to prevent feedback loops
- The `propose_rule` action type provides a built-in path for proposal materialization without requiring pre-configured automation rules (solving the chicken-and-egg problem)

These are implementation strategies, not normative requirements. Alternative architectures (in-store evaluation, separate proposal store, etc.) are valid as long as they satisfy the behavioral contract above.
