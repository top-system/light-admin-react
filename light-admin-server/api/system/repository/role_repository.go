package repository

import (
	"context"
	"errors"
	"time"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/uuid"
)

// RoleRepository is the sqlc/pgx-backed persistence layer for roles.
//
// It holds a store.Store (the pool-bound store.Store by default). Within a
// transaction, callers obtain a transaction-scoped copy via WithTx so every
// statement runs on the same pgx transaction.
type RoleRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewRoleRepository creates a new role repository bound to the pool-level Queries.
func NewRoleRepository(q store.Store, logger lib.Logger) RoleRepository {
	return RoleRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy of the repository bound to the given transaction-scoped
// Queries. Use it inside lib.TxManager.RunInTx to share a single transaction
// across multiple repositories.
func (a RoleRepository) WithTx(q store.Store) RoleRepository {
	a.q = q
	return a
}

// Query lists roles matching the given filters with pagination. It preserves the
// pagination semantics of the previous implementation: an unset page size
// defaults to 15 (see dto.PaginationParam.GetPageSize) and a zero total skips the
// list query entirely.
func (a RoleRepository) Query(param *system.RoleQueryParam) (*system.RoleQueryResult, error) {
	ctx := context.Background()

	filter := newRoleFilter(param)

	total, err := a.q.CountRoles(ctx, filter.count())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Roles, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListRoles(ctx, filter.list(limit, offset))
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainRole(r))
		}
	}

	return &system.RoleQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

// Get returns a single non-deleted role by ID.
func (a RoleRepository) Get(id string) (*system.Role, error) {
	row, err := a.q.GetRole(context.Background(), id)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainRole(row), nil
}

// GetByCode returns a single non-deleted role by code.
func (a RoleRepository) GetByCode(code string) (*system.Role, error) {
	row, err := a.q.GetRoleByCode(context.Background(), code)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainRole(row), nil
}

// Create inserts a new role, assigning a UUID when the ID is empty (mirroring the
// previous GORM create-time callback). The role's ID is written back so callers
// can reference it after creation.
func (a RoleRepository) Create(role *system.Role) error {
	if role.ID == "" {
		role.ID = uuid.NewID()
	}

	err := a.q.CreateRole(context.Background(), store.CreateRoleParams{
		ID:        role.ID,
		Name:      role.Name,
		Code:      role.Code,
		Sort:      int32(role.Sort),
		Status:    int32(role.Status),
		DataScope: int32(role.DataScope),
		CreateBy:  role.CreateBy,
		UpdateBy:  role.UpdateBy,
		IsDeleted: int32(role.IsDeleted),
		Now:       time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Update updates the mutable fields of a role (mirrors the previous Select-scoped
// GORM update: name, code, sort, status, data_scope, update_by).
func (a RoleRepository) Update(id string, role *system.Role) error {
	err := a.q.UpdateRole(context.Background(), store.UpdateRoleParams{
		ID:        id,
		Name:      role.Name,
		Code:      role.Code,
		Sort:      int32(role.Sort),
		Status:    int32(role.Status),
		DataScope: int32(role.DataScope),
		UpdateBy:  role.UpdateBy,
		Now:       time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete soft-deletes a role (is_deleted = 1).
func (a RoleRepository) Delete(id string) error {
	if err := a.q.SoftDeleteRole(context.Background(), store.SoftDeleteRoleParams{ID: id, Now: time.Now()}); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// UpdateStatus updates a role's status.
func (a RoleRepository) UpdateStatus(id string, status int) error {
	err := a.q.UpdateRoleStatus(context.Background(), store.UpdateRoleStatusParams{
		ID:     id,
		Status: int32(status),
		Now:    time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// roleFilter translates a RoleQueryParam into sqlc filter arguments once, so the
// list and count queries stay in sync. A zero-value field disables its predicate,
// matching the previous GORM query builder (e.g. Status == 0 means "any status").
type roleFilter struct {
	ids        []string
	name       *string
	code       *string
	userID     *string
	queryValue *string
	status     *int32
}

func newRoleFilter(param *system.RoleQueryParam) roleFilter {
	f := roleFilter{}
	if len(param.IDs) > 0 {
		f.ids = param.IDs
	}
	if param.Name != "" {
		f.name = ptr(param.Name)
	}
	if param.Code != "" {
		f.code = ptr(param.Code)
	}
	if param.UserID != "" {
		f.userID = ptr(param.UserID)
	}
	if param.QueryValue != "" {
		f.queryValue = ptr("%" + param.QueryValue + "%")
	}
	if param.Status != 0 {
		f.status = ptr(int32(param.Status))
	}
	return f
}

func (f roleFilter) list(limit, offset *int32) store.ListRolesParams {
	return store.ListRolesParams{
		Ids:        f.ids,
		Name:       f.name,
		Code:       f.code,
		UserID:     f.userID,
		QueryValue: f.queryValue,
		Status:     f.status,
		Limit:      limit,
		Offset:     offset,
	}
}

func (f roleFilter) count() store.CountRolesParams {
	return store.CountRolesParams{
		Ids:        f.ids,
		Name:       f.name,
		Code:       f.code,
		UserID:     f.userID,
		QueryValue: f.queryValue,
		Status:     f.status,
	}
}

func toDomainRole(r store.TRole) *system.Role {
	return &system.Role{
		ID:         r.ID,
		Name:       r.Name,
		Code:       r.Code,
		Sort:       int(r.Sort),
		Status:     int(r.Status),
		DataScope:  int(r.DataScope),
		CreateBy:   r.CreateBy,
		CreateTime: dto.DateTime(r.CreateTime),
		UpdateBy:   r.UpdateBy,
		UpdateTime: dto.DateTime(r.UpdateTime),
		IsDeleted:  int(r.IsDeleted),
	}
}
