// Package logtest provides reusable test doubles for the shared logging port.
//
// Before this package each hexagonal domain hand-rolled its own logger test
// scaffolding: testify Mock* types, ad-hoc recordingLogger structs, and the
// notorious setupLenientLoggerMock that handed back a single shared event for
// every severity (so a test could not tell whether the code logged at Info or
// Error). Recorder fixes that by capturing the level of every emitted event,
// and Noop offers a zero-assertion logger for tests that do not care about
// log output.
package logtest

import (
	"fmt"
	"sync"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/port"
)

// Entry is a single emitted log record captured by a Recorder.
type Entry struct {
	// Level is the severity the event was started at: "debug", "info",
	// "warn", "error" or "fatal".
	Level string
	// Msg is the finalizing message passed to Msg/Msgf.
	Msg string
	// Err is the error attached via Err, if any.
	Err error
	// Fields holds every key/value attached to the event, including the
	// error under the "error" key when Err was called with a non-nil error.
	Fields map[string]any
}

// Recorder is a port.Logger (also satisfying the encryption domain's
// Fatal-extended LoggerPort) that records every finalized event. It is safe
// for concurrent use.
type Recorder struct {
	mu      sync.Mutex
	entries []Entry
}

// NewRecorder returns an empty Recorder.
func NewRecorder() *Recorder { return &Recorder{} }

func (r *Recorder) start(level string) port.LogEvent {
	return &recorderEvent{rec: r, entry: Entry{Level: level, Fields: map[string]any{}}}
}

// Debug starts a debug-level event.
func (r *Recorder) Debug() port.LogEvent { return r.start("debug") }

// Info starts an info-level event.
func (r *Recorder) Info() port.LogEvent { return r.start("info") }

// Warn starts a warn-level event.
func (r *Recorder) Warn() port.LogEvent { return r.start("warn") }

// Error starts an error-level event.
func (r *Recorder) Error() port.LogEvent { return r.start("error") }

// Fatal starts a fatal-level event. Unlike the production logger it does not
// terminate the process, so tests can assert on fatal logging.
func (r *Recorder) Fatal() port.LogEvent { return r.start("fatal") }

func (r *Recorder) append(e Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, e)
}

// Entries returns a copy of every recorded entry in emission order. Each
// entry's Fields map is also copied, so callers can mutate the result without
// corrupting the recorder's internal state.
func (r *Recorder) Entries() []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Entry, len(r.entries))
	for i, e := range r.entries {
		fields := make(map[string]any, len(e.Fields))
		for k, v := range e.Fields {
			fields[k] = v
		}
		e.Fields = fields
		out[i] = e
	}
	return out
}

// Count returns the number of recorded entries at the given level.
func (r *Recorder) Count(level string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, e := range r.entries {
		if e.Level == level {
			n++
		}
	}
	return n
}

// Messages returns the finalizing messages of every entry at the given level.
func (r *Recorder) Messages(level string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var msgs []string
	for _, e := range r.entries {
		if e.Level == level {
			msgs = append(msgs, e.Msg)
		}
	}
	return msgs
}

type recorderEvent struct {
	rec   *Recorder
	entry Entry
}

func (e *recorderEvent) Err(err error) port.LogEvent {
	e.entry.Err = err
	if err != nil {
		e.entry.Fields["error"] = err
	}
	return e
}

func (e *recorderEvent) Str(key, value string) port.LogEvent {
	e.entry.Fields[key] = value
	return e
}

func (e *recorderEvent) Int(key string, value int) port.LogEvent {
	e.entry.Fields[key] = value
	return e
}

func (e *recorderEvent) Int32(key string, value int32) port.LogEvent {
	e.entry.Fields[key] = value
	return e
}

func (e *recorderEvent) Int64(key string, value int64) port.LogEvent {
	e.entry.Fields[key] = value
	return e
}

func (e *recorderEvent) Bool(key string, value bool) port.LogEvent {
	e.entry.Fields[key] = value
	return e
}

func (e *recorderEvent) Dur(key string, value time.Duration) port.LogEvent {
	e.entry.Fields[key] = value
	return e
}

func (e *recorderEvent) Float64(key string, value float64) port.LogEvent {
	e.entry.Fields[key] = value
	return e
}

func (e *recorderEvent) Interface(key string, value any) port.LogEvent {
	e.entry.Fields[key] = value
	return e
}

func (e *recorderEvent) Msg(msg string) {
	e.entry.Msg = msg
	e.rec.append(e.entry)
}

func (e *recorderEvent) Msgf(format string, args ...any) {
	e.Msg(fmt.Sprintf(format, args...))
}

// Noop is a logger that discards everything. Use it for tests that exercise
// behavior unrelated to logging.
type Noop struct{}

// NewNoop returns a logger that records nothing.
func NewNoop() Noop { return Noop{} }

func (Noop) Debug() port.LogEvent { return noopEvent{} }
func (Noop) Info() port.LogEvent  { return noopEvent{} }
func (Noop) Warn() port.LogEvent  { return noopEvent{} }
func (Noop) Error() port.LogEvent { return noopEvent{} }
func (Noop) Fatal() port.LogEvent { return noopEvent{} }

type noopEvent struct{}

func (e noopEvent) Err(error) port.LogEvent                 { return e }
func (e noopEvent) Str(string, string) port.LogEvent        { return e }
func (e noopEvent) Int(string, int) port.LogEvent           { return e }
func (e noopEvent) Int32(string, int32) port.LogEvent       { return e }
func (e noopEvent) Int64(string, int64) port.LogEvent       { return e }
func (e noopEvent) Bool(string, bool) port.LogEvent         { return e }
func (e noopEvent) Dur(string, time.Duration) port.LogEvent { return e }
func (e noopEvent) Float64(string, float64) port.LogEvent   { return e }
func (e noopEvent) Interface(string, any) port.LogEvent     { return e }
func (e noopEvent) Msg(string)                              {}
func (e noopEvent) Msgf(string, ...any)                     {}

var (
	_ port.Logger   = (*Recorder)(nil)
	_ port.Logger   = Noop{}
	_ port.LogEvent = (*recorderEvent)(nil)
	_ port.LogEvent = noopEvent{}
)
