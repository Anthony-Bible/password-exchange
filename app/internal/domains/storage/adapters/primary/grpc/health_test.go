package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	database "github.com/Anthony-Bible/password-exchange/app/pkg/pb/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
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

// TestUpdateHealthStatus_FlipsToNotServingWhenDomainCheckFails verifies that
// when StorageService.HealthCheck returns an error, the standard health
// service reports NOT_SERVING. This is what makes the web service's /readyz
// (which calls Check via grpc_health_v1) catch real DB outages instead of
// just network/process failures.
func TestUpdateHealthStatus_FlipsToNotServingWhenDomainCheckFails(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{healthErr: errors.New("db down")}
	server := NewGRPCServer(svc, "", &recordingLogger{}, &stubValidator{})

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	server.updateHealthStatus(context.Background(), healthSrv, time.Second)

	resp, err := healthSrv.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: ""})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, resp.GetStatus())
}

// TestUpdateHealthStatus_FlipsBackToServingWhenDomainCheckRecovers verifies
// that after a transient failure the next successful check restores SERVING
// status, so a recovered DB un-trips readiness instead of staying degraded.
func TestUpdateHealthStatus_FlipsBackToServingWhenDomainCheckRecovers(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{healthErr: errors.New("db down")}
	server := NewGRPCServer(svc, "", &recordingLogger{}, &stubValidator{})

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	server.updateHealthStatus(context.Background(), healthSrv, time.Second)
	resp, err := healthSrv.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: ""})
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, resp.GetStatus())

	svc.healthErr = nil
	server.updateHealthStatus(context.Background(), healthSrv, time.Second)
	resp, err = healthSrv.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: ""})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())
}

// TestRunHealthStatusLoop_ExitsOnContextCancel verifies the polling loop
// terminates when the caller cancels the context, preventing a goroutine
// leak at server shutdown.
func TestRunHealthStatusLoop_ExitsOnContextCancel(t *testing.T) {
	t.Parallel()

	svc := &stubStorageService{}
	server := NewGRPCServer(svc, "", &recordingLogger{}, &stubValidator{})
	healthSrv := health.NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		server.runHealthStatusLoop(ctx, healthSrv, 10*time.Millisecond, time.Second)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runHealthStatusLoop did not return after context cancel")
	}
}
