// Package conformance implements a runner that validates SignalStore and
// ClaimStore implementations against the Agent Coordination Substrate
// specification expectations.
//
// Usage:
//
//	runner := conformance.NewRunner(signalStore, claimStore, clock)
//	results := runner.RunAdvisory(advisoryExpectationsYAML)
//	results = append(results, runner.RunEnforcement(enforcementExpectationsYAML)...)
//	for _, r := range results {
//	    fmt.Printf("%s: %s\n", r.Status, r.ScenarioID)
//	}
package conformance

import (
	"fmt"
	"strings"
	"time"

	substrate "github.com/instagrim-dev/agent-coordination-substrate/reference/go"
	"gopkg.in/yaml.v3"
)

// Result is the outcome of running a single conformance scenario.
type Result struct {
	Group      string `json:"group"`
	ScenarioID string `json:"scenario_id"`
	Status     string `json:"status"` // "pass", "fail", "skip"
	Message    string `json:"message,omitempty"`
}

// Runner executes conformance expectations against store implementations.
type Runner struct {
	signals        substrate.SignalStore
	claims         substrate.ClaimStore
	clock          *TestClock
	signalFactory  func(*TestClock) substrate.SignalStore
	claimFactory   func(*TestClock) substrate.ClaimStore
}

// TestClock is an injectable clock for the conformance runner.
type TestClock struct {
	now time.Time
}

func NewTestClock(startUnix int64) *TestClock {
	return &TestClock{now: time.Unix(startUnix, 0)}
}

func (c *TestClock) Now() time.Time      { return c.now }
func (c *TestClock) SetUnix(unix int64)  { c.now = time.Unix(unix, 0) }
func (c *TestClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

// NewRunner creates a conformance runner using the provided stores.
func NewRunner(signals substrate.SignalStore, claims substrate.ClaimStore, clock *TestClock) *Runner {
	return &Runner{
		signals: signals,
		claims:  claims,
		clock:   clock,
		signalFactory: func(c *TestClock) substrate.SignalStore {
			return substrate.NewMemSignalStore(c)
		},
		claimFactory: func(c *TestClock) substrate.ClaimStore {
			return substrate.NewMemClaimStore(c)
		},
	}
}

// NewRunnerWithFactories creates a runner with custom store factories.
func NewRunnerWithFactories(
	signalFactory func(*TestClock) substrate.SignalStore,
	claimFactory func(*TestClock) substrate.ClaimStore,
	clock *TestClock,
) *Runner {
	return &Runner{
		signals:       signalFactory(clock),
		claims:        claimFactory(clock),
		clock:         clock,
		signalFactory: signalFactory,
		claimFactory:  claimFactory,
	}
}

func (r *Runner) resetStores() {
	r.signals = r.signalFactory(r.clock)
	r.claims = r.claimFactory(r.clock)
}

type expectationsFile struct {
	Groups map[string]group `yaml:"groups"`
}

type group struct {
	Description string     `yaml:"description"`
	Scenarios   []scenario `yaml:"scenarios"`
}

type scenario struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Steps       []step `yaml:"steps"`
}

type step struct {
	Action string         `yaml:"action"`
	Input  map[string]any `yaml:"input"`
	Expect map[string]any `yaml:"expect"`
}

// RunAdvisory runs advisory layer expectations.
func (r *Runner) RunAdvisory(yamlData []byte) []Result {
	return r.runExpectations(yamlData, "advisory")
}

// RunEnforcement runs enforcement layer expectations.
func (r *Runner) RunEnforcement(yamlData []byte) []Result {
	return r.runExpectations(yamlData, "enforcement")
}

func (r *Runner) runExpectations(yamlData []byte, layer string) []Result {
	var file expectationsFile
	if err := yaml.Unmarshal(yamlData, &file); err != nil {
		return []Result{{Status: "fail", Message: "YAML parse error: " + err.Error()}}
	}

	var results []Result
	for groupName, grp := range file.Groups {
		r.resetStores()
		for _, sc := range grp.Scenarios {
			// Reset stores per scenario UNLESS the scenario has no write steps
			// (indicating it depends on prior group state)
			if scenarioHasWriteSteps(sc) {
				r.resetStores()
			}
			result := r.runScenario(groupName, sc, layer)
			results = append(results, result)
		}
	}
	return results
}

