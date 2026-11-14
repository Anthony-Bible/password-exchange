package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

// The shared sre-bible instance has ~19 usable connections split between two
// apps (#568), so the pool must stay small and give idle connections back.

func TestPoolSettings_UnsetMaxOpenUsesTheSharedBudgetDefault(t *testing.T) {
	for _, configured := range []int{0, -1} {
		assert.Equal(t, 2, poolSettingsFor(configured).MaxOpen, "configured=%d", configured)
	}
}

func TestPoolSettings_IdleConnectionsAreReleasedWithinTwoMinutes(t *testing.T) {
	assert.Equal(t, 2*time.Minute, poolSettingsFor(2).MaxIdleTime)
}

func TestPoolSettings_Invariants(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		configured := rapid.IntRange(-10, 100).Draw(rt, "configured")
		s := poolSettingsFor(configured)

		if configured > 0 && s.MaxOpen != configured {
			rt.Fatalf("an explicit MaxOpenConns (%d) must be honoured, got %d", configured, s.MaxOpen)
		}
		if s.MaxOpen < 1 {
			rt.Fatalf("pool must allow at least one connection, got %d", s.MaxOpen)
		}
		if s.MaxIdle < 0 || s.MaxIdle > s.MaxOpen {
			rt.Fatalf("idle pool %d must be within [0, %d]", s.MaxIdle, s.MaxOpen)
		}
		if s.MaxIdleTime <= 0 || s.MaxIdleTime > s.MaxLifetime {
			rt.Fatalf("idle timeout %v must be positive and no longer than lifetime %v", s.MaxIdleTime, s.MaxLifetime)
		}
	})
}
