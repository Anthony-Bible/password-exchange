package encryption

import (
	"context" // Add this
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	encryptionDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	encryptionGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/primary/grpc"
	memoryKeygen "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/secondary/memory"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/ports" // Add this
	"github.com/go-kit/kit/transport/amqp"
	// "github.com/rs/zerolog/log" // Remove this
)

type Config struct {
	config.PassConfig `mapstructure:",squash"`
	Channel           *amqp.Channel
}


func (conf Config) startServer(logger ports.Logger) error {
	// Use hexagonal architecture
	return conf.startHexagonalServer(logger)
}

func (conf Config) startHexagonalServer(logger ports.Logger) error {
	address := "0.0.0.0:50051" // Ensure this is the correct/intended address for encryption service
	
	// Create key generator (secondary adapter)
	keyGenerator := memoryKeygen.NewKeyGenerator()
	
	// Create encryption service (domain)
	encryptionService := encryptionDomain.NewEncryptionService(keyGenerator)
	
	// Create gRPC server (primary adapter)
	// This now requires the logger instance
	grpcServer := encryptionGRPC.NewGRPCServer(encryptionService, address, logger)
	
	// Start the server
	logger.Info(context.Background(), "Starting encryption service with hexagonal architecture", "address", address)
	if err := grpcServer.Start(); err != nil {
		logger.Error(context.Background(), "Failed to start hexagonal encryption server", "error", err)
		return err
	}
	return nil
}

