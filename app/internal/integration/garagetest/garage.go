//go:build integration

// Package garagetest provides a shared helper for integration tests that need
// a real S3-compatible object store. It spins up a Garage (garagehq.deuxfleurs.fr)
// container via testcontainers and provisions a bucket + key pair ready for use.
// Compiled only under the `integration` build tag so the default `go test ./...`
// run stays fast and Docker-free.
package garagetest

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// defaultGarageImage is from docker.io, NOT mirror.gcr.io, because the
	// dxflrs namespace is not mirrored to mirror.gcr.io. Override with
	// PASSWORDEXCHANGE_TEST_GARAGE_IMAGE when a different registry or version is needed.
	defaultGarageImage = "docker.io/dxflrs/garage:v1.0.1"
	imageEnvVar        = "PASSWORDEXCHANGE_TEST_GARAGE_IMAGE"

	s3Port = "3900/tcp"

	garageToml = `
metadata_dir = "/tmp/garage/meta"
data_dir = "/tmp/garage/data"
db_engine = "sqlite"
replication_factor = 1

[rpc_bind_addr]
ip = "0.0.0.0"
port = 3901

[s3_api]
s3_region = "us-east-1"
api_bind_addr = "0.0.0.0:3900"
root_domain = ".s3.garage.localhost"

[s3_web]
bind_addr = "0.0.0.0:3902"
root_domain = ".web.garage.localhost"
index = "index.html"
error_document = "error/@@CODE@@.html"

[admin]
api_bind_addr = "0.0.0.0:3903"
`
)

func garageImage() string {
	if img := os.Getenv(imageEnvVar); img != "" {
		return img
	}
	return defaultGarageImage
}

// StartGarage launches a throwaway Garage container, provisions a bucket named
// "test-bucket" and an access key, and returns the S3 endpoint plus credentials.
// All resources are torn down automatically via t.Cleanup. Intended for use when
// each test needs its own isolated container; for shared-container TestMain use,
// call MustStart instead.
func StartGarage(t *testing.T) (endpoint, accessKey, secretKey, bucket string) {
	t.Helper()

	if _, ok := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); !ok {
		t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	tomlPath := t.TempDir() + "/garage.toml"
	require.NoError(t, os.WriteFile(tomlPath, []byte(garageToml), 0o644))

	endpoint, accessKey, secretKey, bucket, container, err := startContainer(context.Background(), tomlPath)
	require.NoError(t, err, "start garage container")

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate garage container: %v", err)
		}
	})
	return endpoint, accessKey, secretKey, bucket
}

// MustStart launches a Garage container and provisions it, returning the S3
// endpoint, credentials, bucket name, and a teardown func. It is intended for
// use in TestMain where *testing.T is unavailable. On any setup error it calls
// log.Fatal. The caller must call teardown() (not deferred — os.Exit in TestMain
// does not run defers):
//
//	endpoint, key, secret, bucket, teardown := garagetest.MustStart()
//	code := m.Run()
//	teardown()
//	os.Exit(code)
func MustStart() (endpoint, accessKey, secretKey, bucket string, teardown func()) {
	if _, ok := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED"); !ok {
		if err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"); err != nil {
			log.Fatalf("garagetest: set TESTCONTAINERS_RYUK_DISABLED: %v", err)
		}
	}

	tmp, err := os.MkdirTemp("", "garagetest-*")
	if err != nil {
		log.Fatalf("garagetest: create temp dir: %v", err)
	}
	tomlPath := tmp + "/garage.toml"
	if err := os.WriteFile(tomlPath, []byte(garageToml), 0o644); err != nil {
		log.Fatalf("garagetest: write garage.toml: %v", err)
	}

	endpoint, accessKey, secretKey, bucket, container, err := startContainer(context.Background(), tomlPath)
	if err != nil {
		log.Fatalf("garagetest: %v", err)
	}

	teardown = func() {
		_ = os.RemoveAll(tmp)
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("garagetest: terminate container: %v", err)
		}
	}
	return endpoint, accessKey, secretKey, bucket, teardown
}

