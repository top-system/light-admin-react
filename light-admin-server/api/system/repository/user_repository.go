package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/uuid"
)

// UserRepository is the sqlc/pgx-backed persistence layer for users.
//
// It holds a store.Store (the pool-bound store.Store by default). Within a
// transaction, callers obtain a transaction-scoped copy via WithTx so every
// statement runs on the same pgx transaction.
type UserRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewUserRepository creates a new user repository bound to the pool-level Queries.
func NewUserRepository(q store.Store, logger lib.Logger) UserRepository {
	return UserRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy of the repository bound to the given transaction-scoped
// Queries. Use it inside lib.TxManager.RunInTx to share a single transaction
// across multiple repositories.
func (a UserRepository) WithTx(q store.Store) UserRepository {
	a.q = q
	return a
}

// Query lists users matching the given filters with pagination. It preserves the
// pagination semantics of the previous implementation: an unset page size
// defaults to 15 (see dto.PaginationParam.GetPageSize).
func (a UserRepository) Query(param *system.UserQueryParam) (*system.UserQueryResult, error) {
	ctx := context.Background()

	filter := newUserFilter(param)

	total, err := a.q.CountUsers(ctx, filter.count())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Users, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListUsers(ctx, filter.list(limit, offset))
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			u := toDomainUser(r)
			if !param.QueryPassword {
				u.Password = ""
			}
			list = append(list, u)
		}
	}

	return &system.UserQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

