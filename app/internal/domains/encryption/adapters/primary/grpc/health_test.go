package grpc

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	pb "github.com/Anthony-Bible/password-exchange/app/pkg/pb/encryption"
	"google.golang.org/grpc/test/bufconn"
)

func startHealthTestServer(t *testing.T) (grpc_health_v1.HealthClient, func()) {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	adapter := NewGRPCServer(&stubService{}, "", stubLogger{})
	pb.RegisterMessageServiceServer(srv, adapter)
	adapter.registerHealthServer(srv)

	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- srv.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		_ = conn.Close()
		srv.Stop()
		_ = lis.Close()
		if serveErr := <-serveErrCh; serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			t.Errorf("gRPC Serve returned unexpected error: %v", serveErr)
		}
	}
	return grpc_health_v1.NewHealthClient(conn), cleanup
}

func TestHealthCheck_Serving(t *testing.T) {
	client, cleanup := startHealthTestServer(t)
	t.Cleanup(cleanup)

	resp, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: ""})

	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())
}
