package database

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type runResult struct {
	exitCode int
	timedOut bool
	output   string
}

// buildAppBinary compiles the application entrypoint (package app) into a temp dir.
func buildAppBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "app")
	cmd := exec.Command("go", "build", "-o", bin, "github.com/Anthony-Bible/password-exchange/app")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "go build: %s", out)
	return bin
}

// runBinary runs `<bin> database` from workdir with the given extra env and
// waits up to timeoutSec for it to exit.
func runBinary(t *testing.T, bin, workdir string, timeoutSec int, env map[string]string) runResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "database")
	cmd.Dir = workdir
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf

	err := cmd.Run()
	res := runResult{output: buf.String()}
	if ctx.Err() != nil {
		res.timedOut = true
		return res
	}
	if ee, ok := err.(*exec.ExitError); ok {
		res.exitCode = ee.ExitCode()
	}
	return res
}