// Get returns a single non-deleted user by ID.
func (a UserRepository) Get(id string) (*system.User, error) {
	row, err := a.q.GetUser(context.Background(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainUser(row), nil
}

// GetByUsername returns a single non-deleted user by username (password included).
func (a UserRepository) GetByUsername(username string) (*system.User, error) {
	row, err := a.q.GetUserByUsername(context.Background(), username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainUser(row), nil
}

// Create inserts a new user, assigning a UUID when the ID is empty.
func (a UserRepository) Create(user *system.User) error {
	if user.ID == "" {
		user.ID = uuid.NewID()
	}

	err := a.q.CreateUser(context.Background(), store.CreateUserParams{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Gender:    int32(user.Gender),
		Password:  user.Password,
		DeptID:    user.DeptID,
		Avatar:    user.Avatar,
		Mobile:    user.Mobile,
		Status:    int32(user.Status),
		Email:     user.Email,
		CreateBy:  user.CreateBy,
		UpdateBy:  user.UpdateBy,
		IsDeleted: int32(user.IsDeleted),
		Openid:    user.OpenID,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Update updates the mutable profile fields of a user (mirrors the previous
// Select-scoped GORM update).
func (a UserRepository) Update(id string, user *system.User) error {
	err := a.q.UpdateUser(context.Background(), store.UpdateUserParams{
		ID:       id,
		Username: user.Username,
		Nickname: user.Nickname,
		Gender:   int32(user.Gender),
		DeptID:   user.DeptID,
		Avatar:   user.Avatar,
		Mobile:   user.Mobile,
		Status:   int32(user.Status),
		Email:    user.Email,
		UpdateBy: user.UpdateBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete soft-deletes a user (is_deleted = 1).
func (a UserRepository) Delete(id string) error {
	if err := a.q.SoftDeleteUser(context.Background(), id); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// UpdateStatus updates a user's status.
func (a UserRepository) UpdateStatus(id string, status int) error {
	err := a.q.UpdateUserStatus(context.Background(), store.UpdateUserStatusParams{
		ID:     id,
		Status: int32(status),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// UpdatePassword updates a user's password hash.
func (a UserRepository) UpdatePassword(id string, password string) error {
	err := a.q.UpdateUserPassword(context.Background(), store.UpdateUserPasswordParams{
		ID:       id,
		Password: password,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// UpdateProfile updates only the non-empty profile fields supplied by the user.
// Empty fields are left unchanged (COALESCE with a NULL argument).
func (a UserRepository) UpdateProfile(id string, profile *system.ProfileForm) error {
	params := store.UpdateUserProfileParams{ID: id}
	if profile.Nickname != "" {
		params.Nickname = ptr(profile.Nickname)
	}
	if profile.Gender > 0 {
		params.Gender = ptr(int32(profile.Gender))
	}
	if profile.Avatar != "" {
		params.Avatar = ptr(profile.Avatar)
	}
	if profile.Mobile != "" {
		params.Mobile = ptr(profile.Mobile)
	}
	if profile.Email != "" {
		params.Email = ptr(profile.Email)
	}

	if err := a.q.UpdateUserProfile(context.Background(), params); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// userFilter translates a UserQueryParam into sqlc filter arguments once, so the
// list and count queries stay in sync.
type userFilter struct {
	username   *string
	nickname   *string
	status     *int32
	deptID     *string
	keywords   *string
	createFrom pgtype.Timestamptz
	createTo   pgtype.Timestamptz
	roleIDs    []string
}

func newUserFilter(param *system.UserQueryParam) userFilter {
	f := userFilter{}
	if param.Username != "" {
		f.username = ptr(param.Username)
	}
	if param.Nickname != "" {
		f.nickname = ptr(param.Nickname)
	}
	if param.Status != nil {
		f.status = ptr(int32(*param.Status))
	}
	if param.DeptID != "" {
		f.deptID = ptr(param.DeptID)
	}
	// QueryValue and Keywords are identical LIKE predicates; use whichever is set.
	if kw := param.Keywords; kw != "" {
		f.keywords = ptr("%" + kw + "%")
	} else if qv := param.QueryValue; qv != "" {
		f.keywords = ptr("%" + qv + "%")
	}
	f.createFrom = startOfDayFilter(param.CreateTimeFrom)
	f.createTo = endOfDayFilter(param.CreateTimeTo)
	if len(param.RoleIDs) > 0 {
		f.roleIDs = param.RoleIDs
	}
	return f
}

func (f userFilter) list(limit, offset *int32) store.ListUsersParams {
	return store.ListUsersParams{
		Username:   f.username,
		Nickname:   f.nickname,
		Status:     f.status,
		DeptID:     f.deptID,
		Keywords:   f.keywords,
		CreateFrom: f.createFrom,
		CreateTo:   f.createTo,
		RoleIds:    f.roleIDs,
		Limit:      limit,
		Offset:     offset,
	}
}

func (f userFilter) count() store.CountUsersParams {
	return store.CountUsersParams{
		Username:   f.username,
		Nickname:   f.nickname,
		Status:     f.status,
		DeptID:     f.deptID,
		Keywords:   f.keywords,
		CreateFrom: f.createFrom,
		CreateTo:   f.createTo,
		RoleIds:    f.roleIDs,
	}
}

// pageBounds replicates the previous pagination math: page/size when both are
// positive, size-only limit otherwise, and no limit when size is non-positive.
func pageBounds(pp dto.PaginationParam) (limit, offset *int32) {
	current, pageSize := pp.GetPageNum(), pp.GetPageSize()
	switch {
	case current > 0 && pageSize > 0:
		l := int32(pageSize)
		o := int32((current - 1) * pageSize)
		return &l, &o
	case pageSize > 0:
		l := int32(pageSize)
		return &l, nil
	default:
		return nil, nil
	}
}

func toDomainUser(r store.TUser) *system.User {
	return &system.User{
		ID:         r.ID,
		Username:   r.Username,
		Nickname:   r.Nickname,
		Gender:     int(r.Gender),
		Password:   r.Password,
		DeptID:     r.DeptID,
		Avatar:     r.Avatar,
		Mobile:     r.Mobile,
		Status:     int(r.Status),
		Email:      r.Email,
		CreateTime: dto.DateTime(r.CreateTime.Time),
		CreateBy:   r.CreateBy,
		UpdateTime: dto.DateTime(r.UpdateTime.Time),
		UpdateBy:   r.UpdateBy,
		IsDeleted:  int(r.IsDeleted),
		OpenID:     r.Openid,
	}
}

// startOfDayFilter parses a "2006-01-02" or "2006-01-02 15:04:05" string for a
// >= comparison. An empty or unparseable value disables the filter.
func startOfDayFilter(s string) pgtype.Timestamptz {
	if t, ok := parseFlexibleTime(s); ok {
		return pgtype.Timestamptz{Time: t, Valid: true}
	}
	return pgtype.Timestamptz{}
}

// endOfDayFilter parses a date/datetime for a <= comparison. A date-only value
// is extended to 23:59:59 to include the whole day (matching the previous
// `value + " 23:59:59"` behaviour).
func endOfDayFilter(s string) pgtype.Timestamptz {
	if s == "" {
		return pgtype.Timestamptz{}
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); err == nil {
		return pgtype.Timestamptz{Time: t, Valid: true}
	}
	if d, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return pgtype.Timestamptz{Time: d.Add(23*time.Hour + 59*time.Minute + 59*time.Second), Valid: true}
	}
	return pgtype.Timestamptz{}
}

func parseFlexibleTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func ptr[T any](v T) *T { return &v }
