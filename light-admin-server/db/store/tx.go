package store

import "context"

// TxManager runs a unit of work inside a single database transaction. The
// callback receives a transaction-bound Store; every repository call made
// through that Store shares the same transaction.
//
// It replaces the previous GORM `WithTrx(*gorm.DB)` pattern: services compose
// multiple store-backed repositories inside one RunInTx callback, and every
// repository call in that callback runs on the same transaction.
type TxManager interface {
	RunInTx(ctx context.Context, fn func(Store) error) error
}
