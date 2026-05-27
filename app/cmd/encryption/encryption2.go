package encryption

import (
	encryptionGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/primary/grpc"
	loggerAdapter "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/secondary/logger"
	memoryKeygen "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/secondary/memory"
	viperConfig "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/secondary/viper"
	encryptionDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/go-kit/kit/transport/amqp"
)

type Config struct {
	config.PassConfig `mapstructure:",squash"`
	Channel           *amqp.Channel
}

func (conf Config) startServer() {
	// Use hexagonal architecture
	conf.startHexagonalServer()
}

func (conf Config) startHexagonalServer() {
	// Wire secondary adapters
	logger := loggerAdapter.NewAdapter()
	cfg := viperConfig.NewViperConfigAdapter()

	if err := cfg.ValidateListenAddress(); err != nil {
		logger.Fatal().Err(err).Msg("Invalid encryption listen address")
	}
	address := cfg.GetListenAddress()

	// Create key generator (secondary adapter)
	keyGenerator := memoryKeygen.NewKeyGenerator(logger)

	// Create encryption service (domain)
	encryptionService := encryptionDomain.NewEncryptionService(keyGenerator, logger)

	// Create gRPC server (primary adapter)
	grpcServer := encryptionGRPC.NewGRPCServer(encryptionService, address, logger)

	// Start the server
	logger.Info().Str("address", address).Msg("Starting encryption service with hexagonal architecture")
	if err := grpcServer.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start hexagonal encryption server")
	}
}