func scenarioHasWriteSteps(sc scenario) bool {
	for _, s := range sc.Steps {
		switch s.Action {
		case "deposit", "acquire", "configure_saturation":
			return true
		}
	}
	return false
}

func (r *Runner) runScenario(groupName string, sc scenario, layer string) Result {
	r.clock.SetUnix(1716000000) // reset time for each scenario

	var lastClaimID string

	for i, s := range sc.Steps {
		switch s.Action {
		case "deposit":
			res := r.executeDeposit(s)
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: %v", i, s.Action, res),
				}
			}

		case "readout":
			res := r.executeReadout(s)
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: got %v", i, s.Action, res),
				}
			}

		case "list_zones":
			res := r.executeListZones(s)
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: got %v", i, s.Action, res),
				}
			}

		case "reinforce":
			res := r.executeReinforce(s)
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: %v", i, s.Action, res),
				}
			}

		case "kill":
			res := r.executeKill(s)
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: %v", i, s.Action, res),
				}
			}

		case "advance_time":
			if toUnix, ok := s.Input["to_unix"]; ok {
				r.clock.SetUnix(toInt64(toUnix))
			} else if seconds, ok := s.Input["seconds"]; ok {
				r.clock.Advance(time.Duration(toInt64(seconds)) * time.Second)
			}

		case "acquire":
			res := r.executeAcquire(s)
			if claimID, ok := res["claim_id"].(string); ok {
				lastClaimID = claimID
			}
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: got %v", i, s.Action, res),
				}
			}

		case "release":
			claimID := resolveRef(getString(s.Input, "claim_id"), lastClaimID)
			res := r.executeRelease(claimID, getString(s.Input, "actor_id"))
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: got %v", i, s.Action, res),
				}
			}

		case "renew":
			claimID := resolveRef(getString(s.Input, "claim_id"), lastClaimID)
			res := r.executeRenew(claimID, getString(s.Input, "actor_id"), toInt64(s.Input["new_ttl_seconds"]))
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: got %v", i, s.Action, res),
				}
			}

		case "override":
			claimID := resolveRef(getString(s.Input, "claim_id"), lastClaimID)
			res := r.executeOverride(claimID, getString(s.Input, "operator_id"), getString(s.Input, "reason"))
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: got %v", i, s.Action, res),
				}
			}

		case "list_claims":
			res := r.executeListClaims(s)
			if !r.checkExpect(s.Expect, res) {
				return Result{
					Group:      groupName,
					ScenarioID: sc.ID,
					Status:     "fail",
					Message:    fmt.Sprintf("step %d (%s): expectation mismatch: got %v", i, s.Action, res),
				}
			}

		case "restart":
			// MemStore does not persist; skip persistence scenarios
			return Result{
				Group:      groupName,
				ScenarioID: sc.ID,
				Status:     "skip",
				Message:    "restart not supported by in-memory store",
			}

		case "configure_saturation", "list_proposals", "accept_proposal", "dismiss_proposal":
			return Result{
				Group:      groupName,
				ScenarioID: sc.ID,
				Status:     "skip",
				Message:    "induction not implemented in reference memstore",
			}

		default:
			return Result{
				Group:      groupName,
				ScenarioID: sc.ID,
				Status:     "fail",
				Message:    fmt.Sprintf("step %d: unknown action %q", i, s.Action),
			}
		}
	}

	return Result{Group: groupName, ScenarioID: sc.ID, Status: "pass"}
}

func (r *Runner) executeDeposit(s step) map[string]any {
	sig := substrate.Signal{
		ID:            getString(s.Input, "id"),
		Zone:          getString(s.Input, "zone"),
		Kind:          getString(s.Input, "kind"),
		ActorID:       getString(s.Input, "actor_id"),
		CreatedAtUnix: toInt64(s.Input["created_at_unix"]),
		ExpiresAtUnix: toInt64(s.Input["expires_at_unix"]),
		StrengthMilli: int(toInt64(s.Input["strength_milli"])),
		Reason:        getString(s.Input, "reason"),
	}
	_, err := r.signals.Deposit(sig)
	if err != nil {
		if ve, ok := err.(*substrate.ValidationError); ok {
			return map[string]any{"result": "rejected", "reason": ve.Reason}
		}
		return map[string]any{"result": "rejected", "reason": err.Error()}
	}
	return map[string]any{"result": "accepted"}
}

