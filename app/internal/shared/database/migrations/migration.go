package migrations

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrator handles database migrations.
type Migrator struct {
	db            *sql.DB
	migrationsDir string
	driverName    string
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(db *sql.DB, migrationsDir string) (*Migrator, error) {
	driverType := fmt.Sprintf("%T", db.Driver())
	driverName := ""

	switch driverType {
	case "*sqlite3.SQLiteDriver":
		driverName = "sqlite3"
	case "*mysql.MySQLDriver":
		driverName = "mysql"
	default:
		// Attempt to support the two main drivers even if not perfectly matched
		// but for now we follow the requirement for automatic detection.
		return nil, fmt.Errorf("unsupported database driver type: %s", driverType)
	}

	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
		driverName:    driverName,
	}, nil
}

// Up applies all pending migrations.
func (m *Migrator) Up() error {
	mig, err := m.getMigrateInstance()
	if err != nil {
		return err
	}
	defer mig.Close()

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
	defer mig.Close()

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
	defer mig.Close()

	version, dirty, err := mig.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("could not get migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}

	return version, dirty, nil
}

func (m *Migrator) getMigrateInstance() (*migrate.Migrate, error) {
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

	return mig, nil
}

// Up applies all pending migrations using package-level wrapper.
func Up(db *sql.DB, migrationsDir string) error {
	m, err := NewMigrator(db, migrationsDir)
	if err != nil {
		return err
	}
	return m.Up()
}

// Down rolls back the last migration using package-level wrapper.
func Down(db *sql.DB, migrationsDir string) error {
	m, err := NewMigrator(db, migrationsDir)
	if err != nil {
		return err
	}
	return m.Down()
}

// Version returns the current migration version and whether it's dirty using package-level wrapper.
func Version(db *sql.DB, migrationsDir string) (uint, bool, error) {
	m, err := NewMigrator(db, migrationsDir)
	if err != nil {
		return 0, false, err
	}
	return m.Version()
}
