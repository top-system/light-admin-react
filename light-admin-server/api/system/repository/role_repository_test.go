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

// roleMockQuerier implements only the sqlc.Querier methods exercised by the role
// repository tests. Unimplemented methods stay nil (embedded interface) and panic
// if ever called, keeping the mock honest about what each test depends on.
type roleMockQuerier struct {
	sqlc.Querier

	getRole      func(ctx context.Context, id string) (sqlc.TRole, error)
	getRoleByCod func(ctx context.Context, code string) (sqlc.TRole, error)
	listRoles    func(ctx context.Context, arg sqlc.ListRolesParams) ([]sqlc.TRole, error)
	countRoles   func(ctx context.Context, arg sqlc.CountRolesParams) (int64, error)
	createRole   func(ctx context.Context, arg sqlc.CreateRoleParams) error

	lastListParams   sqlc.ListRolesParams
	lastCountParams  sqlc.CountRolesParams
	lastCreateParams sqlc.CreateRoleParams
}

func (m *roleMockQuerier) GetRole(ctx context.Context, id string) (sqlc.TRole, error) {
	return m.getRole(ctx, id)
}

func (m *roleMockQuerier) GetRoleByCode(ctx context.Context, code string) (sqlc.TRole, error) {
	return m.getRoleByCod(ctx, code)
}

func (m *roleMockQuerier) ListRoles(ctx context.Context, arg sqlc.ListRolesParams) ([]sqlc.TRole, error) {
	m.lastListParams = arg
	return m.listRoles(ctx, arg)
}

func (m *roleMockQuerier) CountRoles(ctx context.Context, arg sqlc.CountRolesParams) (int64, error) {
	m.lastCountParams = arg
	return m.countRoles(ctx, arg)
}

func (m *roleMockQuerier) CreateRole(ctx context.Context, arg sqlc.CreateRoleParams) error {
	m.lastCreateParams = arg
	if m.createRole != nil {
		return m.createRole(ctx, arg)
	}
	return nil
}

func newTestRoleRepo(q sqlc.Querier) RoleRepository {
	return RoleRepository{q: q, logger: lib.Logger{}}
}

func sampleRoleRow() sqlc.TRole {
	now := time.Date(2026, 7, 6, 10, 30, 0, 0, time.Local)
	return sqlc.TRole{
		ID:         "r1",
		Name:       "管理员",
		Code:       "ADMIN",
		Sort:       3,
		Status:     1,
		DataScope:  2,
		CreateBy:   "u1",
		CreateTime: pgtype.Timestamptz{Time: now, Valid: true},
		UpdateBy:   "u1",
		UpdateTime: pgtype.Timestamptz{Time: now, Valid: true},
		IsDeleted:  0,
	}
}

func TestRoleRepository_Get_MapsRow(t *testing.T) {
	row := sampleRoleRow()
	repo := newTestRoleRepo(&roleMockQuerier{
		getRole: func(ctx context.Context, id string) (sqlc.TRole, error) {
			assert.Equal(t, "r1", id)
			return row, nil
		},
	})

	got, err := repo.Get("r1")
	require.NoError(t, err)
	assert.Equal(t, "管理员", got.Name)
	assert.Equal(t, "ADMIN", got.Code)
	assert.Equal(t, 3, got.Sort)
	assert.Equal(t, 1, got.Status)
	assert.Equal(t, 2, got.DataScope)
	assert.Equal(t, row.CreateTime.Time, time.Time(got.CreateTime))
}

func TestRoleRepository_Get_NotFound(t *testing.T) {
	repo := newTestRoleRepo(&roleMockQuerier{
		getRole: func(ctx context.Context, id string) (sqlc.TRole, error) {
			return sqlc.TRole{}, pgx.ErrNoRows
		},
	})

	_, err := repo.Get("missing")
	assert.ErrorIs(t, err, apperrors.DatabaseRecordNotFound)
}

func TestRoleRepository_GetByCode_NotFound(t *testing.T) {
	repo := newTestRoleRepo(&roleMockQuerier{
		getRoleByCod: func(ctx context.Context, code string) (sqlc.TRole, error) {
			return sqlc.TRole{}, pgx.ErrNoRows
		},
	})

	_, err := repo.GetByCode("nope")
	assert.ErrorIs(t, err, apperrors.DatabaseRecordNotFound)
}

