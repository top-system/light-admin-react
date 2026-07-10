package lib

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver (pure Go, no CGO)

	"github.com/top-system/light-admin/db/pgstore"
	"github.com/top-system/light-admin/db/sqlitestore"
	"github.com/top-system/light-admin/db/store"
)

// TxManager is the engine-neutral transaction manager consumed by services.
// It is an alias of store.TxManager so business code keeps its lib.TxManager
// dependency while the concrete implementation is selected per engine.
type TxManager = store.TxManager

// DataLayer bundles the engine-neutral data-access facades over one concrete
// database handle. Which engine backs them is decided once, at open time, from
// config.Database.Engine (docs/multi-database-plan.md §3.4).
type DataLayer struct {
	Store     store.Store
	TxManager TxManager
	close     func()
}

// Close releases the underlying pool/handle. FX callers get this via the
// lifecycle hook in NewDataLayer; standalone callers (cmd/setup) defer it.
func (d *DataLayer) Close() {
	if d.close != nil {
		d.close()
	}
}

// OpenDataLayer connects to the configured database engine and returns the
// neutral store.Store + TxManager over it. It is the single place that knows
// which adapter (pgstore / sqlitestore / mysqlstore) wraps which handle.
func OpenDataLayer(config Config, logger Logger) (*DataLayer, error) {
	switch config.Database.EngineName() {
	case "postgres":
		pool, err := newPgxPool(config, logger)
		if err != nil {
			return nil, err
		}
		return &DataLayer{
			Store:     pgstore.New(pool),
			TxManager: pgxTxManager{pool: pool, logger: logger},
			close:     pool.Close,
		}, nil

	case "sqlite":
		handle, err := openSQLite(config, logger)
		if err != nil {
			return nil, err
		}
		return &DataLayer{
			Store: sqlitestore.New(handle),
			TxManager: sqlTxManager{
				db:       handle,
				newStore: func(tx *sql.Tx) store.Store { return sqlitestore.New(tx) },
				logger:   logger,
			},
			close: func() { _ = handle.Close() },
		}, nil

	default:
		return nil, fmt.Errorf("unsupported Database.Engine %q (postgres, sqlite)", config.Database.Engine)
	}
}

// NewDataLayer is the FX provider: it opens the configured engine's data layer
// and closes it on shutdown. A connection failure is fatal, matching the
// previous pgx-only behaviour.
func NewDataLayer(lc fx.Lifecycle, config Config, logger Logger) (store.Store, TxManager) {
	dl, err := OpenDataLayer(config, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Error opening the data layer: %v", err))
		os.Exit(1)
	}
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			dl.Close()
			return nil
		},
	})
	return dl.Store, dl.TxManager
}

// newPgxPool builds a pgxpool.Pool from the database configuration (the
// PostgreSQL branch of the data layer).
func newPgxPool(config Config, logger Logger) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(config.Database.PgxDSN())
	if err != nil {
		return nil, fmt.Errorf("parsing pgx pool config: %w", err)
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
		return nil, fmt.Errorf("creating pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging PostgreSQL via pgx: %w", err)
	}

	logger.Info(fmt.Sprintf("pgx pool established (max_conns=%d)", poolConfig.MaxConns))
	return pool, nil
}

// openSQLite opens the SQLite database file via modernc.org/sqlite. The DSN
// pins WAL/busy_timeout/foreign_keys pragmas (see DatabaseConfig.SQLiteDSN);
// the pool is capped at a single connection to serialize writers instead of
// surfacing SQLITE_BUSY under concurrency (docs/multi-database-plan.md §2.3).
func openSQLite(config Config, logger Logger) (*sql.DB, error) {
	path := config.Database.SQLitePath()
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating sqlite directory %s: %w", dir, err)
		}
	}

	handle, err := sql.Open("sqlite", config.Database.SQLiteDSN())
	if err != nil {
		return nil, fmt.Errorf("opening sqlite %s: %w", path, err)
	}
	handle.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := handle.PingContext(ctx); err != nil {
		_ = handle.Close()
		return nil, fmt.Errorf("pinging sqlite %s: %w", path, err)
	}

	logger.Info(fmt.Sprintf("sqlite database opened (%s)", path))
	return handle, nil
}

// pgxTxManager implements TxManager over a pgx connection pool.
type pgxTxManager struct {
	pool   *pgxpool.Pool
	logger Logger
}

var _ store.TxManager = pgxTxManager{}

// RunInTx begins a transaction, invokes fn with a transaction-bound Store, and
// commits on success. Any error (or panic) rolls the transaction back.
func (m pgxTxManager) RunInTx(ctx context.Context, fn func(store.Store) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(pgstore.New(tx)); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			m.logger.Error(fmt.Sprintf("tx rollback failed: %v (original error: %v)", rbErr, err))
		}
		return err
	}

	return tx.Commit(ctx)
}

// sqlTxManager implements TxManager over database/sql (sqlite, mysql). The
// newStore hook binds the engine's adapter to the transaction so every
// repository call inside the callback shares it.
type sqlTxManager struct {
	db       *sql.DB
	newStore func(*sql.Tx) store.Store
	logger   Logger
}

var _ store.TxManager = sqlTxManager{}

// RunInTx begins a transaction, invokes fn with a transaction-bound Store, and
// commits on success. Any error (or panic) rolls the transaction back.
func (m sqlTxManager) RunInTx(ctx context.Context, fn func(store.Store) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(m.newStore(tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			m.logger.Error(fmt.Sprintf("tx rollback failed: %v (original error: %v)", rbErr, err))
		}
		return err
	}

	return tx.Commit()
}
