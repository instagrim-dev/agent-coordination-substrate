# Conformance Runner

Validates any implementation of the Agent Coordination Substrate against the
specification's conformance expectations.

## How It Works

The runner reads `expectations.yaml` files (one per layer) and executes each
scenario as a sequence of operations against your store implementation.

### Architecture

```
expectations.yaml → Runner → SignalStore / ClaimStore interface → Results
```

The runner:
1. Parses YAML expectations (advisory + enforcement layers)
2. Resets store state per scenario for isolation
3. Executes each step: deposit, readout, acquire, release, override, etc.
4. Compares actual results against expected outcomes
5. Reports pass/fail/skip per scenario

### Isolation Model

- Stores are reset per scenario (each scenario gets a clean slate)
- Exception: read-only scenarios (no deposits/acquires) carry forward group state
- Clock resets to epoch at each scenario start
- `advance_time` steps move the clock forward

## Usage

### As a library (in tests)

```go
import (
    conformance "github.com/instagrim-dev/agent-coordination-substrate/reference/conformance-runner"
    substrate "github.com/instagrim-dev/agent-coordination-substrate/reference/go"
)

func TestMyStore(t *testing.T) {
    clock := conformance.NewTestClock(1716000000)
    runner := conformance.NewRunner(
        mySignalStore,
        myClaimStore,
        clock,
    )

    advisory, _ := os.ReadFile("advisory/conformance/expectations.yaml")
    for _, r := range runner.RunAdvisory(advisory) {
        if r.Status == "fail" {
            t.Errorf("[%s] %s: %s", r.Group, r.ScenarioID, r.Message)
        }
    }
}
```

### As a CLI

```bash
cd reference/conformance-runner
go run ./cmd/conformance-runner ../../
```

Output:
```
=== Advisory Layer ===
  PASS  [deposit] deposit_valid_signal
  PASS  [deposit] deposit_rejects_missing_zone
  ...

=== Enforcement Layer ===
  PASS  [acquire] acquire_hard_claim
  PASS  [acquire] acquire_soft_claim
  ...

Summary: 42 passed, 0 failed, 9 skipped
```

## Custom Store Implementations

To validate your own store, implement the `SignalStore` and `ClaimStore`
interfaces from the reference package:

```go
type SignalStore interface {
    Deposit(sig Signal) (Signal, error)
    Readout(zone string) ZonePressure
    ListZones(glob string) []ZonePressure
    Reinforce(signalID, actorID string, newExpiresAtUnix int64) error
    Kill(signalID, actorID string) error
}

type ClaimStore interface {
    Acquire(zone, mode, actorID, reason string, ttlSeconds int64) (Claim, error)
    Release(claimID, actorID string) error
    Renew(claimID, actorID string, newTTLSeconds int64) (Claim, error)
    Override(claimID, operatorID, reason string) error
    ListClaims(zoneGlob, actorID string) []Claim
}
```

Pass your implementation to `NewRunner` and run both layers.

## Skipped Scenarios

Scenarios are skipped when the store doesn't support the operation:
- **`restart`** — requires persistence (in-memory stores skip these)
- **`induction`** — requires saturation spec configuration (Level 4)

Skipped scenarios don't count as failures. They indicate the implementation
doesn't claim that conformance level.
