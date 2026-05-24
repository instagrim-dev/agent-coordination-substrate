package conformance

import (
	"fmt"
	"os"
	"testing"

	substrate "github.com/instagrim-dev/agent-coordination-substrate/reference/go"
)

func TestAdvisoryConformance(t *testing.T) {
	t.Parallel()

	yamlData, err := os.ReadFile("../../advisory/conformance/expectations.yaml")
	if err != nil {
		t.Fatalf("Failed to read advisory expectations: %v", err)
	}

	clock := NewTestClock(1716000000)
	signals := substrate.NewMemSignalStore(clock)
	claims := substrate.NewMemClaimStore(clock)
	runner := NewRunner(signals, claims, clock)

	results := runner.RunAdvisory(yamlData)

	var passed, failed, skipped int
	for _, r := range results {
		switch r.Status {
		case "pass":
			passed++
		case "fail":
			t.Errorf("FAIL [%s] %s: %s", r.Group, r.ScenarioID, r.Message)
			failed++
		case "skip":
			t.Logf("SKIP [%s] %s: %s", r.Group, r.ScenarioID, r.Message)
			skipped++
		}
	}
	t.Logf("Advisory conformance: %d passed, %d failed, %d skipped", passed, failed, skipped)
}

func TestEnforcementConformance(t *testing.T) {
	t.Parallel()

	yamlData, err := os.ReadFile("../../enforcement/conformance/expectations.yaml")
	if err != nil {
		t.Fatalf("Failed to read enforcement expectations: %v", err)
	}

	clock := NewTestClock(1716000000)
	signals := substrate.NewMemSignalStore(clock)
	claims := substrate.NewMemClaimStore(clock)
	runner := NewRunner(signals, claims, clock)

	results := runner.RunEnforcement(yamlData)

	var passed, failed, skipped int
	for _, r := range results {
		switch r.Status {
		case "pass":
			passed++
		case "fail":
			t.Errorf("FAIL [%s] %s: %s", r.Group, r.ScenarioID, r.Message)
			failed++
		case "skip":
			t.Logf("SKIP [%s] %s: %s", r.Group, r.ScenarioID, r.Message)
			skipped++
		}
	}
	t.Logf("Enforcement conformance: %d passed, %d failed, %d skipped", passed, failed, skipped)
}

// TestBrokenStoreFailsConformance verifies the runner correctly detects
// a non-conforming implementation.
func TestBrokenStoreFailsConformance(t *testing.T) {
	t.Parallel()

	yamlData, err := os.ReadFile("../../enforcement/conformance/expectations.yaml")
	if err != nil {
		t.Fatalf("Failed to read enforcement expectations: %v", err)
	}

	clock := NewTestClock(1716000000)
	runner := NewRunnerWithFactories(
		func(c *TestClock) substrate.SignalStore { return substrate.NewMemSignalStore(c) },
		func(c *TestClock) substrate.ClaimStore { return &brokenClaimStore{} },
		clock,
	)

	results := runner.RunEnforcement(yamlData)

	var failed int
	for _, r := range results {
		if r.Status == "fail" {
			failed++
		}
	}

	if failed == 0 {
		t.Fatal("Expected broken store to produce failures, but all scenarios passed or skipped")
	}
	t.Logf("Broken store correctly failed %d scenarios", failed)
}

// brokenClaimStore never conflicts and never validates — a deliberately
// non-conforming implementation used to test the runner's ability to detect failures.
type brokenClaimStore struct {
	id int
}

func (b *brokenClaimStore) Acquire(zone, mode, actorID, reason string, ttlSeconds int64) (substrate.Claim, error) {
	b.id++
	return substrate.Claim{
		ID:      fmt.Sprintf("broken-%d", b.id),
		Zone:    zone,
		Mode:    mode,
		State:   substrate.StateActive,
		ActorID: actorID,
	}, nil
}

func (b *brokenClaimStore) Release(claimID, actorID string) error {
	return nil
}

func (b *brokenClaimStore) Renew(claimID, actorID string, newTTLSeconds int64) (substrate.Claim, error) {
	return substrate.Claim{}, nil
}

func (b *brokenClaimStore) Override(claimID, operatorID, reason string) error {
	return nil
}

func (b *brokenClaimStore) ListClaims(zoneGlob, actorID string) []substrate.Claim {
	return nil
}
