# Transport Binding: HTTP

**Version:** 0.1.0  
**Status:** Normative  
**Scope:** HTTP transport for all three substrate layers  
**Depends on:** Advisory SPEC (Layer 1), Enforcement SPEC (Layer 2), Induction SPEC (Layer 3)

This document defines the HTTP transport binding for the Agent Coordination Substrate. It maps the behavioral contracts defined in the advisory, enforcement, and induction specifications to concrete HTTP endpoints, request/response formats, authentication, error codes, and event streams.

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [RFC 2119](https://www.ietf.org/rfc/rfc2119.txt).

---

## 1. Protocol Overview

The transport binding exposes the substrate as an HTTP/1.1 (or HTTP/2) JSON API with optional Server-Sent Events for real-time observation.

| Component | Method | Description |
|-----------|--------|-------------|
| Signal deposit | `POST /signals` | Create an advisory signal |
| Signal reinforce | `PATCH /signals/{id}/ttl` | Extend signal TTL |
| Signal kill | `DELETE /signals/{id}` | Immediately expire a signal |
| Zone readout | `GET /zones/{zone}` | Read pressure state |
| Zone list | `GET /zones` | List zones matching a glob |
| Claim acquire | `POST /claims` | Request an enforcement claim |
| Claim release | `DELETE /claims/{id}` | Release a claim |
| Claim renew | `PATCH /claims/{id}/ttl` | Extend claim TTL |
| Claim override | `POST /claims/{id}/override` | Operator force-release |
| Claim list | `GET /claims` | List active claims |
| Saturation configure | `POST /induction/specs` | Register a saturation spec |
| Proposal list | `GET /induction/proposals` | List manifest proposals |
| Proposal accept | `POST /induction/proposals/{id}/accept` | Promote a proposal |
| Proposal dismiss | `POST /induction/proposals/{id}/dismiss` | Reject a proposal |
| Event stream | `GET /events` | SSE stream of substrate events |

---

## 2. Common Conventions

### 2.1 Content Type

All request and response bodies MUST use `application/json` with UTF-8 encoding.

```
Content-Type: application/json; charset=utf-8
```

### 2.2 Authentication

Requests MUST include a Bearer token in the `Authorization` header:

```
Authorization: Bearer <token>
```

The token MUST resolve to an actor identity. The substrate MUST reject requests with missing or invalid tokens. Token format and validation are implementation-defined.

Operator-level operations (override, force-release) REQUIRE tokens with elevated privilege. The privilege model is implementation-defined, but MUST distinguish between agent-level and operator-level authorization.

### 2.3 Actor Identity Resolution

The substrate MUST derive `actor_id` and `actor_kind` from the authenticated token. Clients MUST NOT self-assert `actor_id` in request bodies — the server populates it from the authentication context.

Exception: the `actor_kind` field MAY be included in requests as an advisory hint when the authentication system cannot distinguish actor categories.

### 2.4 Time

All timestamps are Unix seconds (integer). The substrate MUST use its own clock for `created_at_unix`. Clients provide `ttl_seconds`; the server computes `expires_at_unix = now + ttl_seconds`.

### 2.5 Zone Encoding

Zone paths appear in URL path segments. The `/` separator in zone names maps to the URL path separator. Clients MUST percent-encode reserved characters per RFC 3986.

| Zone | URL path |
|------|----------|
| `backend/auth` | `/zones/backend/auth` |
| `backend/auth/tokens` | `/zones/backend/auth/tokens` |

For glob queries, zones are passed as query parameters (§4.2).

---

## 3. Advisory Layer Endpoints

### 3.1 Deposit Signal

```
POST /signals
```

**Request body:**

```json
{
  "zone": "backend/auth",
  "kind": "quality_concern",
  "ttl_seconds": 3600,
  "strength_milli": 800,
  "reason": "Auth module has no rate limiting",
  "properties": {}
}
```

**Required fields:** `zone`, `kind`, `ttl_seconds`  
**Optional fields:** `strength_milli` (default 1000), `reason`, `properties`, `session_id`, `workflow_id`

**Response (201 Created):**

```json
{
  "id": "sig-a7f3e2",
  "zone": "backend/auth",
  "kind": "quality_concern",
  "actor_id": "agent-alpha",
  "created_at_unix": 1716000000,
  "expires_at_unix": 1716003600,
  "strength_milli": 800
}
```

**Errors:**

| Status | Reason | Condition |
|--------|--------|-----------|
| 400 | `empty_zone` | Zone is empty string |
| 400 | `empty_kind` | Kind is empty string |
| 400 | `invalid_ttl` | TTL is zero or negative |
| 401 | `unauthorized` | Missing or invalid token |
| 429 | `capacity_exceeded` | Store at capacity |

### 3.2 Reinforce Signal

```
PATCH /signals/{id}/ttl
```

**Request body:**

```json
{
  "additional_ttl_seconds": 3600
}
```

**Response (200 OK):**

```json
{
  "id": "sig-a7f3e2",
  "expires_at_unix": 1716007200
}
```

**Errors:**

| Status | Reason | Condition |
|--------|--------|-----------|
| 403 | `actor_mismatch` | Caller is not the signal's actor |
| 404 | `not_found` | Signal does not exist or is expired |

### 3.3 Kill Signal

```
DELETE /signals/{id}
```

**Response (204 No Content)** on success.

**Errors:**

| Status | Reason | Condition |
|--------|--------|-----------|
| 403 | `actor_mismatch` | Caller is not the signal's actor and lacks operator privilege |
| 404 | `not_found` | Signal does not exist |

Operators MAY kill any signal (elevated privilege).

---

## 4. Zone Readout Endpoints

### 4.1 Zone Pressure

```
GET /zones/{zone}
```

**Response (200 OK):**

```json
{
  "zone": "backend/auth",
  "pressure": {
    "state": "convergent",
    "score": 0.72
  },
  "trend": {
    "state": "surging",
    "current_deposit_count": 4,
    "recent_deposit_count": 2,
    "short_window_seconds": 300,
    "medium_window_seconds": 3600,
    "newest_signal_created_at_unix": 1716003500,
    "oldest_signal_created_at_unix": 1716000000
  },
  "deposit_count": 6,
  "independent_lineages": 3,
  "contributing_signals": [...],
  "evidence_truncated": false
}
```

**Response (200 OK) for empty zone:**

```json
{
  "zone": "empty/zone",
  "pressure": { "state": "quiet", "score": 0.0 },
  "trend": { "state": "stale" },
  "deposit_count": 0,
  "independent_lineages": 0,
  "contributing_signals": [],
  "evidence_truncated": false
}
```

A conforming server MUST return a valid readout for any zone path, including zones with no signals.

### 4.2 Zone List

```
GET /zones?glob={pattern}&limit={n}
```

**Query parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `glob` | string | Zone glob pattern (optional; omit for all zones) |
| `limit` | integer | Maximum zones to return (optional) |
| `offset` | integer | Pagination offset (optional) |

**Response (200 OK):**

```json
{
  "zones": [
    { "zone": "backend/auth", "deposit_count": 3, "pressure_state": "convergent" },
    { "zone": "backend/api", "deposit_count": 1, "pressure_state": "active" }
  ],
  "total": 2,
  "truncated": false
}
```

---

## 5. Enforcement Layer Endpoints

### 5.1 Acquire Claim

```
POST /claims
```

**Request body:**

```json
{
  "zone": "backend/auth",
  "mode": "hard",
  "ttl_seconds": 600,
  "reason": "Refactoring auth middleware",
  "requested_paths": ["internal/auth/middleware.go"]
}
```

**Required fields:** `zone`, `mode`, `ttl_seconds`  
**Optional fields:** `reason`, `requested_paths`, `negotiation_note`

**Response (201 Created):**

```json
{
  "id": "claim-b8e4f1",
  "zone": "backend/auth",
  "mode": "hard",
  "state": "active",
  "actor_id": "agent-alpha",
  "actor_kind": "agent",
  "created_at_unix": 1716000000,
  "expires_at_unix": 1716000600,
  "reason": "Refactoring auth middleware",
  "requested_paths": ["internal/auth/middleware.go"]
}
```

**Response (201 Created, soft claim with contention):**

```json
{
  "id": "claim-c9f5g2",
  "zone": "backend/auth",
  "mode": "soft",
  "state": "active",
  "actor_id": "agent-beta",
  "created_at_unix": 1716000100,
  "expires_at_unix": 1716000700,
  "contention": [
    {
      "claim_id": "claim-b8e4f1",
      "actor_id": "agent-alpha",
      "mode": "hard",
      "reason": "Refactoring auth middleware"
    }
  ]
}
```

**Response (409 Conflict, hard claim blocked):**

```json
{
  "error": "conflict",
  "message": "Zone is held by an active hard claim",
  "conflicting_claim": {
    "id": "claim-b8e4f1",
    "actor_id": "agent-alpha",
    "zone": "backend/auth",
    "reason": "Refactoring auth middleware",
    "expires_at_unix": 1716000600
  }
}
```

**Errors:**

| Status | Reason | Condition |
|--------|--------|-----------|
| 400 | `empty_zone` | Zone is empty |
| 400 | `invalid_mode` | Mode is not `hard` or `soft` |
| 400 | `invalid_ttl` | TTL is zero or negative |
| 401 | `unauthorized` | Missing or invalid token |
| 409 | `conflict` | Hard claim held by another actor |
| 429 | `capacity_exceeded` | Store at capacity |

### 5.2 Release Claim

```
DELETE /claims/{id}
```

**Response (204 No Content)** on success.

**Errors:**

| Status | Reason | Condition |
|--------|--------|-----------|
| 403 | `actor_mismatch` | Caller is not the claim's actor and lacks operator privilege |
| 404 | `not_found` | Claim does not exist |

Releasing an already-released or expired claim SHOULD return 204 (idempotent).

### 5.3 Renew Claim

```
PATCH /claims/{id}/ttl
```

**Request body:**

```json
{
  "ttl_seconds": 600
}
```

**Response (200 OK):**

```json
{
  "id": "claim-b8e4f1",
  "expires_at_unix": 1716001200
}
```

`expires_at_unix` is recomputed as `now + ttl_seconds` (renewal from current time, not original creation).

**Errors:**

| Status | Reason | Condition |
|--------|--------|-----------|
| 403 | `actor_mismatch` | Caller is not the claim's actor |
| 404 | `not_found` | Claim does not exist or is expired |
| 410 | `expired` | Claim has already expired (acquire a new one) |

### 5.4 Override Claim

```
POST /claims/{id}/override
```

Requires operator-level privilege.

**Request body:**

```json
{
  "reason": "Unblocking deployment pipeline"
}
```

**Response (200 OK):**

```json
{
  "id": "claim-b8e4f1",
  "state": "overridden",
  "overridden_by": "operator-admin",
  "overridden_at_unix": 1716000300,
  "override_reason": "Unblocking deployment pipeline"
}
```

**Errors:**

| Status | Reason | Condition |
|--------|--------|-----------|
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `insufficient_privilege` | Token lacks operator-level access |
| 404 | `not_found` | Claim does not exist |

### 5.5 Override and Reserve (OPTIONAL)

```
POST /claims/{id}/override-and-reserve
```

Atomic override + acquire. Eliminates the race window between override and subsequent acquire.

**Request body:**

```json
{
  "override_reason": "Unblocking deployment pipeline",
  "new_claim": {
    "mode": "hard",
    "ttl_seconds": 300,
    "actor_id_hint": "agent-deployer",
    "reason": "Deployment in progress"
  }
}
```

**Response (200 OK):**

```json
{
  "overridden_claim": {
    "id": "claim-b8e4f1",
    "state": "overridden",
    "overridden_by": "operator-admin"
  },
  "new_claim": {
    "id": "claim-d1a6h3",
    "zone": "backend/auth",
    "mode": "hard",
    "state": "active",
    "actor_id": "agent-deployer",
    "expires_at_unix": 1716000600
  }
}
```

### 5.6 List Claims

```
GET /claims?zone={glob}&actor_id={id}&state={state}
```

**Query parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `zone` | string | Filter by zone glob (optional) |
| `actor_id` | string | Filter by actor (optional) |
| `state` | string | Filter by state: `active`, `released`, `expired`, `overridden` (optional) |
| `limit` | integer | Maximum results (optional) |
| `offset` | integer | Pagination offset (optional) |

**Response (200 OK):**

```json
{
  "claims": [...],
  "total": 3,
  "truncated": false
}
```

---

## 6. Induction Layer Endpoints

### 6.1 Configure Saturation Spec

```
POST /induction/specs
```

**Request body:**

```json
{
  "zone_glob": "backend/**",
  "threshold_deposits": 5,
  "threshold_lineages": 3,
  "window_seconds": 7200,
  "kind_filter": "quality_concern",
  "manifest_ttl_seconds": 86400
}
```

**Response (201 Created):**

```json
{
  "id": "spec-e2f7a1",
  "zone_glob": "backend/**",
  "threshold_deposits": 5,
  "threshold_lineages": 3,
  "window_seconds": 7200,
  "enabled": true
}
```

### 6.2 List Proposals

```
GET /induction/proposals?zone={glob}&state={state}
```

**Response (200 OK):**

```json
{
  "proposals": [
    {
      "manifest_id": "manifest-f3g8b2",
      "zone": "backend/auth",
      "state": "pending",
      "confidence": 0.85,
      "actor_count": 3,
      "signal_ids": ["sig-a1", "sig-b2", "sig-c3"],
      "proposed_at_unix": 1716003600,
      "expires_at_unix": 1716090000
    }
  ],
  "total": 1,
  "truncated": false
}
```

### 6.3 Accept Proposal

```
POST /induction/proposals/{id}/accept
```

Requires operator-level or configured privilege.

**Response (200 OK):**

```json
{
  "manifest_id": "manifest-f3g8b2",
  "state": "accepted",
  "accepted_by": "operator-admin",
  "accepted_at_unix": 1716004000
}
```

### 6.4 Dismiss Proposal

```
POST /induction/proposals/{id}/dismiss
```

**Request body:**

```json
{
  "reason": "False positive — unrelated signals"
}
```

**Response (200 OK):**

```json
{
  "manifest_id": "manifest-f3g8b2",
  "state": "dismissed",
  "dismissed_by": "operator-admin",
  "dismissed_at_unix": 1716004000,
  "dismiss_reason": "False positive — unrelated signals"
}
```

---

## 7. Event Stream

```
GET /events
```

Server-Sent Events stream for real-time substrate observation. The client opens an SSE connection; the server pushes events as substrate state changes.

### 7.1 Event Format

Each SSE event has a `type` field and a JSON `data` payload:

```
event: signal.deposited
data: {"id":"sig-a7f3e2","zone":"backend/auth","kind":"quality_concern","actor_id":"agent-alpha"}

event: claim.acquired
data: {"id":"claim-b8e4f1","zone":"backend/auth","mode":"hard","actor_id":"agent-alpha","expires_at_unix":1716000600}

event: claim.released
data: {"id":"claim-b8e4f1","zone":"backend/auth","actor_id":"agent-alpha"}

event: claim.expired
data: {"id":"claim-b8e4f1","zone":"backend/auth"}

event: claim.overridden
data: {"id":"claim-b8e4f1","zone":"backend/auth","overridden_by":"operator-admin"}

event: proposal.generated
data: {"manifest_id":"manifest-f3g8b2","zone":"backend/auth","confidence":0.85}

event: proposal.accepted
data: {"manifest_id":"manifest-f3g8b2","accepted_by":"operator-admin"}

event: proposal.dismissed
data: {"manifest_id":"manifest-f3g8b2","dismissed_by":"operator-admin"}
```

### 7.2 Event Types

| Event | Trigger |
|-------|---------|
| `signal.deposited` | New signal accepted |
| `signal.reinforced` | Signal TTL extended |
| `signal.killed` | Signal explicitly killed |
| `signal.expired` | Signal TTL elapsed |
| `claim.acquired` | New claim created |
| `claim.released` | Claim explicitly released |
| `claim.expired` | Claim TTL elapsed |
| `claim.overridden` | Claim dismissed by operator |
| `claim.contention` | Soft claim overlaps existing claims |
| `proposal.generated` | Induction produced a manifest |
| `proposal.accepted` | Proposal promoted |
| `proposal.dismissed` | Proposal rejected |

### 7.3 Filtering

Clients MAY filter the event stream using query parameters:

```
GET /events?zone=backend/**&types=claim.acquired,claim.released
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `zone` | string | Zone glob filter (only events for matching zones) |
| `types` | string | Comma-separated event types |
| `actor_id` | string | Only events involving this actor |

### 7.4 Reconnection

The server MUST include an `id` field with each event (monotonically increasing). Clients reconnect with `Last-Event-ID` header to resume from where they left off. The server SHOULD buffer recent events for reconnection (buffer size implementation-defined, RECOMMENDED minimum 1000 events or 5 minutes).

---

## 8. Error Response Format

All error responses use a consistent JSON envelope:

```json
{
  "error": "conflict",
  "message": "Zone is held by an active hard claim",
  "details": {}
}
```

| Field | Type | Description |
|-------|------|-------------|
| `error` | string | Machine-readable error code |
| `message` | string | Human-readable description |
| `details` | object | Additional context (error-specific) |

### 8.1 Standard Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `empty_zone` | 400 | Zone path is empty |
| `empty_kind` | 400 | Signal kind is empty |
| `invalid_ttl` | 400 | TTL is zero or negative |
| `invalid_mode` | 400 | Mode is not `hard` or `soft` |
| `unauthorized` | 401 | Missing or invalid authentication |
| `insufficient_privilege` | 403 | Token lacks required privilege level |
| `actor_mismatch` | 403 | Caller does not own this resource |
| `not_found` | 404 | Resource does not exist |
| `conflict` | 409 | Hard claim conflict |
| `expired` | 410 | Resource has expired |
| `capacity_exceeded` | 429 | Store at capacity |

---

## 9. Security Considerations

### 9.1 Authentication

Every request MUST be authenticated. The transport binding does not permit anonymous access. Implementations MUST validate tokens before processing any operation.

### 9.2 Actor Isolation

The substrate MUST enforce actor isolation: an actor cannot release, reinforce, or renew another actor's resources without operator privilege. The server derives actor identity from the token — clients cannot forge identity.

### 9.3 TTL Manipulation

Clients specify TTL; the server computes expiry from its own clock. Clients cannot set arbitrary `expires_at_unix` directly. Implementations SHOULD enforce maximum TTL caps to prevent denial-of-service through unreasonably long claims.

### 9.4 Zone Injection

Zone paths from URL segments are validated before use. Implementations MUST reject path traversal attempts (`../`) and MUST normalize zone paths (strip leading/trailing slashes, collapse repeated separators).

### 9.5 Rate Limiting

Implementations SHOULD rate-limit requests per actor. Rate limits SHOULD be documented and return `429 Too Many Requests` with a `Retry-After` header.

### 9.6 Information Disclosure

Zone readout is available to all authenticated actors. Implementations that require zone-level access control MUST define an authorization model beyond the scope of this specification.

### 9.7 TLS

All transport MUST be encrypted. Implementations MUST support TLS 1.2 or later. Plaintext HTTP MUST NOT be used in production deployments.

---

## 10. IANA Considerations

### 10.1 Media Type

This specification uses `application/json` (RFC 8259). No new media type registration is required.

### 10.2 Well-Known URI (OPTIONAL)

Implementations MAY register a well-known URI for substrate discovery:

```
/.well-known/agent-coordination-substrate
```

Response:

```json
{
  "version": "0.1.0",
  "endpoints": {
    "signals": "/signals",
    "zones": "/zones",
    "claims": "/claims",
    "induction": "/induction",
    "events": "/events"
  },
  "capabilities": ["advisory", "enforcement", "induction", "events"]
}
```

---

## 11. Conformance

A conforming HTTP transport implementation MUST:

1. Implement all advisory endpoints (§3, §4)
2. Implement all enforcement endpoints (§5)
3. Return correct HTTP status codes per this specification
4. Authenticate all requests (§9.1)
5. Derive actor identity from authentication context (§2.3)
6. Enforce TTL from server clock (§2.4)
7. Return errors in the standard format (§8)

A conforming implementation SHOULD additionally:

- Implement the event stream (§7)
- Implement induction endpoints (§6)
- Support the `Last-Event-ID` reconnection protocol (§7.4)
- Implement rate limiting (§9.5)

---

## 12. Relationship to Behavioral Specifications

This document defines HOW substrate operations are invoked over HTTP. The WHAT — behavioral semantics, invariants, and lifecycle rules — is defined in:

- [advisory/SPEC.md](../advisory/SPEC.md) — Signal deposit, mortality, pressure classification
- [enforcement/SPEC.md](../enforcement/SPEC.md) — Claim acquisition, conflict resolution, override
- [induction/SPEC.md](../induction/SPEC.md) — Saturation detection, manifest generation, proposal lifecycle

The transport binding MUST NOT contradict the behavioral specifications. Where this document specifies HTTP-specific behavior (status codes, headers, URL structure), it supplements — not replaces — the normative behavioral contract.
