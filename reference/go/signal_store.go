package substrate

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// MemSignalStore is an in-memory implementation of the advisory signal store.
type MemSignalStore struct {
	mu      sync.RWMutex
	signals map[string]Signal
	clock   Clock
}

// NewMemSignalStore creates a new in-memory signal store.
func NewMemSignalStore(clock Clock) *MemSignalStore {
	if clock == nil {
		clock = RealClock{}
	}
	return &MemSignalStore{
		signals: make(map[string]Signal),
		clock:   clock,
	}
}

// Deposit adds a signal to the store. Returns the signal (with generated ID if empty) or an error.
func (s *MemSignalStore) Deposit(sig Signal) (Signal, error) {
	if sig.Zone == "" {
		return Signal{}, &ValidationError{Field: "zone", Reason: "empty_zone"}
	}
	if sig.ActorID == "" {
		return Signal{}, &ValidationError{Field: "actor_id", Reason: "empty_actor_id"}
	}
	if sig.ExpiresAtUnix <= sig.CreatedAtUnix {
		return Signal{}, &ValidationError{Field: "expires_at_unix", Reason: "invalid_ttl"}
	}
	if sig.Kind == "" {
		return Signal{}, &ValidationError{Field: "kind", Reason: "empty_kind"}
	}
	if sig.StrengthMilli < 0 {
		return Signal{}, &ValidationError{Field: "strength_milli", Reason: "must_be_positive"}
	}
	if sig.StrengthMilli == 0 {
		sig.StrengthMilli = 1000
	}
	if sig.ID == "" {
		sig.ID = generateID("sig")
	}

	s.mu.Lock()
	s.gcLocked(s.clock.Now().Unix())
	s.signals[sig.ID] = sig
	s.mu.Unlock()
	return sig, nil
}

// Readout returns the zone pressure for the given zone.
// Does not mutate the store; expired signals are filtered at read time.
func (s *MemSignalStore) Readout(zone string) ZonePressure {
	now := s.clock.Now().Unix()

	s.mu.RLock()
	defer s.mu.RUnlock()

	var signals []Signal
	actors := make(map[string]struct{})
	sessions := make(map[string]struct{})

	for _, sig := range s.signals {
		if sig.Zone == zone {
			if sig.ExpiresAtUnix > 0 && sig.ExpiresAtUnix <= now {
				continue // expired, skip without mutating
			}
			signals = append(signals, sig)
			actors[sig.ActorID] = struct{}{}
			if sig.SessionID != "" {
				sessions[sig.SessionID] = struct{}{}
			}
		}
	}

	lineages := len(actors)
	pressure := classifyPressure(signals, lineages)
	trend := classifyTrend(signals, now)

	return ZonePressure{
		Zone:                zone,
		DepositCount:        len(signals),
		SessionCount:        len(sessions),
		IndependentLineages: lineages,
		Pressure:            pressure,
		Trend:               trend,
		ContributingSignals: signals,
	}
}

// ListZones returns zone pressure readouts for all zones matching the glob.
// Does not mutate the store; expired signals are filtered by Readout.
func (s *MemSignalStore) ListZones(glob string) []ZonePressure {
	now := s.clock.Now().Unix()

	s.mu.RLock()
	zones := make(map[string]struct{})
	for _, sig := range s.signals {
		if sig.ExpiresAtUnix > 0 && sig.ExpiresAtUnix <= now {
			continue
		}
		if glob == "" || ZoneMatch(glob, sig.Zone) {
			zones[sig.Zone] = struct{}{}
		}
	}
	s.mu.RUnlock()

	var results []ZonePressure
	for z := range zones {
		results = append(results, s.Readout(z))
	}
	return results
}

// Reinforce extends the TTL of a signal owned by the given actor.
func (s *MemSignalStore) Reinforce(signalID, actorID string, newExpiresAtUnix int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sig, ok := s.signals[signalID]
	if !ok {
		return &ValidationError{Field: "signal_id", Reason: "not_found"}
	}
	if sig.ActorID != actorID {
		return &ValidationError{Field: "actor_id", Reason: "actor_mismatch"}
	}
	sig.ExpiresAtUnix = newExpiresAtUnix
	s.signals[signalID] = sig
	return nil
}

// Kill immediately removes a signal owned by the given actor (or by an operator).
func (s *MemSignalStore) Kill(signalID, actorID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sig, ok := s.signals[signalID]
	if !ok {
		return &ValidationError{Field: "signal_id", Reason: "not_found"}
	}
	if sig.ActorID != actorID {
		return &ValidationError{Field: "actor_id", Reason: "actor_mismatch"}
	}
	delete(s.signals, signalID)
	return nil
}

// GC removes expired signals from the store. Callers may invoke this
// periodically to reclaim memory; Deposit also triggers gc on each write.
func (s *MemSignalStore) GC() {
	s.mu.Lock()
	s.gcLocked(s.clock.Now().Unix())
	s.mu.Unlock()
}

// gcLocked removes expired signals. Must be called with s.mu write-held.
func (s *MemSignalStore) gcLocked(nowUnix int64) {
	for id, sig := range s.signals {
		if sig.ExpiresAtUnix > 0 && sig.ExpiresAtUnix <= nowUnix {
			delete(s.signals, id)
		}
	}
}

func classifyPressure(signals []Signal, lineages int) Pressure {
	if len(signals) == 0 {
		return Pressure{State: PressureQuiet, Score: 0}
	}
	if lineages >= 3 {
		return Pressure{
			State:               PressureConvergent,
			Score:               float64(lineages) / 5.0,
			IndependentLineages: lineages,
		}
	}
	if lineages >= 2 {
		return Pressure{
			State:               PressureWarming,
			Score:               float64(len(signals)) / 10.0,
			IndependentLineages: lineages,
		}
	}
	return Pressure{
		State:               PressureActive,
		Score:               float64(len(signals)) / 10.0,
		IndependentLineages: lineages,
	}
}

func classifyTrend(signals []Signal, nowUnix int64) Trend {
	if len(signals) == 0 {
		return Trend{State: TrendStale}
	}
	shortWindow := int64(3600)
	mediumWindow := int64(86400)

	var newest, oldest int64
	var shortCount, recentCount int
	for _, sig := range signals {
		if oldest == 0 || sig.CreatedAtUnix < oldest {
			oldest = sig.CreatedAtUnix
		}
		if sig.CreatedAtUnix > newest {
			newest = sig.CreatedAtUnix
		}
		if nowUnix-sig.CreatedAtUnix <= shortWindow {
			shortCount++
		}
		if nowUnix-sig.CreatedAtUnix <= mediumWindow {
			recentCount++
		}
	}

	state := TrendStale
	if shortCount > 0 {
		state = TrendCurrent
	} else if recentCount > 0 {
		state = TrendRecent
	}

	return Trend{
		State:                     state,
		CurrentDepositCount:       shortCount,
		RecentDepositCount:        recentCount,
		ShortWindowSeconds:        int(shortWindow),
		MediumWindowSeconds:       int(mediumWindow),
		NewestSignalCreatedAtUnix: newest,
		OldestSignalCreatedAtUnix: oldest,
	}
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}
