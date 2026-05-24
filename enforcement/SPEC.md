# Enforcement Substrate Specification

**Version:** 0.1.0  
**Status:** Normative (Claims); Informative (future primitives)  
**Layer:** Enforcement (Layer 2)

This document specifies the behavioral contract for the enforcement layer of the Agent Coordination Substrate. Enforcement primitives gate operations with mandatory mortality. Unlike advisory signals (which shape attention), enforcement primitives MAY block tool execution — but always expire, and operators can always override.

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [RFC 2119](https://www.ietf.org/rfc/rfc2119.txt).

---

## 1. Overview

The enforcement layer enables agents and operators to:

1. **Claim** exclusive or advisory access to named zones
2. **Block** operations that conflict with active claims (hard enforcement)
3. **Warn** about operations in claimed zones without blocking (soft enforcement)
4. **Release** claims explicitly, or let them expire automatically
5. **Override** any enforcement as an authorized operator

Enforcement primitives satisfy all six design properties defined in [design-principles.md](../docs/design-principles.md): mortality, actor identity, zone addressing, inspectability, overridability, and composability.

---

## 2. Enforcement Primitive Contract

Every enforcement primitive (claims and future primitives) MUST satisfy these shared properties:

| Property | Requirement |
|----------|-------------|
| **Mortality** | MUST have a finite TTL. No permanent enforcement. |
| **Actor identity** | MUST identify the actor that created it. |
| **Zone addressing** | MUST be scoped to a named zone. |
| **Inspectability** | MUST be observable by operators and agents. |
| **Overridability** | MUST be dismissable by an authorized operator. |
| **Composability** | MUST compose with other enforcement primitives and advisory signals. |

---

## 3. Claims

Claims are the primary enforcement primitive. A claim asserts that an actor is working within a zone and requests coordination from other actors.

### 3.1 Claim Envelope

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Globally unique identifier |
| `zone` | string | Yes | Target zone path (same addressing as advisory layer) |
| `mode` | string | Yes | `hard` or `soft` (see §3.2) |
| `state` | string | Yes | Lifecycle state (see §3.4) |
| `actor_id` | string | Yes | Identity of the claiming actor |
| `actor_kind` | string | No | Category: `agent`, `operator`, `remote_agent`, `mcp_client`, `http_client`, `system` |
| `reason` | string | No | Human-readable claim rationale |
| `created_at_unix` | integer | Yes | Claim creation timestamp (Unix seconds) |
| `expires_at_unix` | integer | Conditional | Mandatory for remote-origin claims; RECOMMENDED for all |
| `updated_at_unix` | integer | No | Last modification timestamp |
| `requested_paths` | array of string | No | Specific paths within the zone (advisory precision hint) |
| `negotiation_note` | string | No | Message to conflicting actors |

### 3.2 Claim Modes

| Mode | Behavior |
|------|----------|
| `hard` | Blocks conflicting operations. Another actor attempting to claim the same zone receives a conflict error. |
| `soft` | Warns but does not block. Another actor MAY claim the same zone; both actors receive visibility of the contention. |

A conforming implementation MUST support both modes. The default when mode is ambiguous SHOULD be `soft`.

### 3.3 TTL and Mortality

- Claims with `expires_at_unix` set MUST auto-expire when that time passes
- Remote-origin claims (from HTTP, MCP, or external surfaces) MUST include `expires_at_unix`
- Local claims (from in-process agents) SHOULD include `expires_at_unix`
- The substrate MUST garbage-collect expired claims automatically
- Implementations MAY enforce a maximum TTL (cap overly-long claims)
- Claims without `expires_at_unix` MUST be released either explicitly or on process shutdown

### 3.4 Claim Lifecycle

```
active → released
active → expired
active → overridden
```

| State | Meaning |
|-------|---------|
| `active` | Claim is in effect |
| `released` | Explicitly released by the owning actor |
| `expired` | TTL elapsed, automatically removed |
| `overridden` | Dismissed by an authorized operator |

> **Note:** A `negotiated` transitional state is described in Appendix C for future consideration; conforming implementations are not required to support it.

---

## 4. Actor Identity

### 4.1 Actor Model

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `actor_id` | string | Yes | Stable identifier for the claiming actor |
| `actor_kind` | string | No | Category of actor |

### 4.2 Local vs. Remote Actors

| Origin | Identity Source | TTL Requirement |
|--------|----------------|-----------------|
| **Local** (in-process agent) | Session ID or agent ID | RECOMMENDED |
| **Remote** (HTTP, MCP, external) | Provided in request, validated by auth | REQUIRED |

Remote claims MUST include TTL because the substrate cannot rely on process lifecycle for cleanup. Local claims SHOULD include TTL but MAY rely on process-shutdown cleanup as a fallback.

### 4.3 Identity Semantics

- `actor_id` MUST be stable for the duration of the actor's participation
- The substrate MUST track `actor_id` on all claims
- The substrate MUST support querying claims by `actor_id`
- An actor MAY hold multiple claims simultaneously (in different zones)

---

## 5. Claim Operations

### 5.1 Acquire

Request a claim on a zone.

**Input:**
| Field | Type | Required |
|-------|------|----------|
| `zone` | string | Yes |
| `mode` | string | Yes (`hard` or `soft`) |
| `actor_id` | string | Yes |
| `actor_kind` | string | No |
| `reason` | string | No |
| `ttl_seconds` | integer | Conditional (required for remote) |
| `requested_paths` | array | No |

**Output:**
- **Success:** claim object with assigned `id`, `state: active`, computed `expires_at_unix`
- **Conflict (hard mode):** error identifying the conflicting claim (including its `actor_id`, `zone`, `reason`)
- **Contention (soft mode):** success with contention metadata (existing soft claims in the same zone)

**Validation rules:**
- MUST reject if `zone` is empty
- MUST reject if `actor_id` is empty
- MUST reject remote-origin claims without TTL
- MUST reject if `mode` is not `hard` or `soft`
- For hard claims: MUST reject if an active hard claim exists in the same zone from a different actor
- Implementations MAY enforce per-actor claim limits

### 5.1.1 Capacity Protection

Implementations MAY reject acquisitions when internal capacity limits are reached. When rejecting for capacity, the implementation MUST return error reason `capacity_exceeded`. Implementations SHOULD document their capacity limits.

### 5.2 Release

Explicitly release a claim before TTL expiry.

**Input:** `claim_id`, `actor_id`  
**Output:** success or error

**Rules:**
- An actor MAY only release their own claims
- An operator MAY release any claim (see §6, Override)
- Releasing an already-released or expired claim SHOULD succeed idempotently

### 5.3 Renew

Extend the TTL of an active claim.

**Input:** `claim_id`, `actor_id`, `new_ttl_seconds`  
**Output:** success with updated `expires_at_unix`, or error

**Rules:**
- An actor MAY only renew their own claims
- Renewal resets TTL from current time (not from original creation)
- Renewing an expired claim MUST fail (acquire a new one instead)
- Implementations MAY enforce the same maximum TTL cap on renewals

### 5.4 Override (Operator Dismiss)

Force-release any claim regardless of ownership.

**Input:** `claim_id`, `operator_id`, `reason`  
**Output:** success with claim moved to `overridden` state

**Rules:**
- Override MUST be available for every enforcement primitive (non-negotiable property)
- Override MUST record: who overrode, when, and why
- Override MUST NOT require the original actor's cooperation
- Override SHOULD require elevated privilege (but MUST NOT be impossible)
- After override, the zone is immediately available for new claims

### Atomic Override-and-Reserve (OPTIONAL)

The non-atomic override sequence (override → zone available → new acquire) has a race window where a third party may acquire the zone before the operator's intended actor. Implementations MAY provide an atomic `override_and_reserve` operation that simultaneously overrides the existing claim and acquires a new claim for a specified actor, eliminating this window. When provided, this operation MUST record the same audit trail as a separate override followed by acquire.

---

## 6. Conflict Resolution

### 6.1 Hard Claim Conflicts

When an actor attempts to acquire a hard claim in a zone where another actor holds an active hard claim:

- The substrate MUST reject the acquisition
- The rejection MUST include the conflicting claim's metadata:
  - `conflicting_claim_id`
  - `conflicting_actor_id`
  - `conflicting_reason` (if available)
  - `conflicting_expires_at_unix` (so the requester knows when to retry)

### 6.2 Soft Claim Contention

When an actor acquires a soft claim in a zone where other soft claims exist:

- The substrate MUST allow the acquisition (soft claims never block)
- The substrate SHOULD return contention metadata:
  - List of other active soft claims in the zone
  - Their actor IDs and reasons
- Both actors SHOULD receive visibility of the contention (implementation-defined notification mechanism)

### 6.3 Mixed Mode (Hard + Soft)

- A hard claim blocks new hard claims in the same zone
- A hard claim does NOT block new soft claims
- A soft claim does NOT block new hard claims
- When both exist, the **strictest enforcement wins** for operation gating:
  - If a hard claim is active, operations are blocked for actors that are not the owner of the hard claim
  - A soft claim does not confer ownership status for enforcement gating purposes; "non-owner" means specifically "not the owner of the active hard claim"
  - Soft claims in the same zone provide additional context but don't escalate blocking

### 6.4 Zone Overlap

Claims address specific zones. Zone hierarchy does NOT imply claim hierarchy:

- A claim on `backend/auth` does NOT automatically claim `backend/auth/tokens`
- A claim on `backend` does NOT automatically claim `backend/auth`
- To claim a subtree, acquire claims on each specific zone

---

## 7. Enforcement Gating

### 7.1 Operation Blocking (Hard Claims)

When a hard claim is active in a zone:

- Operations that would modify resources in that zone SHOULD be blocked for non-owning actors
- The blocking MUST be inspectable (the blocked actor can see why they're blocked)
- The blocking MUST be mortal (expires with the claim)
- The blocking MUST be overridable (operator can dismiss the claim)

### 7.2 Warning (Soft Claims)

When a soft claim is active in a zone:

- Operations proceed without blocking
- The substrate SHOULD surface a warning to actors operating in the zone
- The warning MUST include the soft claim's `actor_id` and `reason`

### 7.3 Gating Invariants

- Enforcement MUST NOT be permanent (always mortal)
- Enforcement MUST NOT be unbreakable (always overridable)
- Enforcement state MUST be visible to all participants
- Enforcement MUST identify who created it and why

---

## 8. Readout Surface Requirements

A conforming implementation MUST expose:

### 8.1 List Claims

List active claims, optionally filtered by zone or actor.

**Input:** optional zone glob, optional actor_id filter  
**Output:** array of claim objects

### 8.2 Claim Status

Detailed status for a single claim.

**Input:** claim_id  
**Output:** full claim object including all metadata

### 8.3 Contention Report

Report zones with multiple active claims (contention points).

**Input:** optional zone glob  
**Output:** zones with claim count > 1, with all claims listed per zone

### 8.4 Readout Invariants

- Read operations MUST NOT mutate enforcement state
- Read operations MUST include actor identity for all claims returned
- Read operations MUST indicate claim mode (hard/soft) and remaining TTL

---

## 9. Persistence

### 9.1 Persistence Guarantee

Active claims MUST persist across process restart (if they have `expires_at_unix` set and haven't expired):

- Claims deposited before a crash MUST be enforceable after restart
- Expired claims MUST be garbage-collected on restart
- Claims without `expires_at_unix` (local-only, process-lifetime) MAY be discarded on restart

### 9.2 Crash Recovery

After crash recovery, the substrate MUST:

1. Load persisted claims
2. Garbage-collect claims where `expires_at_unix < now`
3. Restore enforcement state from surviving claims
4. Resume conflict detection

### 9.3 Shutdown Cleanup

On graceful shutdown:

- All claims without `expires_at_unix` SHOULD be released
- Claims with `expires_at_unix` SHOULD persist for potential restart
- The substrate SHOULD log released claims for audit

---

## 10. Multi-Surface Access

### 10.1 Required Surfaces

A conforming implementation MUST provide at least ONE of:

- HTTP REST API
- Agent tool (callable within an agent loop)
- CLI command

### 10.2 Remote Deposit

Claims deposited via remote surfaces (HTTP, MCP) MUST:

- Include authentication/authorization
- Include `expires_at_unix` (mandatory TTL for remote claims)
- Include `actor_id` from the authenticated identity

### 10.3 Surface Parity

All surfaces MUST enforce the same conflict rules. A hard claim acquired via HTTP blocks agents using the agent tool, and vice versa.

---

## 11. Conformance Levels

### Level 1: Basic Claims

The implementation provides claim acquire/release/expire with hard and soft modes.

**Requirements:**
- Claim envelope validation (§3.1)
- Hard and soft modes (§3.2)
- TTL enforcement and garbage collection (§3.3)
- Claim lifecycle state machine (§3.4)
- Conflict detection for hard claims (§6.1)
- Actor identity on all claims (§4)
- Persistence across restart (§9)
- Override capability (§5.4)

### Level 2: Contention-Aware

The implementation provides contention visibility and negotiation.

**Additional requirements beyond Level 1:**
- Soft claim contention reporting (§6.2)
- Contention report readout (§8.3)
- Mixed-mode resolution (§6.3)

### Level 3: Remote-Capable

The implementation accepts claims from remote surfaces with full authentication.

**Additional requirements beyond Level 1:**
- Mandatory TTL validation for remote claims (§4.2)
- Authenticated remote deposit (§10.2)
- Surface parity (§10.3)

---

## 12. Future Enforcement Primitives (Informative)

The following primitives are planned but not yet normative. They will share all properties defined in §2 (mortality, actor identity, zone addressing, inspectability, overridability, composability).

### 12.1 Quarantine Zones

A zone-wide enforcement that prevents all modification until evidence requirements are satisfied. Unlike claims (which assert "I am working here"), quarantine asserts "this zone is unsafe until proven otherwise."

### 12.2 Evidence Gates

A conditional enforcement that blocks a specific operation until specified evidence is produced. The gate defines what evidence is needed; any actor providing that evidence releases the gate.

### 12.3 Capacity Reservations

A resource-bounded enforcement that limits concurrent activity within a zone. Unlike claims (binary: claimed or not), reservations manage a numeric capacity (e.g., "at most 3 concurrent agents in this zone").

### 12.4 Rate Boundaries

A time-windowed enforcement that limits the rate of operations within a zone. Self-healing: automatically releases when the window resets.

---

## 13. Relationship to Advisory Layer

Enforcement and advisory are orthogonal layers:

| Property | Advisory (signals) | Enforcement (claims) |
|----------|-------------------|---------------------|
| Can block operations? | MUST NOT | MAY (hard mode) |
| Response required? | Voluntary | Mandatory (or override) |
| Accumulation semantics? | Yes (pressure) | No (each claim is independent) |
| Induction capable? | Yes | No |
| Operator override? | N/A | Required |

A zone MAY have both active signals AND active claims simultaneously. They do not interfere:

- Signals continue to accumulate regardless of claim state
- Claims enforce regardless of signal state
- Readouts SHOULD combine both for complete zone awareness

---

## Appendix A: Naming Conventions

| Normative term | NOT used in this spec |
|----------------|----------------------|
| claim | lock, lease (except in theory/lineage) |
| enforcement | blocking, gatekeeping |
| zone | trail, namespace |
| actor | holder, owner |
| override | break, force-release |
| mortality | decay, timeout |
| contention | collision, race |

---

## Appendix B: Comparison with Distributed Leases

This spec's claims draw from the distributed lease tradition (Gray & Cheriton 1989) but differ in key ways:

| Property | Distributed Leases | This Spec's Claims |
|----------|-------------------|-------------------|
| Scope | Single resource | Named zone (path-like) |
| Mode | Binary (granted/denied) | Soft + hard |
| Consensus | Required (Paxos, Raft) | Not required (single-host) |
| Transfer | Sometimes supported | Not supported (release + re-acquire) |
| Override | Rarely | Always available |
| Visibility | Holder-only | All participants |

See [docs/theory.md](../docs/theory.md) for the full intellectual lineage.

---

## Appendix C: Negotiated State (Informative — Future Consideration)

A `negotiated` transitional state allows a claim to be modified through contention resolution rather than requiring release and re-acquisition. The lifecycle would extend as:

```
active → negotiated → active (with modified scope)
```

| State | Meaning |
|-------|---------|
| `negotiated` | Claim is being modified through contention resolution between conflicting actors |

This state is not required for conformance. Implementations exploring contention resolution protocols MAY implement this state experimentally, prefixed as `x_negotiated` until a future normative revision incorporates it.
