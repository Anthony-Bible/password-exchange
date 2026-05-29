package logtest

import "testing"

func TestRecorderDistinguishesLevels(t *testing.T) {
	rec := NewRecorder()

	rec.Info().Str("k", "v").Msg("info one")
	rec.Warn().Int("n", 2).Msg("warn one")
	rec.Error().Err(errBoom).Msg("error one")

	if got := rec.Count("info"); got != 1 {
		t.Errorf("info count = %d, want 1", got)
	}
	if got := rec.Count("warn"); got != 1 {
		t.Errorf("warn count = %d, want 1", got)
	}
	if got := rec.Count("debug"); got != 0 {
		t.Errorf("debug count = %d, want 0", got)
	}

	entries := rec.Entries()
	if len(entries) != 3 {
		t.Fatalf("entries = %d, want 3", len(entries))
	}
	if entries[0].Fields["k"] != "v" || entries[0].Msg != "info one" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[2].Err != errBoom {
		t.Errorf("expected error captured on third entry, got %+v", entries[2])
	}
}

func TestNoopDiscardsEverything(t *testing.T) {
	n := NewNoop()
	// Must not panic and must support full chaining including Msgf.
	n.Debug().Str("k", "v").Int("n", 1).Int32("a", 2).Int64("b", 3).
		Bool("c", true).Float64("d", 1.5).Interface("e", struct{}{}).Msgf("%d", 1)
	n.Fatal().Err(errBoom).Msg("dying")
}

var errBoom = errString("boom")

type errString string

func (e errString) Error() string { return string(e) }
