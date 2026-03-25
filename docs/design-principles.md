# Design Principles

Every primitive in the Agent Coordination Substrate — advisory or enforcement — MUST satisfy all six properties defined here.

---

## The Six Properties

### 1. Mortality

Every signal, claim, and enforcement primitive MUST have a finite time-to-live (TTL).

**Rationale:** In a system of autonomous agents, no single agent can be trusted to release resources. Agents crash, get stuck in loops, lose context, or simply forget. Mortality ensures the substrate self-heals without manual intervention.

**Normative requirements:**
- Every deposit MUST include a TTL field
- The substrate MUST garbage-collect expired entries automatically
- No "infinite" or "permanent" TTL is permitted
- Renewal is an explicit operation that resets TTL from the current time

**Design consequence:** The substrate tends toward emptiness. Left alone, all signals decay and all claims release. This is a feature: stale coordination state is more dangerous than missing coordination state.

---

### 2. Actor Identity

Every deposit and claim MUST identify the actor that created it.

**Rationale:** Attribution is required for debugging, audit, conflict resolution, and operator cognition. When two claims conflict, the operator needs to know which agents are contending. When a signal pattern emerges, the operator needs to know who is contributing.

**Normative requirements:**
- The `actor_id` field MUST be present on all deposits, claims, and modifications
- `actor_id` MUST be a stable identifier for the duration of an agent session
- The substrate MUST NOT accept deposits without actor identity
- Actor identity MUST be included in all readouts and exports

**Design consequence:** Anonymous coordination is not supported. This is intentional: in engineering systems (as opposed to biological systems), attribution enables debugging.

---

### 3. Zone Addressing

All primitives MUST be scoped to a named zone.

**Rationale:** Global broadcast doesn't scale and doesn't allow selective attention. Zones provide locality — an agent working on `backend/auth` doesn't need to observe signals in `frontend/nav`. Zones also provide governance scope: a claim on `internal/db` doesn't block work in `internal/ui`.

**Normative requirements:**
- Every deposit MUST target a named zone
- Zone names MUST be hierarchical (path-like, `/` separated)
- Queries MUST support glob-matching against zone names
- Claims MUST specify their enforcement zone explicitly
- Cross-zone claims MUST enumerate each zone separately (no implicit "everything")

**Design consequence:** There is no "global signal." If you need broad coordination, deposit into a well-known zone (e.g., `workspace/health`). But the default is scoped.

---

### 4. Inspectability

All substrate state MUST be observable by operators and by agents.

**Rationale:** A coordination substrate that cannot be inspected is indistinguishable from a black box. Operators need to understand why agent A is blocked, why signals are accumulating in zone Z, and what the current enforcement state looks like.

**Normative requirements:**
- The substrate MUST provide a read API that returns current state (signals, claims, pressure)
- Readouts MUST include actor identity, zone, remaining TTL, and typed metadata
- The substrate SHOULD provide human-readable summary formats (not just machine JSON)
- Historical state (recently expired signals) SHOULD be available for debugging
- Enforcement state (active claims, blocked operations) MUST be visible in real-time

**Design consequence:** The substrate is transparent by default. No hidden state, no internal-only counters, no "just trust the system" posture.

---

### 5. Overridability

All enforcement primitives MUST be dismissable by an authorized operator.

**Rationale:** No automated system should create unbounded blocking without a human escape hatch. Claims gate operations — but an operator who understands the situation MUST be able to override any claim and proceed.

**Normative requirements:**
- Every enforcement primitive MUST support an `override` or `dismiss` operation
- Override MUST be attributed (who dismissed, when, optional reason)
- Override MUST be auditable (recorded, not silent)
- Override MAY require elevated privilege (but MUST NOT be impossible)
- The substrate MUST NOT prevent override due to its own internal state

**Design consequence:** Enforcement is a strong suggestion with teeth — not a hard wall. The system is designed for humans-in-the-loop, not for unsupervised lockdown.

---

### 6. Composability

Primitives MUST compose — advisory with enforcement, enforcement with enforcement — without requiring coordination between the primitives themselves.

**Rationale:** The substrate must not become a monolith. Different concerns (code quality, capacity, security, workflow state) should each deposit their own signals and claims independently. The substrate aggregates; it doesn't centrally plan.

**Normative requirements:**
- Multiple claims MAY be active in the same zone simultaneously (contention is valid state)
- Multiple signals MAY target the same zone (accumulation is the mechanism)
- Advisory signals MUST NOT interfere with enforcement claims (orthogonal layers)
- Enforcement claims from different actors MAY overlap — resolution is by operator decision, not by priority scheme
- The substrate MUST NOT enforce ordering between deposits from different actors

**Design consequence:** You can run quality signals, capacity claims, and security enforcement in the same workspace zone simultaneously. They don't interfere with each other. They compose.

---

## Advisory vs. Enforcement: Contract Boundary

| Property | Advisory (signals) | Enforcement (claims) |
|----------|-------------------|---------------------|
| **Can block operations?** | MUST NOT | MAY |
| **Agent response required?** | Voluntary | Mandatory (or operator override) |
| **Induction capable?** | Yes (threshold → manifest) | No (claims are explicit) |
| **Multi-signal accumulation?** | Yes (pressure semantics) | No (each claim is independent) |
| **Operator readout?** | Recommended | Required |
| **Override?** | N/A (nothing to override) | Required |

The boundary is absolute: if it blocks, it's enforcement. If it doesn't block, it's advisory. No primitive may be ambiguous about which side of this boundary it lives on.
