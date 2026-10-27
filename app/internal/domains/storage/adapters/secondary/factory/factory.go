package factory

import (
	"fmt"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/mysql"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/adapters/secondary/postgres"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/contracts"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/storage/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
)

// NewMessageRepository builds the storage adapter selected by driver
// (PASSWORDEXCHANGE_DBDRIVER). An empty driver means mysql.
func NewMessageRepository(
	driver string,
	cfg contracts.DatabaseConfig,
	logger secondary.LoggerPort,
	validator secondary.ValidationPort,
) (secondary.MessageRepository, error) {
	switch driver {
	case "", "mysql":
		return mysql.NewMySQLAdapter(cfg, logger, validator), nil
	case "postgres":
		return postgres.NewPostgresAdapter(cfg, logger, validator), nil
	default:
		return nil, fmt.Errorf("unsupported PASSWORDEXCHANGE_DBDRIVER %q (expected \"mysql\" or \"postgres\")", driver)
	}
}

// DatabaseConfigFromPass copies the connection settings out of PassConfig.
func DatabaseConfigFromPass(pc config.PassConfig) contracts.DatabaseConfig {
	return contracts.DatabaseConfig{
		Host:         pc.DbHost,
		Port:         pc.DbPort,
		User:         pc.DbUser,
		Password:     pc.DbPass,
		Name:         pc.DbName,
		MaxOpenConns: pc.DbMaxOpenConns,
	}
}
