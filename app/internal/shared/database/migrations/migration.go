package migrations

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	pgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Dialect selects the golang-migrate database driver used to apply migrations.
type Dialect string

// Supported dialects.
const (
	MySQL    Dialect = "mysql"
	Postgres Dialect = "postgres"
	SQLite3  Dialect = "sqlite3"
)

// Migrator handles database migrations.
//
// golang-migrate's Close() closes the *sql.DB it was given, and every method
// builds and closes its own migrate instance, so a Migrator supports exactly
// one Up/Down/Version/Force call per *sql.DB.
type Migrator struct {
	db      *sql.DB
	dialect Dialect
	source  fs.FS
}

// NewMigrator creates a new Migrator that applies the migrations found at the
// root of source to db using the given dialect.
func NewMigrator(db *sql.DB, dialect Dialect, source fs.FS) (*Migrator, error) {
	switch dialect {
	case MySQL, Postgres, SQLite3:
	default:
		return nil, fmt.Errorf("unsupported database dialect: %q", dialect)
	}
	if source == nil {
		return nil, errors.New("migration source must not be nil")
	}

	return &Migrator{
		db:      db,
		dialect: dialect,
		source:  source,
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

// Force sets the migration version.
func (m *Migrator) Force(version int) error {
	mig, err := m.getMigrateInstance()
	if err != nil {
		return err
	}
	defer mig.Close()

	if err := mig.Force(version); err != nil {
		return fmt.Errorf("could not force migration version: %w", err)
	}

	return nil
}

func (m *Migrator) getMigrateInstance() (*migrate.Migrate, error) {
	var driver database.Driver
	var err error

	var driverName string

	// Open the source first so a bad source doesn't leave the driver (and with
	// it the *sql.DB) half-initialised.
	src, err := iofs.New(m.source, ".")
	if err != nil {
		return nil, fmt.Errorf("could not open migration source: %w", err)
	}

	switch m.dialect {
	case MySQL:
		driverName = "mysql"
		driver, err = mysql.WithInstance(m.db, &mysql.Config{})
	case Postgres:
		driverName = "pgx5"
		driver, err = pgx.WithInstance(m.db, &pgx.Config{})
	case SQLite3:
		driverName = "sqlite3"
		driver, err = sqlite3.WithInstance(m.db, &sqlite3.Config{})
	default:
		err = fmt.Errorf("unsupported database dialect: %q", m.dialect)
	}

	if err != nil {
		_ = src.Close()
		return nil, fmt.Errorf("could not create database driver: %w", err)
	}

	mig, err := migrate.NewWithInstance("iofs", src, driverName, driver)
	if err != nil {
		_ = src.Close()
		_ = driver.Close()
		return nil, fmt.Errorf("could not create migrate instance: %w", err)
	}

	return mig, nil
}

// Up applies all pending migrations using package-level wrapper.
func Up(db *sql.DB, dialect Dialect, source fs.FS) error {
	m, err := NewMigrator(db, dialect, source)
	if err != nil {
		return err
	}
	return m.Up()
}

// Down rolls back the last migration using package-level wrapper.
func Down(db *sql.DB, dialect Dialect, source fs.FS) error {
	m, err := NewMigrator(db, dialect, source)
	if err != nil {
		return err
	}
	return m.Down()
}

// Version returns the current migration version and whether it's dirty using package-level wrapper.
func Version(db *sql.DB, dialect Dialect, source fs.FS) (uint, bool, error) {
	m, err := NewMigrator(db, dialect, source)
	if err != nil {
		return 0, false, err
	}
	return m.Version()
}

// Force sets the migration version using package-level wrapper.
func Force(db *sql.DB, dialect Dialect, source fs.FS, version int) error {
	m, err := NewMigrator(db, dialect, source)
	if err != nil {
		return err
	}
	return m.Force(version)
}
