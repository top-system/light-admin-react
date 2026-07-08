package repository

import (
	"context"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

// UserRoleRepository is the sqlc/pgx-backed persistence layer for the
// user<->role association table.
type UserRoleRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewUserRoleRepository creates a new user role repository bound to the
// pool-level Queries.
func NewUserRoleRepository(q store.Store, logger lib.Logger) UserRoleRepository {
	return UserRoleRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a UserRoleRepository) WithTx(q store.Store) UserRoleRepository {
	a.q = q
	return a
}

// Query lists user-role associations by user filters. Ordering is fixed to
// user_id DESC (matching the previous default) and pagination follows the same
// default-page-size-15 semantics as the rest of the system.
func (a UserRoleRepository) Query(param *system.UserRoleQueryParam) (*system.UserRoleQueryResult, error) {
	ctx := context.Background()

	var userID *string
	if param.UserID != "" {
		userID = ptr(param.UserID)
	}
	var userIDs []string
	if len(param.UserIDs) > 0 {
		userIDs = param.UserIDs
	}

	total, err := a.q.CountUserRoles(ctx, store.CountUserRolesParams{
		UserID:  userID,
		UserIds: userIDs,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.UserRoles, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListUserRoles(ctx, store.ListUserRolesParams{
			UserID:  userID,
			UserIds: userIDs,
			Limit:   limit,
			Offset:  offset,
		})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, &system.UserRole{UserID: r.UserID, RoleID: r.RoleID})
		}
	}

	return &system.UserRoleQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

// GetRoleIDsByUserID returns the role IDs assigned to a user.
func (a UserRoleRepository) GetRoleIDsByUserID(userID string) ([]string, error) {
	roleIDs, err := a.q.GetRoleIDsByUserID(context.Background(), userID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return roleIDs, nil
}

// GetUserIDsByRoleID returns the user IDs assigned to a role.
func (a UserRoleRepository) GetUserIDsByRoleID(roleID string) ([]string, error) {
	userIDs, err := a.q.GetUserIDsByRoleID(context.Background(), roleID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return userIDs, nil
}

// Create inserts a single user-role association (idempotent).
func (a UserRoleRepository) Create(userRole *system.UserRole) error {
	err := a.q.CreateUserRole(context.Background(), store.CreateUserRoleParams{
		UserID: userRole.UserID,
		RoleID: userRole.RoleID,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// BatchCreate inserts multiple associations atomically in a single statement.
func (a UserRoleRepository) BatchCreate(userRoles []*system.UserRole) error {
	if len(userRoles) == 0 {
		return nil
	}

	userIDs := make([]string, len(userRoles))
	roleIDs := make([]string, len(userRoles))
	for i, ur := range userRoles {
		userIDs[i] = ur.UserID
		roleIDs[i] = ur.RoleID
	}

	err := a.q.BatchCreateUserRoles(context.Background(), store.BatchCreateUserRolesParams{
		UserIds: userIDs,
		RoleIds: roleIDs,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByUserID removes all associations for a user.
func (a UserRoleRepository) DeleteByUserID(userID string) error {
	if err := a.q.DeleteUserRolesByUserID(context.Background(), userID); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByRoleID removes all associations for a role.
func (a UserRoleRepository) DeleteByRoleID(roleID string) error {
	if err := a.q.DeleteUserRolesByRoleID(context.Background(), roleID); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}
