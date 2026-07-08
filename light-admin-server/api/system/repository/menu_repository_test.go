package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

type menuMockQuerier struct {
	store.Store

	getMenu    func(ctx context.Context, id string) (store.TMenu, error)
	countMenus func(ctx context.Context, arg store.CountMenusParams) (int64, error)
	listMenus  func(ctx context.Context, arg store.ListMenusParams) ([]store.TMenu, error)

	lastListParams store.ListMenusParams
}

func (m *menuMockQuerier) GetMenu(ctx context.Context, id string) (store.TMenu, error) {
	return m.getMenu(ctx, id)
}

func (m *menuMockQuerier) CountMenus(ctx context.Context, arg store.CountMenusParams) (int64, error) {
	return m.countMenus(ctx, arg)
}

func (m *menuMockQuerier) ListMenus(ctx context.Context, arg store.ListMenusParams) ([]store.TMenu, error) {
	m.lastListParams = arg
	return m.listMenus(ctx, arg)
}

func newTestMenuRepo(q store.Store) MenuRepository {
	return MenuRepository{q: q, logger: lib.NopLogger()}
}

func TestMenuRepository_Get_NotFound(t *testing.T) {
	repo := newTestMenuRepo(&menuMockQuerier{
		getMenu: func(ctx context.Context, id string) (store.TMenu, error) {
			return store.TMenu{}, pgx.ErrNoRows
		},
	})
	_, err := repo.Get("x")
	assert.ErrorIs(t, err, apperrors.DatabaseRecordNotFound)
}

func TestMenuOrderMode(t *testing.T) {
	assert.Equal(t, int32(1), menuOrderMode(dto.OrderParam{Key: "sort", Direction: dto.OrderByASC}))
	assert.Equal(t, int32(0), menuOrderMode(dto.OrderParam{}), "default is id DESC (0)")
	assert.Equal(t, int32(0), menuOrderMode(dto.OrderParam{Key: "id"}))
}

func TestMenuRepository_Query_ParentIDPointerAndOrder(t *testing.T) {
	mq := &menuMockQuerier{
		countMenus: func(ctx context.Context, arg store.CountMenusParams) (int64, error) { return 1, nil },
		listMenus: func(ctx context.Context, arg store.ListMenusParams) ([]store.TMenu, error) {
			return []store.TMenu{{ID: "m1", Name: "n"}}, nil
		},
	}
	repo := newTestMenuRepo(mq)

	// A non-nil empty ParentID must still filter (parent_id = ''), while ordering
	// by "sort" selects the whitelisted mode 1.
	empty := ""
	_, err := repo.Query(&system.MenuQueryParam{
		ParentID:   &empty,
		OrderParam: dto.OrderParam{Key: "sort", Direction: dto.OrderByASC},
	})
	require.NoError(t, err)
	require.NotNil(t, mq.lastListParams.ParentID)
	assert.Equal(t, "", *mq.lastListParams.ParentID)
	assert.Equal(t, int32(1), mq.lastListParams.OrderBy)
}

func TestMenuRepository_Query_NilParentIDDisablesFilter(t *testing.T) {
	mq := &menuMockQuerier{
		countMenus: func(ctx context.Context, arg store.CountMenusParams) (int64, error) { return 1, nil },
		listMenus: func(ctx context.Context, arg store.ListMenusParams) ([]store.TMenu, error) {
			return []store.TMenu{{ID: "m1"}}, nil
		},
	}
	repo := newTestMenuRepo(mq)

	_, err := repo.Query(&system.MenuQueryParam{Keywords: "dash"})
	require.NoError(t, err)
	assert.Nil(t, mq.lastListParams.ParentID, "nil ParentID disables the predicate")
	require.NotNil(t, mq.lastListParams.Keywords)
	assert.Equal(t, "%dash%", *mq.lastListParams.Keywords)
}
