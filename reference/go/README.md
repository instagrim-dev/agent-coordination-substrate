# Reference Implementation

A minimal, standalone Go package that implements the Agent Coordination Substrate specification.

## Module

```
github.com/instagrim-dev/agent-coordination-substrate/reference/go
```

**Zero external dependencies.** Uses only the Go standard library.  
**No BMO, no Charm, no third-party frameworks.**

## What's Included

| File | Layer | Purpose |
|------|-------|---------|
| `types.go` | Both | Core types: Signal, Claim, ZonePressure, errors |
| `interfaces.go` | Both | `SignalStore` and `ClaimStore` interfaces (target for conformance) |
| `signal_store.go` | Advisory | `MemSignalStore` — in-memory signal store with mortality GC |
| `claim_store.go` | Enforcement | `MemClaimStore` — in-memory claim store with conflict detection |
| `zone.go` | Both | Zone glob matching (`ZoneMatch`) |
| `substrate_test.go` | Both | Unit tests covering core spec behaviors |

## Quick Start

```go
package main

import (
    "fmt"
    substrate "github.com/instagrim-dev/agent-coordination-substrate/reference/go"
)

func main() {
    // Advisory: deposit a signal
    signals := substrate.NewMemSignalStore(nil)
    sig, _ := signals.Deposit(substrate.Signal{
        Zone:          "backend/auth",
        Kind:          "quality_concern",
        ActorID:       "agent-alpha",
        CreatedAtUnix: 1716000000,
        ExpiresAtUnix: 1716003600,
        StrengthMilli: 800,
    })
    fmt.Println("Deposited:", sig.ID)

    // Read zone pressure
    pressure := signals.Readout("backend/auth")
    fmt.Printf("Zone %s: %d deposits, state=%s\n",
        pressure.Zone, pressure.DepositCount, pressure.Pressure.State)

    // Enforcement: acquire a claim
    claims := substrate.NewMemClaimStore(nil)
    claim, _ := claims.Acquire("backend/auth", substrate.ModeHard, "agent-alpha", "refactoring", 3600)
    fmt.Println("Claimed:", claim.ID, claim.State)

    // Another agent tries to claim the same zone
    _, err := claims.Acquire("backend/auth", substrate.ModeHard, "agent-beta", "also working", 3600)
    if err != nil {
        fmt.Println("Conflict:", err)
    }
}
```

## Running Tests

```bash
go test ./... -v
```

## Design Decisions

1. **No persistence** — The memstore is intentionally volatile. Production deployments
   should implement `SignalStore` and `ClaimStore` backed by a durable store (SQLite, etc.).

2. **Injectable clock** — All time operations use the `Clock` interface, enabling
   deterministic testing and time-travel scenarios in conformance tests.

3. **No external dependencies** — The reference implementation proves the spec is
   implementable with nothing beyond Go's standard library.

4. **Interface-first** — `SignalStore` and `ClaimStore` are the conformance targets.
   Any implementation satisfying these interfaces can be validated by the conformance runner.
