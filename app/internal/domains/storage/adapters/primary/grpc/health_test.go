package grpc

import (
	"context"
	"errors"
	"net"
	"testing"

	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

// TestRegisterHealthServer_OverallServiceReportsServing verifies that after the
// storage adapter registers the standard gRPC health service, a Check call for
// the overall service ("") returns SERVING. k8s grpc: probes rely on this exact
// contract to mark pods Ready.
func TestRegisterHealthServer_OverallServiceReportsServing(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{}
	logger := &recordingLogger{}
	validator := &stubValidator{}
	adapter := NewGRPCServer(svc, "", logger, validator)

	lis := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	database.RegisterDbServiceServer(grpcServer, adapter)
	adapter.registerHealthServer(grpcServer)

	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- grpcServer.Serve(lis)
	}()
	t.Cleanup(func() {
		grpcServer.Stop()
		_ = lis.Close()
		if err := <-serveErrCh; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("gRPC Serve returned unexpected error: %v", err)
		}
	})

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: ""})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())
}
