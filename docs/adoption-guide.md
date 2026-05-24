# Adoption Guide

How to implement the Agent Coordination Substrate in your agent framework.

This guide walks through building a conforming implementation from scratch.
Start with the advisory layer (simpler, no blocking), then add enforcement
when your system needs workspace claims.

---

## Prerequisites

Before implementing, read:
- [Design Principles](design-principles.md) — the six mandatory properties
- [advisory/SPEC.md](../advisory/SPEC.md) — what you're implementing
- [enforcement/SPEC.md](../enforcement/SPEC.md) — what you'll add later

---

## Step 1: Minimum Viable Advisory Implementation

The smallest useful advisory implementation has three components:

### 1.1 Signal Store

A data store that holds signal envelopes. Each signal has:

```
id, zone, kind, actor_id, created_at_unix, expires_at_unix, strength_milli
```

**Requirements:**
- `Deposit(signal)` — validate and store
- `Readout(zone)` — return all live signals in a zone with pressure classification
- Reject deposits with empty zone, empty actor_id, or invalid TTL (expires <= created)
- Generate unique IDs if the caller doesn't provide one
- Default `strength_milli` to 1000 if omitted

**Validation rules (MUST reject):**
| Field | Condition | Error |
|-------|-----------|-------|
| zone | empty | `empty_zone` |
| actor_id | empty | `empty_actor_id` |
| expires_at_unix | <= created_at_unix | `invalid_ttl` |
| kind | empty | `empty_kind` |

### 1.2 Mortality (Garbage Collection)

Signals with `expires_at_unix <= now()` MUST NOT appear in readouts.

Options:
- **Lazy GC**: Check TTL at read time, skip expired signals
- **Background GC**: Periodic sweep removes expired entries
- **Event-driven GC**: TTL triggers fire on expiry

Lazy GC is sufficient for Level 1 conformance.

### 1.3 Zone Addressing

Zones are slash-delimited strings (`backend/auth`, `frontend/nav/sidebar`).

Implement glob matching for zone queries:
- Exact: `backend/auth` matches only `backend/auth`
- Single wildcard: `backend/*` matches `backend/auth` but NOT `backend/auth/tokens`
- Recursive: `backend/**` matches `backend/auth` AND `backend/auth/tokens`

### 1.4 Pressure Classification

Readouts MUST include a pressure state. Minimum viable classification:

| Condition | State |
|-----------|-------|
| No signals | `quiet` |
| 1 actor | `active` |
| 2 actors | `warming` |
| 3+ actors | `convergent` |

The spec allows richer classifications (`saturated`, `contested`, etc.) but the above
satisfies Level 1 conformance.

### 1.5 Verify

Run the advisory conformance expectations (groups: `deposit`, `mortality`,
`zone_addressing`, `pressure_readout`). All scenarios in these groups must pass.

```
Advisory Level 1 = deposit + mortality + zone_addressing + pressure_readout
```

---

## Step 2: Adding Persistence (Level 2)

For crash safety, signals must survive process restart.

**Options:**
- SQLite with WAL mode (recommended — simple, no server)
- Embedded key-value store (BoltDB, BadgerDB)
- External store (Redis, PostgreSQL)

**Requirements:**
- Signals written before crash are readable after restart
- Expired signals are cleaned on restart (lazy GC at read time is sufficient)

Run the `persistence` conformance group to verify.

---

## Step 3: Adding Enforcement (Claims)

The enforcement layer gates operations. Start with claims only.

### 3.1 Claim Store

A data store for claims. Each claim has:

```
id, zone, mode (hard|soft), state, actor_id, created_at_unix, expires_at_unix
```

**Operations:**
- `Acquire(zone, mode, actor_id, reason, ttl_seconds)` — place a claim
- `Release(claim_id, actor_id)` — owner releases their claim
- `Renew(claim_id, actor_id, new_ttl_seconds)` — extend TTL
- `Override(claim_id, operator_id, reason)` — force-release by operator

### 3.2 Conflict Resolution

The core enforcement contract:

| Existing | Incoming | Result |
|----------|----------|--------|
| Hard (Actor A) | Hard (Actor B) | **CONFLICT** — reject incoming |
| Hard (Actor A) | Hard (Actor A) | Accept (idempotent) or reject (implementation choice) |
| Hard (Actor A) | Soft (Actor B) | Accept — soft claims never conflict |
| Soft (any) | Hard (Actor B) | Accept — soft claims don't block |
| Soft (any) | Soft (any) | Accept — soft claims always coexist |

**On conflict, return:**
- The conflicting claim's actor_id
- The conflicting claim's reason (if available)
- The conflicting claim's expires_at_unix (so the caller knows when to retry)

### 3.3 TTL and Expiry

Claims with TTL auto-expire. After expiry:
- The zone is available for new claims
- The expired claim transitions to state `expired`

### 3.4 Override (Operator Escape Hatch)

Any claim can be overridden by an authorized operator:
- Claim transitions to state `overridden`
- Override records: who (operator_id), when (override_at_unix), why (reason)
- Zone becomes immediately available

### 3.5 Owner-Only Release

Only the owning actor can release a claim. Other actors get `actor_mismatch`.
Release is idempotent — releasing an already-released claim succeeds silently.

