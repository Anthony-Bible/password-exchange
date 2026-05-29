package database

import (
	storageGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/primary/grpc"
	storageMySQL "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	storageContracts "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
	"github.com/Anthony-Bible/password-exchange/app/pkg/validation"
)

type Config struct {
	PassConfig config.PassConfig `mapstructure:",squash"`
}

func (conf Config) startServer() {
	// Use hexagonal architecture
	conf.startHexagonalServer()
}

func (conf Config) startHexagonalServer() {
	address := "0.0.0.0:50051"

	// Create database configuration from PassConfig
	dbConfig := storageContracts.DatabaseConfig{
		Host:     conf.PassConfig.DbHost,
		User:     conf.PassConfig.DbUser,
		Password: conf.PassConfig.DbPass,
		Name:     conf.PassConfig.DbName,
	}

	// Wire up the secondary port adapters first so they can be passed into
	// every adapter that depends on them (mysql, grpc).
	storageLogger := logging.NewLogger()
	storageValidator := validation.NewAdapter()

	// Create MySQL adapter (secondary adapter)
	mysqlAdapter := storageMySQL.NewMySQLAdapter(dbConfig, storageLogger, storageValidator)
	defer func() {
		if err := mysqlAdapter.Close(); err != nil {
			logging.Error().Err(err).Msg("Failed to close storage adapter")
		}
	}()

	// Create storage service (domain)
	storageService := storageDomain.NewStorageService(mysqlAdapter, storageLogger, storageValidator)

	// Create gRPC server (primary adapter) — inject logger and validator so
	// the adapter does not reach out to the shared logging/validation packages.
	grpcServer := storageGRPC.NewGRPCServer(storageService, address, storageLogger, storageValidator)

	// Start the server
	logging.Info().Str("address", address).Msg("Starting storage service with hexagonal architecture")
	if err := grpcServer.Start(); err != nil {
		logging.Fatal().Err(err).Msg("Failed to start hexagonal storage server")
	}
}
