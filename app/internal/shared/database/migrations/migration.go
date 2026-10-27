package migrations

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	pgmigrations "github.com/Anthony-Bible/password-exchange/app/migrations"
)

// Dialect names the database flavour a migrator targets.
type Dialect string

const (
	DialectMySQL    Dialect = "mysql"
	DialectPostgres Dialect = "postgres"
	DialectSQLite3  Dialect = "sqlite3"
)

// Migrator handles database migrations.
type Migrator struct {
	db            *sql.DB
	migrationsDir string
	dialect       Dialect
}

// NewMigrator creates a new Migrator instance. For postgres the embedded
// migrations are used and migrationsDir is ignored.
func NewMigrator(db *sql.DB, dialect Dialect, migrationsDir string) (*Migrator, error) {
	switch dialect {
	case DialectMySQL, DialectPostgres, DialectSQLite3:
	default:
		return nil, fmt.Errorf("unsupported database dialect: %q", string(dialect))
	}
	return &Migrator{db: db, migrationsDir: migrationsDir, dialect: dialect}, nil
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
	if m.dialect == DialectPostgres {
		src, err := iofs.New(pgmigrations.PostgresFS, "postgres")
		if err != nil {
			return nil, fmt.Errorf("could not open embedded migrations: %w", err)
		}
		drv, err := pgx.WithInstance(m.db, &pgx.Config{})
		if err != nil {
			return nil, fmt.Errorf("could not create database driver: %w", err)
		}
		mig, err := migrate.NewWithInstance("iofs", src, "pgx5", drv)
		if err != nil {
			return nil, fmt.Errorf("could not create migrate instance: %w", err)
		}
		return mig, nil
	}

	var driver database.Driver
	var err error

	switch m.dialect {
	case DialectMySQL:
		driver, err = mysql.WithInstance(m.db, &mysql.Config{})
	case DialectSQLite3:
		driver, err = sqlite3.WithInstance(m.db, &sqlite3.Config{})
	default:
		return nil, fmt.Errorf("unsupported database dialect: %q", string(m.dialect))
	}
	if err != nil {
		return nil, fmt.Errorf("could not create database driver: %w", err)
	}

	mig, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", m.migrationsDir),
		string(m.dialect),
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create migrate instance: %w", err)
	}

	return mig, nil
}

// Up applies all pending migrations using package-level wrapper.
func Up(db *sql.DB, dialect Dialect, migrationsDir string) error {
	m, err := NewMigrator(db, dialect, migrationsDir)
	if err != nil {
		return err
	}
	return m.Up()
}

// Down rolls back the last migration using package-level wrapper.
func Down(db *sql.DB, dialect Dialect, migrationsDir string) error {
	m, err := NewMigrator(db, dialect, migrationsDir)
	if err != nil {
		return err
	}
	return m.Down()
}

// Version returns the current migration version and whether it's dirty using package-level wrapper.
func Version(db *sql.DB, dialect Dialect, migrationsDir string) (uint, bool, error) {
	m, err := NewMigrator(db, dialect, migrationsDir)
	if err != nil {
		return 0, false, err
	}
	return m.Version()
}

// Force sets the migration version using package-level wrapper.
func Force(db *sql.DB, dialect Dialect, migrationsDir string, version int) error {
	m, err := NewMigrator(db, dialect, migrationsDir)
	if err != nil {
		return err
	}
	return m.Force(version)
}
