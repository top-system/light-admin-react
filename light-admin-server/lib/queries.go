package lib

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/top-system/light-admin/db/sqlc"
)

// NewQueries provides the pool-bound sqlc.Queries used by repositories for all
// non-transactional access. Transactional work goes through TxManager.RunInTx,
// which hands the repository a transaction-bound *sqlc.Queries instead.
func NewQueries(p PgxPool) *sqlc.Queries {
	return sqlc.New(p.Pool)
}

// TxManager runs units of work inside a single PostgreSQL transaction using pgx.
// It replaces the previous GORM `WithTrx(*gorm.DB)` pattern: services compose
// multiple sqlc-backed repositories inside one RunInTx callback, and every
// repository call in that callback shares the same transaction.
type TxManager struct {
	pool   *pgxpool.Pool
	logger Logger
}

// NewTxManager creates a TxManager bound to the pgx pool.
func NewTxManager(p PgxPool, logger Logger) TxManager {
	return TxManager{pool: p.Pool, logger: logger}
}

// RunInTx begins a transaction, invokes fn with a transaction-bound Queries, and
// commits on success. Any error (or panic) rolls the transaction back. The
// transaction-bound Queries must be passed to each repository via its WithTx
// method so all statements execute on the same connection.
func (m TxManager) RunInTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
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

	if err := fn(sqlc.New(tx)); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			m.logger.Zap.Errorf("tx rollback failed: %v (original error: %v)", rbErr, err)
		}
		return err
	}

	return tx.Commit(ctx)
}
