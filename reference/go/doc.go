// Package substrate provides the reference implementation of the
// Agent Coordination Substrate specification.
//
// It implements both the advisory layer (signals, zones, pressure readouts,
// induction) and the enforcement layer (claims, conflict detection, override).
//
// Two store implementations are provided:
//   - memstore: in-memory, suitable for testing and single-process deployments
//   - (planned) sqlitestore: persistent, suitable for crash-safe deployments
//
// The public API mirrors the spec operations: Deposit, Readout, Acquire,
// Release, Override.
package substrate