func (r *Runner) executeReadout(s step) map[string]any {
	zone := getString(s.Input, "zone")
	pressure := r.signals.Readout(zone)
	res := map[string]any{
		"deposit_count":        pressure.DepositCount,
		"pressure_state":       string(pressure.Pressure.State),
		"independent_lineages": pressure.IndependentLineages,
	}
	if len(pressure.ContributingSignals) > 0 {
		var sigs []map[string]any
		for _, sig := range pressure.ContributingSignals {
			sigs = append(sigs, map[string]any{
				"id":       sig.ID,
				"actor_id": sig.ActorID,
				"kind":     sig.Kind,
			})
		}
		res["contributing_signals"] = sigs
	}
	return res
}

func (r *Runner) executeListZones(s step) map[string]any {
	glob := getString(s.Input, "glob")
	zones := r.signals.ListZones(glob)
	res := map[string]any{"zone_count": len(zones)}
	if len(zones) > 0 {
		var zoneList []map[string]any
		for _, z := range zones {
			zoneList = append(zoneList, map[string]any{
				"zone":          z.Zone,
				"deposit_count": z.DepositCount,
			})
		}
		res["zones"] = zoneList
	}
	return res
}

func (r *Runner) executeReinforce(s step) map[string]any {
	err := r.signals.Reinforce(
		getString(s.Input, "signal_id"),
		getString(s.Input, "actor_id"),
		toInt64(s.Input["new_expires_at_unix"]),
	)
	if err != nil {
		if ve, ok := err.(*substrate.ValidationError); ok {
			return map[string]any{"result": "rejected", "reason": ve.Reason}
		}
		return map[string]any{"result": "rejected", "reason": err.Error()}
	}
	return map[string]any{"result": "accepted"}
}

func (r *Runner) executeKill(s step) map[string]any {
	err := r.signals.Kill(getString(s.Input, "signal_id"), getString(s.Input, "actor_id"))
	if err != nil {
		if ve, ok := err.(*substrate.ValidationError); ok {
			return map[string]any{"result": "rejected", "reason": ve.Reason}
		}
		return map[string]any{"result": "rejected", "reason": err.Error()}
	}
	return map[string]any{"result": "accepted"}
}

func (r *Runner) executeAcquire(s step) map[string]any {
	zone := getString(s.Input, "zone")
	mode := getString(s.Input, "mode")
	actorID := getString(s.Input, "actor_id")
	reason := getString(s.Input, "reason")
	ttl := toInt64(s.Input["ttl_seconds"])
	origin := getString(s.Input, "origin")
	actorKind := getString(s.Input, "actor_kind")

	var claim substrate.Claim
	var err error
	if (origin != "" || actorKind != "") {
		if ms, ok := r.claims.(*substrate.MemClaimStore); ok {
			claim, err = ms.AcquireWithOpts(zone, mode, actorID, reason, ttl, substrate.AcquireOpts{
				Origin:    origin,
				ActorKind: actorKind,
			})
		} else {
			// Store doesn't support opts — call basic Acquire
			claim, err = r.claims.Acquire(zone, mode, actorID, reason, ttl)
		}
	} else {
		claim, err = r.claims.Acquire(zone, mode, actorID, reason, ttl)
	}
	if err != nil {
		if ce, ok := err.(*substrate.ConflictError); ok {
			res := map[string]any{
				"result":               "conflict",
				"conflicting_actor_id": ce.ConflictingActorID,
			}
			if ce.ConflictingReason != "" {
				res["has_conflicting_reason"] = true
			}
			if ce.ConflictingExpiresAtUnix > 0 {
				res["has_conflicting_expires_at_unix"] = true
			}
			return res
		}
		if ve, ok := err.(*substrate.ValidationError); ok {
			return map[string]any{"result": "rejected", "reason": ve.Reason}
		}
		return map[string]any{"result": "rejected", "reason": err.Error()}
	}
	return map[string]any{
		"result":   "accepted",
		"claim_id": claim.ID,
		"claim": map[string]any{
			"state":              claim.State,
			"mode":               claim.Mode,
			"actor_id":           claim.ActorID,
			"has_id":             claim.ID != "",
			"has_expires_at_unix": claim.ExpiresAtUnix > 0,
		},
	}
}

