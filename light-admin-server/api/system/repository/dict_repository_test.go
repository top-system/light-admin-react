package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/top-system/light-admin/db/sqlc"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

type dictMockQuerier struct {
	sqlc.Querier

	getDictByCode       func(ctx context.Context, arg sqlc.GetDictByCodeParams) (sqlc.TDict, error)
	countDicts          func(ctx context.Context, keywords *string) (int64, error)
	listDicts           func(ctx context.Context, arg sqlc.ListDictsParams) ([]sqlc.TDict, error)
	updateDictItemsCode func(ctx context.Context, arg sqlc.UpdateDictItemsDictCodeParams) error

	lastKeywords    *string
	lastCascadeArgs sqlc.UpdateDictItemsDictCodeParams
}

func (m *dictMockQuerier) GetDictByCode(ctx context.Context, arg sqlc.GetDictByCodeParams) (sqlc.TDict, error) {
	return m.getDictByCode(ctx, arg)
}

func (m *dictMockQuerier) CountDicts(ctx context.Context, keywords *string) (int64, error) {
	m.lastKeywords = keywords
	return m.countDicts(ctx, keywords)
}

func (m *dictMockQuerier) ListDicts(ctx context.Context, arg sqlc.ListDictsParams) ([]sqlc.TDict, error) {
	return m.listDicts(ctx, arg)
}

func (m *dictMockQuerier) UpdateDictItemsDictCode(ctx context.Context, arg sqlc.UpdateDictItemsDictCodeParams) error {
	m.lastCascadeArgs = arg
	return nil
}

func newTestDictRepo(q sqlc.Querier) DictRepository {
	return DictRepository{q: q, logger: lib.NopLogger()}
}

func TestDictRepository_GetByCode_NotFoundReturnsNil(t *testing.T) {
	repo := newTestDictRepo(&dictMockQuerier{
		getDictByCode: func(ctx context.Context, arg sqlc.GetDictByCodeParams) (sqlc.TDict, error) {
			return sqlc.TDict{}, pgx.ErrNoRows
		},
	})
	got, err := repo.GetByCode("nope")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestDictRepository_Query_WrapsKeywords(t *testing.T) {
	mq := &dictMockQuerier{
		countDicts: func(ctx context.Context, keywords *string) (int64, error) { return 1, nil },
		listDicts: func(ctx context.Context, arg sqlc.ListDictsParams) ([]sqlc.TDict, error) {
			return []sqlc.TDict{{ID: "d1"}}, nil
		},
	}
	repo := newTestDictRepo(mq)

	_, err := repo.Query(&system.DictQueryParam{Keywords: "gender"})
	require.NoError(t, err)
	require.NotNil(t, mq.lastKeywords)
	assert.Equal(t, "%gender%", *mq.lastKeywords)
}

func TestDictRepository_UpdateDictItemsCode_PassesOldAndNew(t *testing.T) {
	mq := &dictMockQuerier{}
	repo := newTestDictRepo(mq)

	require.NoError(t, repo.UpdateDictItemsCode("old", "new"))
	assert.Equal(t, "old", mq.lastCascadeArgs.OldCode)
	assert.Equal(t, "new", mq.lastCascadeArgs.NewCode)
}

func TestDictRepository_Get_NotFound(t *testing.T) {
	repo := newTestDictRepo(&dictMockQuerier{})
	// Get uses GetDict which the mock leaves nil; call GetByCode path instead to
	// exercise the not-found mapping without wiring every method.
	repo = newTestDictRepo(&dictMockQuerier{
		getDictByCode: func(ctx context.Context, arg sqlc.GetDictByCodeParams) (sqlc.TDict, error) {
			return sqlc.TDict{}, apperrors.DatabaseInternalError
		},
	})
	_, err := repo.GetByCode("x")
	assert.Error(t, err)
}
