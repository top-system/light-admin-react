package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
)

type configMockQuerier struct {
	store.Store

	getConfig        func(ctx context.Context, id string) (store.TConfig, error)
	countConfigsByID func(ctx context.Context, arg store.CountConfigsByKeyParams) (int64, error)

	lastCountArgs store.CountConfigsByKeyParams
}

func (m *configMockQuerier) GetConfig(ctx context.Context, id string) (store.TConfig, error) {
	return m.getConfig(ctx, id)
}

func (m *configMockQuerier) CountConfigsByKey(ctx context.Context, arg store.CountConfigsByKeyParams) (int64, error) {
	m.lastCountArgs = arg
	return m.countConfigsByID(ctx, arg)
}

func newTestConfigRepo(q store.Store) ConfigRepository {
	return ConfigRepository{q: q, logger: lib.NopLogger()}
}

func TestConfigRepository_Get_NotFound(t *testing.T) {
	repo := newTestConfigRepo(&configMockQuerier{
		getConfig: func(ctx context.Context, id string) (store.TConfig, error) {
			return store.TConfig{}, store.ErrNoRows
		},
	})
	_, err := repo.Get("x")
	assert.ErrorIs(t, err, apperrors.DatabaseRecordNotFound)
}

func TestConfigRepository_ExistsByKey_PassesExcludeID(t *testing.T) {
	mq := &configMockQuerier{
		countConfigsByID: func(ctx context.Context, arg store.CountConfigsByKeyParams) (int64, error) {
			return 1, nil
		},
	}
	repo := newTestConfigRepo(mq)

	exists, err := repo.ExistsByKey("sys.name", "id-1")
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "sys.name", mq.lastCountArgs.ConfigKey)
	require.NotNil(t, mq.lastCountArgs.ExcludeID)
	assert.Equal(t, "id-1", *mq.lastCountArgs.ExcludeID)
}

func TestConfigRepository_ExistsByKey_NoExcludeIsNil(t *testing.T) {
	mq := &configMockQuerier{
		countConfigsByID: func(ctx context.Context, arg store.CountConfigsByKeyParams) (int64, error) {
			return 0, nil
		},
	}
	repo := newTestConfigRepo(mq)

	exists, err := repo.ExistsByKey("sys.name", "")
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, mq.lastCountArgs.ExcludeID)
}
