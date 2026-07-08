package lib

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

// PgxPool wraps the pgxpool connection pool used by the sqlc-based repositories.
//
// It targets PostgreSQL exclusively (the project's chosen datastore). During the
// GORM -> sqlc migration this pool coexists with lib.Database (GORM): both point
// at the same PostgreSQL instance, and repositories are moved over one module at
// a time.
type PgxPool struct {
	Pool *pgxpool.Pool
}

// NewPgxPool builds a pgxpool.Pool from the database configuration and registers
// a lifecycle hook to close it on shutdown.
//
// The server now requires PostgreSQL; a non-postgres Engine is a fatal
// misconfiguration because the sqlc repositories cannot run against it.
func NewPgxPool(lc fx.Lifecycle, config Config, logger Logger) PgxPool {
	if !config.Database.IsPostgreSQL() {
		logger.Error(fmt.Sprintf(
			"pgx pool requires Database.Engine=postgres (got %q); the sqlc data layer is PostgreSQL-only",
			config.Database.Engine,
		))
		os.Exit(1)
	}

	poolConfig, err := pgxpool.ParseConfig(config.Database.PgxDSN())
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing pgx pool config: %v", err))
		os.Exit(1)
	}

	if n := config.Database.MaxOpenConns; n > 0 {
		poolConfig.MaxConns = int32(n)
	}
	if n := config.Database.MaxIdleConns; n > 0 {
		poolConfig.MinConns = int32(n)
	}
	if n := config.Database.MaxLifetime; n > 0 {
		poolConfig.MaxConnLifetime = time.Duration(n) * time.Second
	}
	poolConfig.MaxConnIdleTime = 10 * time.Minute
	// Pin the session time zone to match the historical GORM behaviour.
	poolConfig.ConnConfig.RuntimeParams["timezone"] = "Asia/Shanghai"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating pgx pool: %v", err))
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		logger.Error(fmt.Sprintf("Error pinging PostgreSQL via pgx: %v", err))
		os.Exit(1)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			pool.Close()
			return nil
		},
	})

	logger.Info(fmt.Sprintf("pgx pool established (max_conns=%d)", poolConfig.MaxConns))
	return PgxPool{Pool: pool}
}
