package substrate

import "sync"

// MemClaimStore is an in-memory implementation of the enforcement claim store.
type MemClaimStore struct {
	mu     sync.RWMutex
	claims map[string]Claim
	clock  Clock
}

// NewMemClaimStore creates a new in-memory claim store.
func NewMemClaimStore(clock Clock) *MemClaimStore {
	if clock == nil {
		clock = RealClock{}
	}
	return &MemClaimStore{
		claims: make(map[string]Claim),
		clock:  clock,
	}
}

// AcquireOpts provides optional parameters for claim acquisition.
type AcquireOpts struct {
	ActorKind string
	Origin    string // "local" or "remote"
}

// Acquire attempts to create a claim. Returns the claim or a ConflictError.
func (s *MemClaimStore) Acquire(zone, mode, actorID, reason string, ttlSeconds int64) (Claim, error) {
	return s.AcquireWithOpts(zone, mode, actorID, reason, ttlSeconds, AcquireOpts{})
}

// AcquireWithOpts attempts to create a claim with additional options.
func (s *MemClaimStore) AcquireWithOpts(zone, mode, actorID, reason string, ttlSeconds int64, opts AcquireOpts) (Claim, error) {
	if zone == "" {
		return Claim{}, &ValidationError{Field: "zone", Reason: "empty_zone"}
	}
	if actorID == "" {
		return Claim{}, &ValidationError{Field: "actor_id", Reason: "empty_actor_id"}
	}
	if mode != ModeHard && mode != ModeSoft {
		return Claim{}, &ValidationError{Field: "mode", Reason: "invalid_mode"}
	}
	if opts.Origin == "remote" && ttlSeconds <= 0 {
		return Claim{}, &ValidationError{Field: "ttl_seconds", Reason: "remote_requires_ttl"}
	}

	now := s.clock.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.gcLocked(now.Unix())

	if mode == ModeHard {
		for _, c := range s.claims {
			if c.Zone == zone && c.Mode == ModeHard && c.State == StateActive && c.ActorID != actorID {
				return Claim{}, &ConflictError{
					ConflictingClaimID:       c.ID,
					ConflictingActorID:       c.ActorID,
					ConflictingReason:        c.Reason,
					ConflictingExpiresAtUnix: c.ExpiresAtUnix,
				}
			}
		}
	}

	claim := Claim{
		ID:            generateID("claim"),
		Zone:          zone,
		Mode:          mode,
		State:         StateActive,
		ActorID:       actorID,
		Reason:        reason,
		CreatedAtUnix: now.Unix(),
		UpdatedAtUnix: now.Unix(),
	}
	if ttlSeconds > 0 {
		claim.ExpiresAtUnix = now.Unix() + ttlSeconds
	}

	s.claims[claim.ID] = claim
	return claim, nil
}

// Release explicitly releases a claim owned by the given actor.
func (s *MemClaimStore) Release(claimID, actorID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	claim, ok := s.claims[claimID]
	if !ok {
		return nil // idempotent
	}
	if claim.State != StateActive {
		return nil // already released/expired
	}
	if actorID != "" && claim.ActorID != actorID {
		return &ValidationError{Field: "actor_id", Reason: "actor_mismatch"}
	}

	claim.State = StateReleased
	claim.UpdatedAtUnix = s.clock.Now().Unix()
	s.claims[claimID] = claim
	return nil
}

// Renew extends the TTL of a claim owned by the given actor.
func (s *MemClaimStore) Renew(claimID, actorID string, newTTLSeconds int64) (Claim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	claim, ok := s.claims[claimID]
	if !ok {
		return Claim{}, &ValidationError{Field: "claim_id", Reason: "not_found"}
	}
	if claim.State != StateActive {
		return Claim{}, &ValidationError{Field: "state", Reason: "not_active"}
	}
	if claim.ActorID != actorID {
		return Claim{}, &ValidationError{Field: "actor_id", Reason: "actor_mismatch"}
	}

	now := s.clock.Now()
	claim.ExpiresAtUnix = now.Unix() + newTTLSeconds
	claim.UpdatedAtUnix = now.Unix()
	s.claims[claimID] = claim
	return claim, nil
}

// Override force-releases a claim regardless of ownership.
func (s *MemClaimStore) Override(claimID, operatorID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	claim, ok := s.claims[claimID]
	if !ok {
		return &ValidationError{Field: "claim_id", Reason: "not_found"}
	}
	if claim.State != StateActive {
		return nil // already released
	}

	now := s.clock.Now()
	claim.State = StateOverridden
	claim.OverrideActorID = operatorID
	claim.OverrideReason = reason
	claim.OverrideAtUnix = now.Unix()
	claim.UpdatedAtUnix = now.Unix()
	s.claims[claimID] = claim
	return nil
}

// ListClaims returns active claims, optionally filtered by zone glob and/or actor.
func (s *MemClaimStore) ListClaims(zoneGlob, actorID string) []Claim {
	s.mu.Lock()
	s.gcLocked(s.clock.Now().Unix())
	s.mu.Unlock()

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Claim
	for _, c := range s.claims {
		if c.State != StateActive {
			continue
		}
		if zoneGlob != "" && !ZoneMatch(zoneGlob, c.Zone) {
			continue
		}
		if actorID != "" && c.ActorID != actorID {
			continue
		}
		result = append(result, c)
	}
	return result
}

// gcLocked removes expired claims. Must be called with s.mu held.
func (s *MemClaimStore) gcLocked(nowUnix int64) {
	for id, c := range s.claims {
		if c.State == StateActive && c.ExpiresAtUnix > 0 && c.ExpiresAtUnix <= nowUnix {
			c.State = StateExpired
			c.UpdatedAtUnix = nowUnix
			s.claims[id] = c
		}
	}
	// Remove non-active claims older than retention (keep for audit for 1h)
	for id, c := range s.claims {
		if c.State != StateActive && nowUnix-c.UpdatedAtUnix > 3600 {
			delete(s.claims, id)
		}
	}
}
