package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the "pgx5" database driver
	_ "github.com/golang-migrate/migrate/v4/database/sqlite" // registers the "sqlite" database driver (modernc)
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrationsDir maps a normalized engine name (lib.DatabaseConfig.EngineName)
// to its embedded migration subtree. The per-engine directories share version
// numbers and file names; only the DDL dialect differs (see
// docs/multi-database-plan.md).
var migrationsDir = map[string]string{
	"postgres": "migrations/postgres",
	"sqlite":   "migrations/sqlite",
	"mysql":    "migrations/mysql",
}

// newMigrator builds a golang-migrate instance from the embedded migrations
// for the given engine and a database URL whose scheme selects the registered
// driver (pgx5://, sqlite://, mysql://) — see lib.DatabaseConfig.MigrateURL.
func newMigrator(engine, databaseURL string) (*migrate.Migrate, error) {
	dir, ok := migrationsDir[engine]
	if !ok {
		return nil, fmt.Errorf("db: unknown engine %q", engine)
	}
	src, err := iofs.New(MigrationsFS, dir)
	if err != nil {
		return nil, err
	}
	return migrate.NewWithSourceInstance("iofs", src, databaseURL)
}

// Up applies all pending migrations. A no-op run (already up to date) is not
// treated as an error.
func Up(engine, databaseURL string) error {
	m, err := newMigrator(engine, databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Down rolls back the most recently applied migration (a single step), which is
// the safe default for `make migrate-down`.
func Down(engine, databaseURL string) error {
	m, err := newMigrator(engine, databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
