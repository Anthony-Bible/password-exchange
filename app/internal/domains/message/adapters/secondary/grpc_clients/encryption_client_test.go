package grpc_clients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionClient_CheckHealth(t *testing.T) {
	// Green Phase: Check if method exists and handles nil client
	c := &EncryptionClient{}
	status, err := c.CheckHealth(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "unhealthy", status)
}