### 3.6 Verify

Run enforcement conformance expectations (groups: `acquire`, `release`,
`ttl_expiry`, `hard_conflict`, `soft_contention`, `override`).

```
Enforcement Level 1 = acquire + release + ttl_expiry + hard_conflict +
                      soft_contention + override + persistence
```

---

## Step 4: Integration with Agent Protocols

### MCP Integration

Expose the substrate as MCP tools:

```
substrate_deposit     — deposit a signal
substrate_readout     — read zone pressure
substrate_acquire     — place a claim
substrate_release     — release a claim
substrate_override    — operator override
substrate_list_claims — list active claims
```

### A2A Integration

For remote agents coordinating via A2A:
- Remote claims MUST have a TTL (the network can partition)
- Use `actor_kind: "remote_agent"` and `origin: "remote"` for remote claims
- TTL max cap recommended: 5 minutes for remote, configurable for local

### HTTP/REST Integration

Expose as REST endpoints:

```
POST   /v1/signals           — deposit
GET    /v1/zones/{zone}      — readout
POST   /v1/claims            — acquire
DELETE /v1/claims/{id}       — release
POST   /v1/claims/{id}/renew — renew
POST   /v1/claims/{id}/override — operator override
```

---

## Step 5: Operator Tooling

Operators need to see the substrate state at a glance:

### Readout Surface

- **Zone listing** — all active zones with pressure state
- **Claim inventory** — active claims with owner, mode, TTL remaining
- **Signal history** — recent deposits with actor and kind

### Override Surface

- **Dismiss claim** — one-click override with reason prompt
- **Kill signal** — remove a signal before TTL

### Alerting

Optional but recommended:
- Alert on zones in `convergent` or `saturated` state for extended periods
- Alert on claims approaching TTL without renewal
- Alert on override frequency (may indicate systemic issues)

---

## Minimum Viable Implementation Checklist

### Advisory Layer (start here)

- [ ] Signal deposit with validation (zone, actor_id, TTL)
- [ ] Signal ID generation
- [ ] TTL-based mortality (lazy GC minimum)
- [ ] Zone glob matching (exact + wildcard + recursive)
- [ ] Zone pressure readout with state classification
- [ ] Run advisory conformance: `deposit`, `mortality`, `zone_addressing`, `pressure_readout` pass

### Enforcement Layer (add when needed)

- [ ] Claim acquire with mode (hard/soft)
- [ ] Hard-hard conflict detection
- [ ] Owner-only release (idempotent)
- [ ] TTL-based expiry
- [ ] Renew operation
- [ ] Operator override with audit trail
- [ ] Run enforcement conformance: all Level 1 groups pass

### Integration (production-ready)

- [ ] Expose via MCP, REST, or native agent tools
- [ ] Persistence (SQLite recommended)
- [ ] Remote claims require TTL
- [ ] Operator readout surface
- [ ] Override surface with audit logging

---

## Reference Implementation

The [reference implementation](../reference/go/) demonstrates all of the above
in ~500 lines of Go with zero dependencies. Use it as:

1. **A pattern to port** — the logic is simple enough to translate to any language
2. **A conformance target** — run the conformance runner against your implementation
3. **A test oracle** — compare your implementation's behavior against the reference

```bash
cd reference/conformance-runner
go run ./cmd/conformance-runner ../../
```

---

## Conformance Levels

Declare which level your implementation targets:

| Level | Advisory | Enforcement |
|-------|----------|-------------|
| **1 — Signal Store** | deposit, mortality, zone addressing, pressure readout | — |
| **2 — Persistent** | + persistence (crash-safe) | — |
| **3 — Induction-Capable** | + saturation specs, manifest proposals | — |
| **4 — Basic Claims** | Level 1+ | acquire, release, TTL, conflict, override, persistence |
| **5 — Contention-Aware** | Level 1+ | + mixed-mode resolution |
| **6 — Remote-Capable** | Level 1+ | + remote TTL validation |

Run the conformance expectations for your declared level. All scenarios in
the relevant groups must pass (skipping groups above your level is acceptable).

---

## FAQ

**Q: Do I need both layers?**
A: No. The advisory layer is useful on its own. Many systems will never need
enforcement. Add it when agents start overwriting each other's work and signals
aren't enough to prevent it.

**Q: Can I use a different storage backend?**
A: Yes. The spec defines behavior, not storage. Implement the `SignalStore` and
`ClaimStore` interfaces (or their equivalent in your language) backed by whatever
store fits your system.

**Q: How does this relate to MCP/A2A?**
A: It's complementary. MCP and A2A handle direct agent-to-agent communication.
The coordination substrate handles indirect coordination through shared state.
Use both — MCP for delegation, substrate for workspace governance.

**Q: What if I only have one agent?**
A: The substrate still provides value for operator visibility (zone pressure shows
what the agent is focused on) and for future multi-agent readiness. But it's
most valuable with 2+ agents sharing a workspace.

**Q: Is this just a distributed lock manager?**
A: No. Distributed locks are binary (held/free) and typically immortal until
explicitly released. Claims are always mortal (mandatory TTL), mode-aware
(hard vs. soft), and include a governance lifecycle (operator override, audit
trail). The advisory layer has no equivalent in traditional lock managers.
