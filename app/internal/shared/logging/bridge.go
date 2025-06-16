package logging

import (
	"context"
	"fmt"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/ports"
)

// BridgeLogger is a logger that delegates to an underlying ports.Logger implementation (either SlogAdapter or ZerologAdapter).
type BridgeLogger struct {
	delegate ports.Logger
}

// NewBridgeLogger creates a new BridgeLogger that delegates to either SlogAdapter or ZerologAdapter.
func NewBridgeLogger(cfg LogConfig, useSlog bool, serviceName string) (ports.Logger, error) {
	var adapter ports.Logger
	var err error

	if useSlog {
		adapter, err = NewSlogAdapter(cfg, serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to create SlogAdapter: %w", err)
		}
	} else {
		adapter, err = NewZerologAdapter(cfg, serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to create ZerologAdapter: %w", err)
		}
	}

	return &BridgeLogger{
		delegate: adapter,
	}, nil
}

func (b *BridgeLogger) Info(ctx context.Context, msg string, args ...any) {
	b.delegate.Info(ctx, msg, args...)
}

func (b *BridgeLogger) Error(ctx context.Context, msg string, args ...any) {
	b.delegate.Error(ctx, msg, args...)
}

func (b *BridgeLogger) Debug(ctx context.Context, msg string, args ...any) {
	b.delegate.Debug(ctx, msg, args...)
}

// With returns a new BridgeLogger with the delegate's With applied.
func (b *BridgeLogger) With(key string, value any) ports.Logger {
	return &BridgeLogger{delegate: b.delegate.With(key, value)}
}

// WithContext returns a new BridgeLogger with the delegate's WithContext applied.
func (b *BridgeLogger) WithContext(ctx context.Context) ports.Logger {
	return &BridgeLogger{delegate: b.delegate.WithContext(ctx)}
}
