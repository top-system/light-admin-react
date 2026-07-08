package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/top-system/light-admin/db/sqlc"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// deptMockQuerier implements only the sqlc.Querier methods exercised by the dept
// repository tests. Unimplemented methods panic (nil embedded interface).
type deptMockQuerier struct {
	sqlc.Querier

	getDept          func(ctx context.Context, id string) (sqlc.TDept, error)
	getDeptByCode    func(ctx context.Context, arg sqlc.GetDeptByCodeParams) (sqlc.TDept, error)
	listDepts        func(ctx context.Context, arg sqlc.ListDeptsParams) ([]sqlc.TDept, error)
	softDeleteByPath func(ctx context.Context, arg sqlc.SoftDeleteDeptByTreePathParams) error

	lastListParams       sqlc.ListDeptsParams
	lastTreePathParams   sqlc.SoftDeleteDeptByTreePathParams
	lastCreateDeptParams sqlc.CreateDeptParams
}

func (m *deptMockQuerier) GetDept(ctx context.Context, id string) (sqlc.TDept, error) {
	return m.getDept(ctx, id)
}

func (m *deptMockQuerier) GetDeptByCode(ctx context.Context, arg sqlc.GetDeptByCodeParams) (sqlc.TDept, error) {
	return m.getDeptByCode(ctx, arg)
}

func (m *deptMockQuerier) ListDepts(ctx context.Context, arg sqlc.ListDeptsParams) ([]sqlc.TDept, error) {
	m.lastListParams = arg
	return m.listDepts(ctx, arg)
}

func (m *deptMockQuerier) CreateDept(ctx context.Context, arg sqlc.CreateDeptParams) error {
	m.lastCreateDeptParams = arg
	return nil
}

func (m *deptMockQuerier) SoftDeleteDeptByTreePath(ctx context.Context, arg sqlc.SoftDeleteDeptByTreePathParams) error {
	m.lastTreePathParams = arg
	if m.softDeleteByPath != nil {
		return m.softDeleteByPath(ctx, arg)
	}
	return nil
}

func newTestDeptRepo(q sqlc.Querier) DeptRepository {
	return DeptRepository{q: q, logger: lib.NopLogger()}
}

func sampleDeptRow() sqlc.TDept {
	now := time.Date(2026, 7, 7, 9, 0, 0, 0, time.Local)
	return sqlc.TDept{
		ID:         "d1",
		Name:       "研发部",
		Code:       "RD",
		ParentID:   "0",
		TreePath:   "0",
		Sort:       1,
		Status:     1,
		CreateTime: pgtype.Timestamptz{Time: now, Valid: true},
		UpdateTime: pgtype.Timestamptz{Time: now, Valid: true},
	}
}

func TestDeptRepository_Get_MapsRow(t *testing.T) {
	repo := newTestDeptRepo(&deptMockQuerier{
		getDept: func(ctx context.Context, id string) (sqlc.TDept, error) {
			assert.Equal(t, "d1", id)
			return sampleDeptRow(), nil
		},
	})

	got, err := repo.Get("d1")
	require.NoError(t, err)
	assert.Equal(t, "研发部", got.Name)
	assert.Equal(t, "RD", got.Code)
	assert.Equal(t, "0", got.TreePath)
	assert.Equal(t, 1, got.Status)
}

func TestDeptRepository_Get_NotFound(t *testing.T) {
	repo := newTestDeptRepo(&deptMockQuerier{
		getDept: func(ctx context.Context, id string) (sqlc.TDept, error) {
			return sqlc.TDept{}, pgx.ErrNoRows
		},
	})

	_, err := repo.Get("missing")
	assert.ErrorIs(t, err, apperrors.DatabaseRecordNotFound)
}

func TestDeptRepository_GetByCode_NotFoundReturnsNil(t *testing.T) {
	repo := newTestDeptRepo(&deptMockQuerier{
		getDeptByCode: func(ctx context.Context, arg sqlc.GetDeptByCodeParams) (sqlc.TDept, error) {
			return sqlc.TDept{}, pgx.ErrNoRows
		},
	})

	// The uniqueness check relies on (nil, nil) rather than an error.
	got, err := repo.GetByCode("NOPE")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestDeptRepository_GetByCode_PassesExcludeID(t *testing.T) {
	mq := &deptMockQuerier{
		getDeptByCode: func(ctx context.Context, arg sqlc.GetDeptByCodeParams) (sqlc.TDept, error) {
			require.NotNil(t, arg.ExcludeID)
			assert.Equal(t, "d1", *arg.ExcludeID)
			return sampleDeptRow(), nil
		},
	}
	repo := newTestDeptRepo(mq)

	_, err := repo.GetByCode("RD", "d1")
	require.NoError(t, err)
}

func TestDeptRepository_Query_MapsFilters(t *testing.T) {
	mq := &deptMockQuerier{
		listDepts: func(ctx context.Context, arg sqlc.ListDeptsParams) ([]sqlc.TDept, error) {
			return []sqlc.TDept{sampleDeptRow()}, nil
		},
	}
	repo := newTestDeptRepo(mq)

	status := 1
	list, err := repo.Query(&system.DeptQueryParam{Keywords: "研发", Status: &status})
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NotNil(t, mq.lastListParams.Keywords)
	assert.Equal(t, "%研发%", *mq.lastListParams.Keywords, "keywords wrapped in LIKE wildcards")
	require.NotNil(t, mq.lastListParams.Status)
	assert.Equal(t, int32(1), *mq.lastListParams.Status)
}

func TestDeptRepository_DeleteByTreePath_BuildsLikePattern(t *testing.T) {
	mq := &deptMockQuerier{}
	repo := newTestDeptRepo(mq)

	require.NoError(t, repo.DeleteByTreePath("d1", "admin"))
	assert.Equal(t, "d1", mq.lastTreePathParams.ID)
	assert.Equal(t, "admin", mq.lastTreePathParams.UpdateBy)
	assert.Equal(t, "%,d1,%", mq.lastTreePathParams.TreePathLike)
}

func TestDeptRepository_Create_AssignsUUID(t *testing.T) {
	mq := &deptMockQuerier{}
	repo := newTestDeptRepo(mq)

	dept := &system.Dept{Name: "测试", Code: "TEST", TreePath: "0"}
	require.NoError(t, repo.Create(dept))
	assert.Len(t, dept.ID, 32)
	assert.Equal(t, dept.ID, mq.lastCreateDeptParams.ID)
	assert.Equal(t, "测试", mq.lastCreateDeptParams.Name)
}
