package lib

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/top-system/light-admin/db/pgstore"
	"github.com/top-system/light-admin/db/store"
)

// NewStore provides the pool-bound engine-neutral store.Store used by
// repositories for all non-transactional access. Transactional work goes
// through TxManager.RunInTx, which hands the repository a transaction-bound
// store.Store instead.
//
// The concrete implementation is selected by the configured database engine.
// Today only PostgreSQL (pgstore over pgx/v5) is wired; additional engines plug
// in here without touching any business code, which depends only on store.Store.
func NewStore(p PgxPool) store.Store {
	return pgstore.New(p.Pool)
}

// TxManager runs units of work inside a single PostgreSQL transaction using pgx.
// It implements store.TxManager: services compose multiple store-backed
// repositories inside one RunInTx callback, and every repository call in that
// callback shares the same transaction.
type TxManager struct {
	pool   *pgxpool.Pool
	logger Logger
}

var _ store.TxManager = TxManager{}

// NewTxManager creates a TxManager bound to the pgx pool.
func NewTxManager(p PgxPool, logger Logger) TxManager {
	return TxManager{pool: p.Pool, logger: logger}
}

// RunInTx begins a transaction, invokes fn with a transaction-bound Store, and
// commits on success. Any error (or panic) rolls the transaction back. The
// transaction-bound Store must be passed to each repository via its WithTx
// method so all statements execute on the same connection.
func (m TxManager) RunInTx(ctx context.Context, fn func(store.Store) error) error {
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
