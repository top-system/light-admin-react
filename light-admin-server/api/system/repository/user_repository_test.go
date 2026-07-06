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
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

// mockQuerier implements only the sqlc.Querier methods exercised by the tests.
// Unimplemented methods stay nil (embedded interface) and panic if ever called,
// which keeps the mock honest about what each test actually depends on.
type mockQuerier struct {
	sqlc.Querier

	getUser    func(ctx context.Context, id string) (sqlc.TUser, error)
	listUsers  func(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.TUser, error)
	countUsers func(ctx context.Context, arg sqlc.CountUsersParams) (int64, error)

	lastListParams  sqlc.ListUsersParams
	lastCountParams sqlc.CountUsersParams
}

func (m *mockQuerier) GetUser(ctx context.Context, id string) (sqlc.TUser, error) {
	return m.getUser(ctx, id)
}

func (m *mockQuerier) ListUsers(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.TUser, error) {
	m.lastListParams = arg
	return m.listUsers(ctx, arg)
}

func (m *mockQuerier) CountUsers(ctx context.Context, arg sqlc.CountUsersParams) (int64, error) {
	m.lastCountParams = arg
	return m.countUsers(ctx, arg)
}

func newTestUserRepo(q sqlc.Querier) UserRepository {
	return UserRepository{q: q, logger: lib.Logger{}}
}

func sampleRow() sqlc.TUser {
	now := time.Date(2026, 7, 6, 10, 30, 0, 0, time.Local)
	return sqlc.TUser{
		ID:         "u1",
		Username:   "alice",
		Nickname:   "Alice",
		Gender:     2,
		Password:   "secret-hash",
		DeptID:     "d1",
		Status:     1,
		Email:      "alice@example.com",
		CreateTime: pgtype.Timestamptz{Time: now, Valid: true},
		UpdateTime: pgtype.Timestamptz{Time: now, Valid: true},
		IsDeleted:  0,
		Openid:     "openid-1",
	}
}

func TestUserRepository_Get_MapsRow(t *testing.T) {
	row := sampleRow()
	repo := newTestUserRepo(&mockQuerier{
		getUser: func(ctx context.Context, id string) (sqlc.TUser, error) {
			assert.Equal(t, "u1", id)
			return row, nil
		},
	})

	got, err := repo.Get("u1")
	require.NoError(t, err)
	assert.Equal(t, "alice", got.Username)
	assert.Equal(t, 2, got.Gender)
	assert.Equal(t, 1, got.Status)
	assert.Equal(t, "openid-1", got.OpenID)
	assert.Equal(t, "secret-hash", got.Password)
	assert.Equal(t, row.CreateTime.Time, time.Time(got.CreateTime))
}

func TestUserRepository_Get_NotFound(t *testing.T) {
	repo := newTestUserRepo(&mockQuerier{
		getUser: func(ctx context.Context, id string) (sqlc.TUser, error) {
			return sqlc.TUser{}, pgx.ErrNoRows
		},
	})

	_, err := repo.Get("missing")
	assert.ErrorIs(t, err, apperrors.DatabaseRecordNotFound)
}

func TestUserRepository_Query_BlanksPasswordAndPaginates(t *testing.T) {
	mq := &mockQuerier{
		countUsers: func(ctx context.Context, arg sqlc.CountUsersParams) (int64, error) {
			return 1, nil
		},
		listUsers: func(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.TUser, error) {
			return []sqlc.TUser{sampleRow()}, nil
		},
	}
	repo := newTestUserRepo(mq)

	qr, err := repo.Query(&system.UserQueryParam{
		Username: "alice",
		// QueryPassword defaults to false -> password must be stripped
	})
	require.NoError(t, err)

	require.Len(t, qr.List, 1)
	assert.Empty(t, qr.List[0].Password, "password must be blanked when QueryPassword is false")
	assert.Equal(t, int64(1), qr.Pagination.Total)
	assert.Equal(t, 15, qr.Pagination.PageSize, "unset page size defaults to 15")

	// The username filter must reach the query as a non-nil argument.
	require.NotNil(t, mq.lastListParams.Username)
	assert.Equal(t, "alice", *mq.lastListParams.Username)
	require.NotNil(t, mq.lastCountParams.Username)
}

func TestUserRepository_Query_KeepsPasswordWhenRequested(t *testing.T) {
	mq := &mockQuerier{
		countUsers: func(ctx context.Context, arg sqlc.CountUsersParams) (int64, error) { return 1, nil },
		listUsers: func(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.TUser, error) {
			return []sqlc.TUser{sampleRow()}, nil
		},
	}
	repo := newTestUserRepo(mq)

	qr, err := repo.Query(&system.UserQueryParam{QueryPassword: true})
	require.NoError(t, err)
	require.Len(t, qr.List, 1)
	assert.Equal(t, "secret-hash", qr.List[0].Password)
}

func TestUserRepository_Query_EmptyResultSkipsList(t *testing.T) {
	listCalled := false
	repo := newTestUserRepo(&mockQuerier{
		countUsers: func(ctx context.Context, arg sqlc.CountUsersParams) (int64, error) { return 0, nil },
		listUsers: func(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.TUser, error) {
			listCalled = true
			return nil, nil
		},
	})

	qr, err := repo.Query(&system.UserQueryParam{})
	require.NoError(t, err)
	assert.Empty(t, qr.List)
	assert.False(t, listCalled, "ListUsers should be skipped when count is zero")
}

func TestPageBounds(t *testing.T) {
	// page 2, size 10 -> limit 10 offset 10
	limit, offset := pageBounds(dto.PaginationParam{PageNum: 2, PageSize: 10})
	require.NotNil(t, limit)
	require.NotNil(t, offset)
	assert.Equal(t, int32(10), *limit)
	assert.Equal(t, int32(10), *offset)

	// size only (page 0) -> limit set, offset nil
	limit, offset = pageBounds(dto.PaginationParam{PageNum: 0, PageSize: 25})
	require.NotNil(t, limit)
	assert.Equal(t, int32(25), *limit)
	assert.Nil(t, offset)
}

func TestEndOfDayFilter(t *testing.T) {
	got := endOfDayFilter("2026-07-06")
	require.True(t, got.Valid)
	assert.Equal(t, 23, got.Time.Hour())
	assert.Equal(t, 59, got.Time.Minute())
	assert.Equal(t, 59, got.Time.Second())

	assert.False(t, endOfDayFilter("").Valid, "empty value disables the filter")
	assert.False(t, endOfDayFilter("not-a-date").Valid)
}
