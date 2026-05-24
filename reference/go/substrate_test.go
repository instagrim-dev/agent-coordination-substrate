package substrate

import (
	"testing"
	"time"
)

// testClock allows advancing time in tests.
type testClock struct {
	now time.Time
}

func (c *testClock) Now() time.Time { return c.now }
func (c *testClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func newTestClock() *testClock {
	return &testClock{now: time.Unix(1716000000, 0)}
}

func TestSignalDeposit(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemSignalStore(clock)

	sig, err := store.Deposit(Signal{
		ID:            "sig-001",
		Zone:          "backend/auth",
		Kind:          "quality_concern",
		ActorID:       "agent-alpha",
		CreatedAtUnix: 1716000000,
		ExpiresAtUnix: 1716003600,
		StrengthMilli: 800,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sig.ID != "sig-001" {
		t.Fatalf("expected id sig-001, got %s", sig.ID)
	}

	readout := store.Readout("backend/auth")
	if readout.DepositCount != 1 {
		t.Fatalf("expected 1 deposit, got %d", readout.DepositCount)
	}
}

func TestSignalRejectsEmptyZone(t *testing.T) {
	t.Parallel()
	store := NewMemSignalStore(nil)
	_, err := store.Deposit(Signal{
		ID:            "sig-002",
		Zone:          "",
		Kind:          "test",
		ActorID:       "agent-alpha",
		CreatedAtUnix: 1716000000,
		ExpiresAtUnix: 1716003600,
	})
	if err == nil {
		t.Fatal("expected error for empty zone")
	}
}

func TestSignalRejectsInvalidTTL(t *testing.T) {
	t.Parallel()
	store := NewMemSignalStore(nil)
	_, err := store.Deposit(Signal{
		ID:            "sig-003",
		Zone:          "test",
		Kind:          "test",
		ActorID:       "agent-alpha",
		CreatedAtUnix: 1716003600,
		ExpiresAtUnix: 1716000000,
	})
	if err == nil {
		t.Fatal("expected error for invalid TTL")
	}
}

func TestSignalMortality(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemSignalStore(clock)

	_, err := store.Deposit(Signal{
		ID:            "sig-010",
		Zone:          "test/mortality",
		Kind:          "test",
		ActorID:       "agent-alpha",
		CreatedAtUnix: 1716000000,
		ExpiresAtUnix: 1716000060,
	})
	if err != nil {
		t.Fatal(err)
	}

	readout := store.Readout("test/mortality")
	if readout.DepositCount != 1 {
		t.Fatalf("expected 1, got %d", readout.DepositCount)
	}

	clock.Advance(61 * time.Second)
	readout = store.Readout("test/mortality")
	if readout.DepositCount != 0 {
		t.Fatalf("expected 0 after TTL, got %d", readout.DepositCount)
	}
}

func TestSignalReinforce(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemSignalStore(clock)

	_, _ = store.Deposit(Signal{
		ID:            "sig-011",
		Zone:          "test/reinforce",
		Kind:          "test",
		ActorID:       "agent-alpha",
		CreatedAtUnix: 1716000000,
		ExpiresAtUnix: 1716000060,
	})

	err := store.Reinforce("sig-011", "agent-alpha", 1716000120)
	if err != nil {
		t.Fatal(err)
	}

	clock.Advance(90 * time.Second)
	readout := store.Readout("test/reinforce")
	if readout.DepositCount != 1 {
		t.Fatalf("expected 1 after reinforce, got %d", readout.DepositCount)
	}
}

func TestSignalReinforceRejectsDifferentActor(t *testing.T) {
	t.Parallel()
	store := NewMemSignalStore(nil)
	_, _ = store.Deposit(Signal{
		ID:            "sig-012",
		Zone:          "test",
		Kind:          "test",
		ActorID:       "agent-alpha",
		CreatedAtUnix: 1716000000,
		ExpiresAtUnix: 1716003600,
	})

	err := store.Reinforce("sig-012", "agent-beta", 1716007200)
	if err == nil {
		t.Fatal("expected actor_mismatch error")
	}
}

func TestClaimAcquireAndRelease(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemClaimStore(clock)

	claim, err := store.Acquire("backend/auth", ModeHard, "agent-alpha", "refactoring", 3600)
	if err != nil {
		t.Fatal(err)
	}
	if claim.State != StateActive {
		t.Fatalf("expected active, got %s", claim.State)
	}

	claims := store.ListClaims("backend/auth", "")
	if len(claims) != 1 {
		t.Fatalf("expected 1 claim, got %d", len(claims))
	}

	err = store.Release(claim.ID, "agent-alpha")
	if err != nil {
		t.Fatal(err)
	}

	claims = store.ListClaims("backend/auth", "")
	if len(claims) != 0 {
		t.Fatalf("expected 0 claims after release, got %d", len(claims))
	}
}

func TestClaimHardConflict(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemClaimStore(clock)

	_, err := store.Acquire("test/conflict", ModeHard, "agent-alpha", "working", 3600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Acquire("test/conflict", ModeHard, "agent-beta", "also working", 3600)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	ce, ok := err.(*ConflictError)
	if !ok {
		t.Fatalf("expected ConflictError, got %T", err)
	}
	if ce.ConflictingActorID != "agent-alpha" {
		t.Fatalf("expected conflicting actor agent-alpha, got %s", ce.ConflictingActorID)
	}
}

func TestClaimSoftCoexistence(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemClaimStore(clock)

	_, err := store.Acquire("test/soft", ModeSoft, "agent-alpha", "reading", 3600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Acquire("test/soft", ModeSoft, "agent-beta", "also reading", 3600)
	if err != nil {
		t.Fatal(err)
	}

	claims := store.ListClaims("test/soft", "")
	if len(claims) != 2 {
		t.Fatalf("expected 2 soft claims, got %d", len(claims))
	}
}

func TestClaimTTLExpiry(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemClaimStore(clock)

	_, err := store.Acquire("test/ttl", ModeHard, "agent-alpha", "working", 60)
	if err != nil {
		t.Fatal(err)
	}

	clock.Advance(61 * time.Second)

	claims := store.ListClaims("test/ttl", "")
	if len(claims) != 0 {
		t.Fatalf("expected 0 after TTL, got %d", len(claims))
	}

	// Zone should now be available
	_, err = store.Acquire("test/ttl", ModeHard, "agent-beta", "new work", 3600)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClaimOverride(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemClaimStore(clock)

	claim, err := store.Acquire("test/override", ModeHard, "agent-alpha", "working", 3600)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Override(claim.ID, "operator-1", "emergency hotfix")
	if err != nil {
		t.Fatal(err)
	}

	claims := store.ListClaims("test/override", "")
	if len(claims) != 0 {
		t.Fatalf("expected 0 after override, got %d", len(claims))
	}

	// Zone should be available for new claims
	_, err = store.Acquire("test/override", ModeHard, "agent-beta", "hotfix", 3600)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClaimReleaseRejectsNonOwner(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	store := NewMemClaimStore(clock)

	claim, err := store.Acquire("test/owner", ModeHard, "agent-alpha", "working", 3600)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Release(claim.ID, "agent-beta")
	if err == nil {
		t.Fatal("expected actor_mismatch error")
	}
}

func TestZoneMatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern string
		zone    string
		want    bool
	}{
		{"backend/auth", "backend/auth", true},
		{"backend/auth", "backend/api", false},
		{"backend/*", "backend/auth", true},
		{"backend/*", "backend/api", true},
		{"backend/*", "backend/auth/tokens", false},
		{"backend/**", "backend/auth", true},
		{"backend/**", "backend/auth/tokens", true},
		{"backend/**", "backend", true},
		{"backend/**", "frontend/nav", false},
		{"**", "anything/at/all", true},
	}
	for _, tt := range tests {
		got := ZoneMatch(tt.pattern, tt.zone)
		if got != tt.want {
			t.Errorf("ZoneMatch(%q, %q) = %v, want %v", tt.pattern, tt.zone, got, tt.want)
		}
	}
}
