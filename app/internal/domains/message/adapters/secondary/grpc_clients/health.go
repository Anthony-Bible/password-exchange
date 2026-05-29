package grpc_clients

import (
	"context"
	"fmt"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// checkServing queries a gRPC health client and returns nil only when the
// overall service reports SERVING. Errors are wrapped with serviceName so the
// caller doesn't have to.
func checkServing(ctx context.Context, healthClient grpc_health_v1.HealthClient, serviceName string) error {
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return fmt.Errorf("%s health check: %w", serviceName, err)
	}
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("%s health check returned status %s", serviceName, resp.GetStatus().String())
	}
	return nil
}