func TestRoleRepository_Query_PaginatesAndMapsFilters(t *testing.T) {
	mq := &roleMockQuerier{
		countRoles: func(ctx context.Context, arg sqlc.CountRolesParams) (int64, error) { return 1, nil },
		listRoles: func(ctx context.Context, arg sqlc.ListRolesParams) ([]sqlc.TRole, error) {
			return []sqlc.TRole{sampleRoleRow()}, nil
		},
	}
	repo := newTestRoleRepo(mq)

	qr, err := repo.Query(&system.RoleQueryParam{
		Name:   "管理员",
		Status: 1,
	})
	require.NoError(t, err)

	require.Len(t, qr.List, 1)
	assert.Equal(t, int64(1), qr.Pagination.Total)
	assert.Equal(t, 15, qr.Pagination.PageSize, "unset page size defaults to 15")

	// The name/status filters must reach both queries as non-nil arguments.
	require.NotNil(t, mq.lastListParams.Name)
	assert.Equal(t, "管理员", *mq.lastListParams.Name)
	require.NotNil(t, mq.lastListParams.Status)
	assert.Equal(t, int32(1), *mq.lastListParams.Status)
	require.NotNil(t, mq.lastCountParams.Name)
	require.NotNil(t, mq.lastCountParams.Status)
}

func TestRoleRepository_Query_EmptyResultSkipsList(t *testing.T) {
	listCalled := false
	repo := newTestRoleRepo(&roleMockQuerier{
		countRoles: func(ctx context.Context, arg sqlc.CountRolesParams) (int64, error) { return 0, nil },
		listRoles: func(ctx context.Context, arg sqlc.ListRolesParams) ([]sqlc.TRole, error) {
			listCalled = true
			return nil, nil
		},
	})

	qr, err := repo.Query(&system.RoleQueryParam{})
	require.NoError(t, err)
	assert.Empty(t, qr.List)
	assert.False(t, listCalled, "ListRoles should be skipped when count is zero")
}

func TestRoleRepository_Create_AssignsUUIDAndMapsParams(t *testing.T) {
	mq := &roleMockQuerier{}
	repo := newTestRoleRepo(mq)

	role := &system.Role{Name: "编辑", Code: "EDITOR", Sort: 5, Status: 1, DataScope: 1, CreateBy: "admin"}
	require.NoError(t, repo.Create(role))

	assert.NotEmpty(t, role.ID, "an empty ID must be populated on create")
	assert.Len(t, role.ID, 32, "generated ID is a 32-char UUID")
	assert.Equal(t, role.ID, mq.lastCreateParams.ID)
	assert.Equal(t, "编辑", mq.lastCreateParams.Name)
	assert.Equal(t, "EDITOR", mq.lastCreateParams.Code)
	assert.Equal(t, int32(5), mq.lastCreateParams.Sort)
	assert.Equal(t, int32(1), mq.lastCreateParams.Status)
	assert.Equal(t, int32(1), mq.lastCreateParams.DataScope)
	assert.Equal(t, "admin", mq.lastCreateParams.CreateBy)
}

func TestRoleRepository_Create_KeepsProvidedID(t *testing.T) {
	mq := &roleMockQuerier{}
	repo := newTestRoleRepo(mq)

	role := &system.Role{ID: "fixed-id", Name: "x", Code: "X"}
	require.NoError(t, repo.Create(role))
	assert.Equal(t, "fixed-id", role.ID)
	assert.Equal(t, "fixed-id", mq.lastCreateParams.ID)
}

func TestNewRoleFilter_ZeroValuesDisablePredicates(t *testing.T) {
	f := newRoleFilter(&system.RoleQueryParam{})
	assert.Nil(t, f.name)
	assert.Nil(t, f.code)
	assert.Nil(t, f.userID)
	assert.Nil(t, f.queryValue)
	assert.Nil(t, f.status)
	assert.Empty(t, f.ids)

	// QueryValue is wrapped in LIKE wildcards; Status maps through as int32.
	f = newRoleFilter(&system.RoleQueryParam{QueryValue: "abc", Status: 1})
	require.NotNil(t, f.queryValue)
	assert.Equal(t, "%abc%", *f.queryValue)
	require.NotNil(t, f.status)
	assert.Equal(t, int32(1), *f.status)
}
