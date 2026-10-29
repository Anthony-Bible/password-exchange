package reminder

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An unknown PASSWORDEXCHANGE_DBDRIVER must make `app reminder` exit non-zero
// (so the CronJob is marked failed) with a message naming the variable.
func TestBinary_UnknownDBDriverExitsNonZeroNamingTheVariable(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "app")
	out, err := exec.Command("go", "build", "-o", bin, "github.com/Anthony-Bible/password-exchange/app").CombinedOutput()
	require.NoError(t, err, "go build: %s", out)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "reminder")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(),
		"PASSWORDEXCHANGE_DBDRIVER=bogus-db",
		"PASSWORDEXCHANGE_DBHOST=127.0.0.1",
		"PASSWORDEXCHANGE_DBPORT=5432",
		"PASSWORDEXCHANGE_DBUSER=u",
		"PASSWORDEXCHANGE_DBPASS=p",
		"PASSWORDEXCHANGE_DBNAME=d",
		"PASSWORDEXCHANGE_REMINDER_ENABLED=true",
	)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf

	runErr := cmd.Run()

	require.NoError(t, ctx.Err(), "reminder must fail fast; output: %s", buf.String())
	var exitErr *exec.ExitError
	require.True(t, errors.As(runErr, &exitErr), "expected a non-zero exit, got %v; output: %s", runErr, buf.String())
	assert.NotZero(t, exitErr.ExitCode())
	assert.Contains(t, buf.String(), "PASSWORDEXCHANGE_DBDRIVER")
	assert.Contains(t, buf.String(), "bogus-db")
}
