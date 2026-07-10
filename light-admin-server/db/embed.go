// Package db exposes the embedded SQL migrations and a golang-migrate runner.
// Migrations are the single source of truth for each engine's schema and are
// also consumed by sqlc (see sqlc.yaml).
package db

import "embed"

// MigrationsFS embeds every migration file so the binary can apply schema
// changes without shipping the .sql files alongside it. Migrations are split by
// engine (migrations/postgres, migrations/sqlite, migrations/mysql); the runner
// selects the per-engine subtree (see migrate.go).
//
//go:embed migrations/postgres/*.sql migrations/sqlite/*.sql
var MigrationsFS embed.FS
