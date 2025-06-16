package database

import (
	"context"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	storageDomain "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/domain"
	storageGRPC "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/primary/grpc"
	storageMySQL "github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging/ports"
	// "github.com/rs/zerolog/log"
)

type Config struct {
	PassConfig config.PassConfig `mapstructure:",squash"`
}


func (conf Config) startServer(logger ports.Logger) error {
	// Use hexagonal architecture
	return conf.startHexagonalServer(logger)
}

func (conf Config) startHexagonalServer(logger ports.Logger) error {
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
	grpcServer := storageGRPC.NewGRPCServer(storageService, address, logger) // New line
	
	// Start the server
	logger.Info(context.Background(), "Starting storage service with hexagonal architecture", "address", address)
	if err := grpcServer.Start(); err != nil {
		logger.Error(context.Background(), "Failed to start hexagonal storage server", "error", err)
		return err
	}
	return nil
}

