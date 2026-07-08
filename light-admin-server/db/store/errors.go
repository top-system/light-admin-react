package store

import "errors"

// ErrNoRows is the engine-neutral "no rows in result set" sentinel. Every
// adapter translates its driver's not-found error into this value (pgstore maps
// pgx.ErrNoRows), so business code checks errors.Is(err, store.ErrNoRows)
// without importing any database driver.
var ErrNoRows = errors.New("store: no rows in result set")
