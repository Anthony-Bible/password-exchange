package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/shared/config"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/dsn"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/database/migrations"
	appmigrations "github.com/Anthony-Bible/password-exchange/app/migrations"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Migrator defines the interface for database migrations.
type Migrator interface {
	Up() error
	Down() error
	Version() (uint, bool, error)
	Create(name string) error
	Force(version int) error
}

// migrationTarget describes how to connect to the configured database and
// where its migrations live.
type migrationTarget struct {
	sqlDriver string // database/sql driver name passed to sql.Open
	dsn       string
	dialect   migrations.Dialect
	source    fs.FS  // migrations applied by Up/Down
	createDir string // directory `migrate create` writes new files to
}

// resolveMigrationTarget picks the connection settings and migration source for
// c.DbDriver. An empty DbDriver means mysql.
func resolveMigrationTarget(c config.PassConfig) (migrationTarget, error) {
	switch driver := strings.ToLower(strings.TrimSpace(c.DbDriver)); driver {
	case "", string(migrations.MySQL):
		return migrationTarget{
			sqlDriver: "mysql",
			dsn: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
				c.DbUser, c.DbPass, c.DbHost, c.DbPort, c.DbName),
			dialect:   migrations.MySQL,
			source:    os.DirFS("migrations"),
			createDir: "migrations",
		}, nil
	case string(migrations.Postgres):
		source, err := fs.Sub(appmigrations.PostgresFS, "postgres")
		if err != nil {
			return migrationTarget{}, fmt.Errorf("could not open embedded postgres migrations: %w", err)
		}
		return migrationTarget{
			sqlDriver: "pgx",
			dsn: dsn.Postgres(dsn.Config{
				Host:     c.DbHost,
				Port:     c.DbPort,
				User:     c.DbUser,
				Password: c.DbPass,
				Name:     c.DbName,
			}),
			dialect:   migrations.Postgres,
			source:    source,
			createDir: filepath.Join("migrations", "postgres"),
		}, nil
	default:
		return migrationTarget{}, fmt.Errorf("unsupported dbdriver %q (want mysql or postgres)", c.DbDriver)
	}
}

// defaultMigrator implements the Migrator interface.
type defaultMigrator struct {
	db        *sql.DB
	dialect   migrations.Dialect
	source    fs.FS
	createDir string
}

// NewMigrator creates a new instance of the default Migrator implementation.
// Migrations are read from source; Create writes new files to createDir.
func NewMigrator(db *sql.DB, dialect migrations.Dialect, source fs.FS, createDir string) (Migrator, error) {
	return &defaultMigrator{
		db:        db,
		dialect:   dialect,
		source:    source,
		createDir: createDir,
	}, nil
}

func (m *defaultMigrator) Up() error {
	return migrations.Up(m.db, m.dialect, m.source)
}

func (m *defaultMigrator) Down() error {
	return migrations.Down(m.db, m.dialect, m.source)
}

func (m *defaultMigrator) Version() (uint, bool, error) {
	return migrations.Version(m.db, m.dialect, m.source)
}

func (m *defaultMigrator) Force(version int) error {
	return migrations.Force(m.db, m.dialect, m.source, version)
}

func (m *defaultMigrator) Create(name string) error {
	timestamp := time.Now().Format("20060102150405")
	upFileName := fmt.Sprintf("%s_%s.up.sql", timestamp, name)
	downFileName := fmt.Sprintf("%s_%s.down.sql", timestamp, name)

	if err := os.MkdirAll(m.createDir, 0o755); err != nil {
		return fmt.Errorf("could not create migrations directory: %w", err)
	}

	upFile, err := os.Create(filepath.Join(m.createDir, upFileName))
	if err != nil {
		return fmt.Errorf("could not create up migration file: %w", err)
	}
	defer upFile.Close()

	downFile, err := os.Create(filepath.Join(m.createDir, downFileName))
	if err != nil {
		return fmt.Errorf("could not create down migration file: %w", err)
	}
	defer downFile.Close()

	fmt.Printf("Created migration files:\n  %s\n  %s\n", upFileName, downFileName)
	return nil
}
