// Package port defines the shared logging interfaces consumed by every
// hexagonal domain (notification, message, encryption, storage).
//
// Previously each domain declared its own structurally identical LoggerPort
// and LogEvent interfaces in ports/secondary and ports/contracts. That
// duplication meant adding a single field (e.g. Int64) required touching four
// ports, four contracts, four adapters and a fistful of test mocks, and it had
// already drifted (storage carried Int32/Int64 while notification did not).
//
// These interfaces are the single source of truth. Domains alias them so a new
// field is added in exactly one place. The interfaces carry the superset of
// every fluent builder method the shared logging.Event provides, which lets
// *logging.Event satisfy LogEvent directly without a per-domain wrapper.
package port

import "time"

// LogEvent represents a structured logging event that can be enriched with
// contextual key/value fields and finalized with a message. Builder methods
// return LogEvent so calls can be chained; Msg (or Msgf) terminates the chain
// and emits the record.
type LogEvent interface {
	// Err attaches an error to the event. A nil error is ignored.
	Err(error) LogEvent
	// Str attaches a string field.
	Str(string, string) LogEvent
	// Int attaches an int field.
	Int(string, int) LogEvent
	// Int32 attaches an int32 field without forcing callers to widen at the
	// call site (e.g. protobuf scalars).
	Int32(string, int32) LogEvent
	// Int64 attaches an int64 field for values that exceed 32-bit range such
	// as sql.Result.RowsAffected.
	Int64(string, int64) LogEvent
	// Bool attaches a boolean field.
	Bool(string, bool) LogEvent
	// Dur attaches a duration field.
	Dur(string, time.Duration) LogEvent
	// Float64 attaches a float64 field.
	Float64(string, float64) LogEvent
	// Interface attaches an arbitrary value field.
	Interface(string, any) LogEvent
	// Msg finalizes the event with a message and emits it. Call last.
	Msg(string)
	// Msgf finalizes the event with a formatted message and emits it.
	Msgf(string, ...any)
}

// Logger starts log events at a given severity. Each method returns a fresh
// LogEvent that must be finalized with Msg/Msgf to be emitted.
type Logger interface {
	// Debug starts a debug-level event for diagnostic output.
	Debug() LogEvent
	// Info starts an info-level event for routine operational messages.
	Info() LogEvent
	// Warn starts a warn-level event for recoverable anomalies.
	Warn() LogEvent
	// Error starts an error-level event for failures.
	Error() LogEvent
}