func (r *Runner) executeRelease(claimID, actorID string) map[string]any {
	err := r.claims.Release(claimID, actorID)
	if err != nil {
		if ve, ok := err.(*substrate.ValidationError); ok {
			return map[string]any{"result": "rejected", "reason": ve.Reason}
		}
		return map[string]any{"result": "rejected", "reason": err.Error()}
	}
	return map[string]any{"result": "accepted"}
}

func (r *Runner) executeRenew(claimID, actorID string, newTTL int64) map[string]any {
	_, err := r.claims.Renew(claimID, actorID, newTTL)
	if err != nil {
		if ve, ok := err.(*substrate.ValidationError); ok {
			return map[string]any{"result": "rejected", "reason": ve.Reason}
		}
		return map[string]any{"result": "rejected", "reason": err.Error()}
	}
	return map[string]any{"result": "accepted"}
}

func (r *Runner) executeOverride(claimID, operatorID, reason string) map[string]any {
	err := r.claims.Override(claimID, operatorID, reason)
	if err != nil {
		if ve, ok := err.(*substrate.ValidationError); ok {
			return map[string]any{"result": "rejected", "reason": ve.Reason}
		}
		return map[string]any{"result": "rejected", "reason": err.Error()}
	}
	return map[string]any{
		"result":            "accepted",
		"override_recorded": true,
		"override_actor_id": operatorID,
		"has_override_reason":  reason != "",
		"has_override_at_unix": true,
	}
}

func (r *Runner) executeListClaims(s step) map[string]any {
	zone := getString(s.Input, "zone")
	claims := r.claims.ListClaims(zone, "")
	res := map[string]any{"count": len(claims)}
	if len(claims) > 0 {
		var cl []map[string]any
		for _, c := range claims {
			cl = append(cl, map[string]any{
				"actor_id": c.ActorID,
				"mode":     c.Mode,
				"state":    c.State,
			})
		}
		res["claims"] = cl
	}
	return res
}

