package buildinfo_test

import (
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/buildinfo"
)

func TestVersionDefaultIsDev(t *testing.T) {
	t.Parallel()
	if buildinfo.Version != "dev" {
		t.Errorf("default Version = %q, want %q", buildinfo.Version, "dev")
	}
}
