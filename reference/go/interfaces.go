package substrate

// SignalStore is the interface for advisory layer implementations.
// Any implementation that satisfies this interface can be validated
// by the conformance runner.
type SignalStore interface {
	// Deposit creates a signal in the store.
	Deposit(sig Signal) (Signal, error)

	// Readout returns zone pressure for the given zone.
	Readout(zone string) ZonePressure

	// ListZones returns zone pressure readouts for all zones matching the glob.
	ListZones(glob string) []ZonePressure

	// Reinforce extends a signal's TTL (owner-only).
	Reinforce(signalID, actorID string, newExpiresAtUnix int64) error

	// Kill removes a signal immediately (owner-only).
	Kill(signalID, actorID string) error
}

// ClaimStore is the interface for enforcement layer implementations.
// Any implementation that satisfies this interface can be validated
// by the conformance runner.
type ClaimStore interface {
	// Acquire attempts to place a claim on a zone.
	Acquire(zone, mode, actorID, reason string, ttlSeconds int64) (Claim, error)

	// Release explicitly releases a claim (owner-only).
	Release(claimID, actorID string) error

	// Renew extends a claim's TTL (owner-only).
	Renew(claimID, actorID string, newTTLSeconds int64) (Claim, error)

	// Override force-releases a claim regardless of ownership (operator-only).
	Override(claimID, operatorID, reason string) error

	// ListClaims returns active claims matching the optional zone glob and actor filter.
	ListClaims(zoneGlob, actorID string) []Claim
}

// Ensure MemSignalStore implements SignalStore.
var _ SignalStore = (*MemSignalStore)(nil)

// Ensure MemClaimStore implements ClaimStore.
var _ ClaimStore = (*MemClaimStore)(nil)
