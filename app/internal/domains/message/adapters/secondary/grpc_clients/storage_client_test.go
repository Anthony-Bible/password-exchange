package grpc_clients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageClient_CheckDatabase(t *testing.T) {
	// Green Phase: Check if method exists and handles nil client
	c := &StorageClient{}
	status, err := c.CheckDatabase(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "unhealthy", status)
}