// startContainer starts and fully provisions a Garage container, returning the
// endpoint, credentials, bucket, the container handle (for teardown), and any error.
func startContainer(ctx context.Context, tomlPath string) (endpoint, accessKey, secretKey, bucket string, container testcontainers.Container, err error) {
	req := testcontainers.ContainerRequest{
		Image:        garageImage(),
		ExposedPorts: []string{s3Port},
		Cmd:          []string{"server"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      tomlPath,
				ContainerFilePath: "/etc/garage.toml",
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForListeningPort(s3Port).WithStartupTimeout(60 * time.Second),
	}

	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", "", "", "", nil, fmt.Errorf("start garage container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return "", "", "", "", container, fmt.Errorf("get container host: %w", err)
	}
	mappedPort, err := container.MappedPort(ctx, s3Port)
	if err != nil {
		return "", "", "", "", container, fmt.Errorf("get mapped port: %w", err)
	}
	endpoint = net.JoinHostPort(host, mappedPort.Port())

	// Give the node a moment to register itself after the TCP port is open.
	time.Sleep(2 * time.Second)

	nodeID, err := getNodeID(ctx, container)
	if err != nil {
		return "", "", "", "", container, err
	}

	if err := assignLayout(ctx, container, nodeID); err != nil {
		return "", "", "", "", container, err
	}

	if _, err := execAndCapture(ctx, container, []string{"garage", "layout", "apply", "--version", "1"}); err != nil {
		return "", "", "", "", container, err
	}
	if _, err := execAndCapture(ctx, container, []string{"garage", "bucket", "create", "test-bucket"}); err != nil {
		return "", "", "", "", container, err
	}

	keyOut, err := execAndCapture(ctx, container, []string{"garage", "key", "create", "test-key"})
	if err != nil {
		return "", "", "", "", container, err
	}
	accessKey, secretKey, err = parseKeyCredentials(keyOut)
	if err != nil {
		return "", "", "", "", container, err
	}

	if _, err := execAndCapture(ctx, container, []string{
		"garage", "bucket", "allow",
		"--read", "--write", "--owner",
		"test-bucket",
		"--key", "test-key",
	}); err != nil {
		return "", "", "", "", container, err
	}

	return endpoint, accessKey, secretKey, "test-bucket", container, nil
}

// getNodeID retrieves the Garage node ID via `garage node id -q`.
func getNodeID(ctx context.Context, container testcontainers.Container) (string, error) {
	out, err := execAndCapture(ctx, container, []string{"garage", "node", "id", "-q"})
	if err != nil {
		return "", fmt.Errorf("get node id: %w", err)
	}
	id := strings.TrimSpace(out)
	if id == "" {
		return "", fmt.Errorf("garage node id returned empty output")
	}
	// Output may be "<hexid>@<host>:<port>" — take only the hex ID part.
	if idx := strings.Index(id, "@"); idx >= 0 {
		id = id[:idx]
	}
	return id, nil
}

// assignLayout assigns the node to a layout zone with a retry loop to handle
// the brief window where the node has not yet registered with itself.
func assignLayout(ctx context.Context, container testcontainers.Container, nodeID string) error {
	const maxAttempts = 10
	for i := range maxAttempts {
		exitCode, _, err := container.Exec(ctx, []string{
			"garage", "layout", "assign",
			"-z", "dc1",
			"-c", "1G",
			nodeID,
		}, tcexec.Multiplexed())
		if err == nil && exitCode == 0 {
			return nil
		}
		if i < maxAttempts-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return fmt.Errorf("garage layout assign did not succeed after %d attempts", maxAttempts)
}

// execAndCapture runs a command in the container and returns stdout on exit 0,
// or an error if the command fails or exits non-zero.
func execAndCapture(ctx context.Context, container testcontainers.Container, cmd []string) (string, error) {
	exitCode, reader, err := container.Exec(ctx, cmd, tcexec.Multiplexed())
	if err != nil {
		return "", fmt.Errorf("exec %v: %w", cmd, err)
	}

	var out string
	if reader != nil {
		data, readErr := io.ReadAll(reader)
		if readErr != nil {
			return "", fmt.Errorf("read output of %v: %w", cmd, readErr)
		}
		out = strings.TrimSpace(string(data))
	}

	if exitCode != 0 {
		return "", fmt.Errorf("command %v exited %d\noutput: %s", cmd, exitCode, out)
	}
	return out, nil
}

// parseKeyCredentials extracts Key ID and Secret key from `garage key create` output.
// Output lines look like:
//
//	Key ID:     GKxxxx
//	Secret key: xxxxxx
func parseKeyCredentials(output string) (keyID, secretKey string, err error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "Key ID:"); ok {
			keyID = strings.TrimSpace(after)
		}
		if after, ok := strings.CutPrefix(line, "Secret key:"); ok {
			secretKey = strings.TrimSpace(after)
		}
	}
	if keyID == "" || secretKey == "" {
		return "", "", fmt.Errorf("could not parse credentials from output:\n%s", output)
	}
	return keyID, secretKey, nil
}