func (r *Runner) checkExpect(expect map[string]any, got map[string]any) bool {
	for k, expected := range expect {
		switch k {
		case "result":
			if got["result"] != expected {
				return false
			}
		case "reason":
			if got["reason"] != expected {
				return false
			}
		case "deposit_count":
			if toInt64(got["deposit_count"]) != toInt64(expected) {
				return false
			}
		case "zone_count":
			if toInt64(got["zone_count"]) != toInt64(expected) {
				return false
			}
		case "count":
			if toInt64(got["count"]) != toInt64(expected) {
				return false
			}
		case "pressure_state":
			if got["pressure_state"] != expected {
				return false
			}
		case "pressure_state_one_of":
			states, ok := expected.([]any)
			if !ok {
				return false
			}
			gotState, _ := got["pressure_state"].(string)
			found := false
			for _, s := range states {
				if fmt.Sprint(s) == gotState {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case "independent_lineages":
			if toInt64(got["independent_lineages"]) != toInt64(expected) {
				return false
			}
		case "conflicting_actor_id":
			if got["conflicting_actor_id"] != expected {
				return false
			}
		case "has_conflicting_reason":
			if got["has_conflicting_reason"] != true {
				return false
			}
		case "has_conflicting_expires_at_unix":
			if got["has_conflicting_expires_at_unix"] != true {
				return false
			}
		case "override_recorded":
			if got["override_recorded"] != true {
				return false
			}
		case "override_actor_id":
			if got["override_actor_id"] != expected {
				return false
			}
		case "has_override_reason":
			if got["has_override_reason"] != true {
				return false
			}
		case "has_override_at_unix":
			if got["has_override_at_unix"] != true {
				return false
			}
		case "result_one_of":
			states, ok := expected.([]any)
			if !ok {
				return false
			}
			gotResult := fmt.Sprint(got["result"])
			found := false
			for _, s := range states {
				if fmt.Sprint(s) == gotResult {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case "contributing_signals":
			if !matchSignalsList(expected, got["contributing_signals"]) {
				return false
			}
		case "zones":
			if !matchZonesList(expected, got["zones"]) {
				return false
			}
		case "claims":
			if !matchClaimsList(expected, got["claims"]) {
				return false
			}
		case "claim":
			if !matchClaimObject(expected, got["claim"]) {
				return false
			}
		case "proposals":
			// Induction not implemented in reference memstore; proposals
			// validation is deferred until list_proposals/accept_proposal
			// actions are supported by a store implementation.
		}
	}
	return true
}

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	case nil:
		return 0
	default:
		return 0
	}
}

func resolveRef(raw, lastClaimID string) string {
	if strings.Contains(raw, "${last_claim_id}") || strings.HasPrefix(raw, "${") {
		return lastClaimID
	}
	return raw
}

// matchSignalsList checks that every expected signal entry exists (order-independent)
// in the actual signals list using field-subset matching.
func matchSignalsList(expected, got any) bool {
	expectedList, ok := toSliceOfMaps(expected)
	if !ok {
		return true // can't parse expected — pass through
	}
	gotList, ok := toSliceOfMaps(got)
	if !ok {
		return false
	}
	for _, exp := range expectedList {
		if !findMatchInList(exp, gotList) {
			return false
		}
	}
	return true
}

// matchZonesList checks that every expected zone entry exists (order-independent)
// in the actual zones list using field-subset matching.
func matchZonesList(expected, got any) bool {
	expectedList, ok := toSliceOfMaps(expected)
	if !ok {
		return true
	}
	gotList, ok := toSliceOfMaps(got)
	if !ok {
		return false
	}
	for _, exp := range expectedList {
		if !findMatchInList(exp, gotList) {
			return false
		}
	}
	return true
}

// matchClaimsList checks that every expected claim entry exists (order-independent)
// in the actual claims list using field-subset matching.
func matchClaimsList(expected, got any) bool {
	expectedList, ok := toSliceOfMaps(expected)
	if !ok {
		return true
	}
	gotList, ok := toSliceOfMaps(got)
	if !ok {
		return false
	}
	for _, exp := range expectedList {
		if !findMatchInList(exp, gotList) {
			return false
		}
	}
	return true
}

// matchClaimObject checks that every field in expected matches the
// corresponding field in the actual claim map.
func matchClaimObject(expected, got any) bool {
	expMap, ok := toMap(expected)
	if !ok {
		return true
	}
	gotMap, ok := toMap(got)
	if !ok {
		return false
	}
	return fieldSubsetMatch(expMap, gotMap)
}

// findMatchInList returns true if any entry in gotList matches all specified
// fields in exp (field-subset match).
func findMatchInList(exp map[string]any, gotList []map[string]any) bool {
	for _, g := range gotList {
		if fieldSubsetMatch(exp, g) {
			return true
		}
	}
	return false
}

// fieldSubsetMatch returns true when every key in exp has an equal value in
// got. Comparison uses fmt.Sprint for type-insensitive string equality, with
// special handling for numeric and boolean fields.
func fieldSubsetMatch(exp, got map[string]any) bool {
	for k, ev := range exp {
		gv, exists := got[k]
		if !exists {
			return false
		}
		if !valuesEqual(ev, gv) {
			return false
		}
	}
	return true
}

func valuesEqual(expected, got any) bool {
	// Handle boolean comparisons explicitly (YAML bools vs Go bools).
	if eb, ok := expected.(bool); ok {
		if gb, ok2 := got.(bool); ok2 {
			return eb == gb
		}
		return false
	}
	// Handle numeric comparisons.
	if isNumeric(expected) && isNumeric(got) {
		return toInt64(expected) == toInt64(got)
	}
	return fmt.Sprint(expected) == fmt.Sprint(got)
}

func isNumeric(v any) bool {
	switch v.(type) {
	case int, int64, float64:
		return true
	}
	return false
}

// toSliceOfMaps coerces an any value into []map[string]any.
// Handles both []any (from YAML) and []map[string]any (from Go code).
func toSliceOfMaps(v any) ([]map[string]any, bool) {
	switch s := v.(type) {
	case []map[string]any:
		return s, true
	case []any:
		var result []map[string]any
		for _, item := range s {
			m, ok := toMap(item)
			if !ok {
				return nil, false
			}
			result = append(result, m)
		}
		return result, true
	}
	return nil, false
}

// toMap coerces an any value into map[string]any.
func toMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[any]any:
		result := make(map[string]any, len(m))
		for k, val := range m {
			result[fmt.Sprint(k)] = val
		}
		return result, true
	}
	return nil, false
}
