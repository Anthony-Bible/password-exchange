package encryption

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	encryptionDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/domain"
	encryptionGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/primary/grpc"
	memoryKeygen "github.com/Anthony-Bible/password-exchange/app/internal/domains/encryption/adapters/secondary/memory"
	"github.com/go-kit/kit/transport/amqp"
	"github.com/rs/zerolog/log"
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
	address := "0.0.0.0:50051"
	
	// Create key generator (secondary adapter)
	keyGenerator := memoryKeygen.NewKeyGenerator()
	
	// Create encryption service (domain)
	encryptionService := encryptionDomain.NewEncryptionService(keyGenerator)
	
	// Create gRPC server (primary adapter)
	grpcServer := encryptionGRPC.NewGRPCServer(encryptionService, address)
	
	// Start the server
	log.Info().Str("address", address).Msg("Starting encryption service with hexagonal architecture")
	if err := grpcServer.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start hexagonal encryption server")
	}
}

