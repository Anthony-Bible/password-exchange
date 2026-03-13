package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/migrations"
	_ "github.com/go-sql-driver/mysql"
)

// Migrator defines the interface for database migrations
type Migrator interface {
	Up() error
	Down() error
	Version() (uint, bool, error)
	Create(name string) error
}

// defaultMigrator implements the Migrator interface
type defaultMigrator struct {
	db            *sql.DB
	migrationsDir string
}

// NewMigrator creates a new instance of the default Migrator implementation
func NewMigrator(db *sql.DB, migrationsDir string) (Migrator, error) {
	return &defaultMigrator{
		db:            db,
		migrationsDir: migrationsDir,
	}, nil
}

func (m *defaultMigrator) Up() error {
	return migrations.Up(m.db, m.migrationsDir)
}

func (m *defaultMigrator) Down() error {
	return migrations.Down(m.db, m.migrationsDir)
}

func (m *defaultMigrator) Version() (uint, bool, error) {
	return migrations.Version(m.db, m.migrationsDir)
}

func (m *defaultMigrator) Create(name string) error {
	timestamp := time.Now().Format("20060102150405")
	upFileName := fmt.Sprintf("%s_%s.up.sql", timestamp, name)
	downFileName := fmt.Sprintf("%s_%s.down.sql", timestamp, name)

	if err := os.MkdirAll(m.migrationsDir, 0o755); err != nil {
		return fmt.Errorf("could not create migrations directory: %w", err)
	}

	upFile, err := os.Create(filepath.Join(m.migrationsDir, upFileName))
	if err != nil {
		return fmt.Errorf("could not create up migration file: %w", err)
	}
	defer upFile.Close()

	downFile, err := os.Create(filepath.Join(m.migrationsDir, downFileName))
	if err != nil {
		return fmt.Errorf("could not create down migration file: %w", err)
	}
	defer downFile.Close()

	fmt.Printf("Created migration files:\n  %s\n  %s\n", upFileName, downFileName)
	return nil
}
