package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// IMigrator defines the interface for database migrations.
type IMigrator interface {
	Up() error
	Down() error
	Version() (uint, bool, error)
	Create(name string) error
	Force(version int) error
	Ping(ctx context.Context) error
	Close() error
}

// Migrator handles database migrations.
type Migrator struct {
	db            *sql.DB
	migrationsDir string
	driverName    string
	mig           *migrate.Migrate
	mu            sync.Mutex
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(db *sql.DB, driverName string, migrationsDir string) (*Migrator, error) {
	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
		driverName:    driverName,
	}, nil
}

// Close closes the migrate instance.
func (m *Migrator) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mig != nil {
		sourceErr, dbErr := m.mig.Close()
		m.mig = nil
		if sourceErr != nil {
			return sourceErr
		}
		return dbErr
	}
	return nil
}

// Ping checks the database connection.
func (m *Migrator) Ping(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

// Create creates a new migration file.
func (m *Migrator) Create(name string) error {
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

// Up applies all pending migrations.
func (m *Migrator) Up() error {
	mig, err := m.getMigrateInstance()
	if err != nil {
		return err
	}

	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	return nil
}

// Down rolls back the last migration.
func (m *Migrator) Down() error {
	mig, err := m.getMigrateInstance()
	if err != nil {
		return err
	}

	if err := mig.Steps(-1); err != nil {
		return fmt.Errorf("could not run down migration: %w", err)
	}

	return nil
}

// Version returns the current migration version and whether it's dirty.
func (m *Migrator) Version() (uint, bool, error) {
	mig, err := m.getMigrateInstance()
	if err != nil {
		return 0, false, err
	}

	version, dirty, err := mig.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("could not get migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}

	return version, dirty, nil
}

// Force sets the migration version.
func (m *Migrator) Force(version int) error {
	mig, err := m.getMigrateInstance()
	if err != nil {
		return err
	}

	if err := mig.Force(version); err != nil {
		return fmt.Errorf("could not force migration version: %w", err)
	}

	return nil
}

func (m *Migrator) getMigrateInstance() (*migrate.Migrate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mig != nil {
		return m.mig, nil
	}

	var driver database.Driver
	var err error

	switch m.driverName {
	case "mysql":
		driver, err = mysql.WithInstance(m.db, &mysql.Config{})
	case "sqlite3":
		driver, err = sqlite3.WithInstance(m.db, &sqlite3.Config{})
	default:
		return nil, fmt.Errorf("unsupported database driver name: %s", m.driverName)
	}

	if err != nil {
		return nil, fmt.Errorf("could not create database driver: %w", err)
	}

	mig, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", m.migrationsDir),
		m.driverName,
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create migrate instance: %w", err)
	}

	m.mig = mig
	return mig, nil
}

// Up applies all pending migrations using package-level wrapper.
func Up(db *sql.DB, driverName string, migrationsDir string) error {
	m, err := NewMigrator(db, driverName, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()
	return m.Up()
}

// Down rolls back the last migration using package-level wrapper.
func Down(db *sql.DB, driverName string, migrationsDir string) error {
	m, err := NewMigrator(db, driverName, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()
	return m.Down()
}

// Version returns the current migration version and whether it's dirty using package-level wrapper.
func Version(db *sql.DB, driverName string, migrationsDir string) (uint, bool, error) {
	m, err := NewMigrator(db, driverName, migrationsDir)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()
	return m.Version()
}

// Force sets the migration version using package-level wrapper.
func Force(db *sql.DB, driverName string, migrationsDir string, version int) error {
	m, err := NewMigrator(db, driverName, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()
	return m.Force(version)
}
