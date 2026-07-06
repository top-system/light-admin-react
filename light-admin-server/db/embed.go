// Package db exposes the embedded SQL migrations and a golang-migrate runner.
// Migrations are the single source of truth for the PostgreSQL schema and are
// also consumed by sqlc (see sqlc.yaml).
package db

import "embed"

// MigrationsFS embeds every migration file so the binary can apply schema
// changes without shipping the .sql files alongside it.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
