/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Anthony-Bible/password-exchange/app/cmd"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/migrations"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	migrator migrations.IMigrator
	cfg      Config
	// runDatabaseServer is a variable to allow mocking in tests.
	runDatabaseServer = func() {
		log.Info().Msg("Starting database server...")
		cfg.startServer()
	}
)

// databaseCmd represents the database command.
var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Database management and server",
	Long:  `Manage database migrations and start the database server.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Only initialize if we are not in a test where migrator might be mocked
		if migrator == nil {
			if err := initConfigAndMigrator(); err != nil {
				return err
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAutoMigrations(); err != nil {
			log.Error().Err(err).Msg("Failed to run auto-migrations")
			return
		}
		runDatabaseServer()
	},
}

// runAutoMigrations executes pending migrations if AutoMigrate is enabled.
func runAutoMigrations() error {
	if !cfg.PassConfig.AutoMigrate {
		return nil
	}

	if migrator == nil {
		return errors.New("migrator not initialized")
	}

	log.Info().Msg("Running auto-migrations...")
	if err := migrator.Up(); err != nil {
		return fmt.Errorf("error running auto-migrations: %w", err)
	}

	return nil
}

// migrateCmd represents the migrate command.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
	Long:  `Manage database migrations using golang-migrate.`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrator == nil {
			return errors.New("migrator not initialized")
		}
		return migrator.Up()
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the last migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrator == nil {
			return errors.New("migrator not initialized")
		}
		return migrator.Down()
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current migration version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrator == nil {
			return errors.New("migrator not initialized")
		}
		version, dirty, err := migrator.Version()
		if err != nil {
			return err
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)
		return nil
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new migration file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrator == nil {
			return errors.New("migrator not initialized")
		}
		return migrator.Create(args[0])
	},
}

var migrateForceCmd = &cobra.Command{
	Use:   "force",
	Short: "Force the migration version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrator == nil {
			return errors.New("migrator not initialized")
		}
		var version int
		if _, err := fmt.Sscanf(args[0], "%d", &version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
		return migrator.Force(version)
	},
}

func initConfigAndMigrator() error {
	config.BindEnvs(cfg)
	if err := viper.Unmarshal(&cfg.PassConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Initialize migrator
	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		cfg.PassConfig.DbUser,
		cfg.PassConfig.DbPass,
		cfg.PassConfig.DbHost,
		cfg.PassConfig.DbPort,
		cfg.PassConfig.DbName,
	)
	db, err := sql.Open("mysql", connString)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	m, err := migrations.NewMigrator(db, "mysql", "migrations")
	if err != nil {
		return fmt.Errorf("error initializing migrator: %w", err)
	}
	migrator = m
	return nil
}

func init() {
	// Setup command hierarchy
	databaseCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateForceCmd)

	// Register with root command
	cmd.RootCmd.AddCommand(databaseCmd)
}
