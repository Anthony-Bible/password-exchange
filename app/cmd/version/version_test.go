package version_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Anthony-Bible/password-exchange/app/cmd/version"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/buildinfo"
)

func TestVersionCommand_PrintsBuildVersion(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	version.Cmd.SetOut(&buf)
	version.Cmd.SetErr(&buf)

	version.Cmd.Run(version.Cmd, nil)

	got := strings.TrimSpace(buf.String())
	if got != buildinfo.Version {
		t.Errorf("version command output = %q, want %q", got, buildinfo.Version)
	}
}
