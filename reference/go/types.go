package substrate

import "time"

// Signal represents an advisory deposit in the coordination substrate.
type Signal struct {
	ID            string            `json:"id"`
	Zone          string            `json:"zone"`
	Kind          string            `json:"kind"`
	ActorID       string            `json:"actor_id"`
	ActorKind     string            `json:"actor_kind,omitempty"`
	CreatedAtUnix int64             `json:"created_at_unix"`
	ExpiresAtUnix int64             `json:"expires_at_unix"`
	StrengthMilli int               `json:"strength_milli"`
	Properties    map[string]any    `json:"properties,omitempty"`
	Source        string            `json:"source,omitempty"`
	Reason        string            `json:"reason,omitempty"`
	SessionID     string            `json:"session_id,omitempty"`
	WorkflowID    string            `json:"workflow_id,omitempty"`
}

// Claim represents an enforcement primitive in the coordination substrate.
type Claim struct {
	ID              string   `json:"id"`
	Zone            string   `json:"zone"`
	Mode            string   `json:"mode"`
	State           string   `json:"state"`
	ActorID         string   `json:"actor_id"`
	ActorKind       string   `json:"actor_kind,omitempty"`
	Reason          string   `json:"reason,omitempty"`
	CreatedAtUnix   int64    `json:"created_at_unix"`
	ExpiresAtUnix   int64    `json:"expires_at_unix,omitempty"`
	UpdatedAtUnix   int64    `json:"updated_at_unix,omitempty"`
	RequestedPaths  []string `json:"requested_paths,omitempty"`
	NegotiationNote string   `json:"negotiation_note,omitempty"`
	OverrideActorID string   `json:"override_actor_id,omitempty"`
	OverrideReason  string   `json:"override_reason,omitempty"`
	OverrideAtUnix  int64    `json:"override_at_unix,omitempty"`
}

// Claim modes.
const (
	ModeHard = "hard"
	ModeSoft = "soft"
)

// Claim states.
const (
	StateActive     = "active"
	StateReleased   = "released"
	StateExpired    = "expired"
	StateOverridden = "overridden"
)

// PressureState represents the classified pressure level of a zone.
type PressureState string

const (
	PressureQuiet             PressureState = "quiet"
	PressureActive            PressureState = "active"
	PressureWarming           PressureState = "warming"
	PressureConvergent        PressureState = "convergent"
	PressureSaturated         PressureState = "saturated"
	PressureContested         PressureState = "contested"
	PressureAvoidable         PressureState = "avoidable"
	PressureBoundarySensitive PressureState = "boundary_sensitive"
)

// TrendState represents the temporal trend of a zone.
type TrendState string

const (
	TrendCurrent TrendState = "current"
	TrendRecent  TrendState = "recent"
	TrendCooling TrendState = "cooling"
	TrendSurging TrendState = "surging"
	TrendStale   TrendState = "stale"
)

// ZonePressure is the readout for a single zone.
type ZonePressure struct {
	Zone               string        `json:"zone"`
	DepositCount       int           `json:"deposit_count"`
	SessionCount       int           `json:"session_count"`
	IndependentLineages int          `json:"independent_lineages"`
	Pressure           Pressure      `json:"pressure"`
	Trend              Trend         `json:"trend"`
	ContributingSignals []Signal     `json:"contributing_signals,omitempty"`
	EvidenceTruncated  bool          `json:"evidence_truncated"`
}

// Pressure holds the classified state and score.
type Pressure struct {
	State            PressureState `json:"state"`
	Score            float64       `json:"score"`
	IndependentLineages int       `json:"independent_lineages"`
}

// Trend holds temporal context.
type Trend struct {
	State                    TrendState `json:"state"`
	CurrentDepositCount      int        `json:"current_deposit_count"`
	RecentDepositCount       int        `json:"recent_deposit_count"`
	ShortWindowSeconds       int        `json:"short_window_seconds"`
	MediumWindowSeconds      int        `json:"medium_window_seconds"`
	NewestSignalCreatedAtUnix int64     `json:"newest_signal_created_at_unix"`
	OldestSignalCreatedAtUnix int64     `json:"oldest_signal_created_at_unix"`
}

// ConflictError is returned when a hard claim conflicts with an existing claim.
type ConflictError struct {
	ConflictingClaimID      string `json:"conflicting_claim_id"`
	ConflictingActorID      string `json:"conflicting_actor_id"`
	ConflictingReason       string `json:"conflicting_reason,omitempty"`
	ConflictingExpiresAtUnix int64 `json:"conflicting_expires_at_unix,omitempty"`
}

func (e *ConflictError) Error() string {
	return "conflict: zone claimed by " + e.ConflictingActorID
}

// ValidationError is returned when input fails validation.
type ValidationError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func (e *ValidationError) Error() string {
	return "validation: " + e.Field + " " + e.Reason
}

// Clock provides time for the substrate. Allows test injection.
type Clock interface {
	Now() time.Time
}

// RealClock uses time.Now.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
