package database

import (
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	storageGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/primary/grpc"
	storageMySQL "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	"github.com/rs/zerolog/log"
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
	dbConfig := storageDomain.DatabaseConfig{
		Host:     conf.PassConfig.DbHost,
		User:     conf.PassConfig.DbUser,
		Password: conf.PassConfig.DbPass,
		Name:     conf.PassConfig.DbName,
	}
	
	// Create MySQL adapter (secondary adapter)
	mysqlAdapter := storageMySQL.NewMySQLAdapter(dbConfig)
	
	// Create storage service (domain)
	storageService := storageDomain.NewStorageService(mysqlAdapter)
	
	// Create gRPC server (primary adapter)
	grpcServer := storageGRPC.NewGRPCServer(storageService, address)
	
	// Start the server
	log.Info().Str("address", address).Msg("Starting storage service with hexagonal architecture")
	if err := grpcServer.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start hexagonal storage server")
	}
}

