package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotificationPublisher_CheckEmail(t *testing.T) {
	// Green Phase: Check if method exists and handles nil connection
	p := &NotificationPublisher{}
	status, err := p.CheckEmail(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "unhealthy", status)
}
