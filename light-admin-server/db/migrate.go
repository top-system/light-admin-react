package db

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the "pgx5" database driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// newMigrator builds a golang-migrate instance from the embedded migrations and
// a pgx5:// database URL (see lib.DatabaseConfig.PgxURL).
func newMigrator(pgxURL string) (*migrate.Migrate, error) {
	src, err := iofs.New(MigrationsFS, "migrations")
	if err != nil {
		return nil, err
	}
	return migrate.NewWithSourceInstance("iofs", src, pgxURL)
}

// Up applies all pending migrations. A no-op run (already up to date) is not
// treated as an error.
func Up(pgxURL string) error {
	m, err := newMigrator(pgxURL)
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
func Down(pgxURL string) error {
	m, err := newMigrator(pgxURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
