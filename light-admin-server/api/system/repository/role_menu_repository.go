package repository

import (
	"context"

	"github.com/top-system/light-admin/db/sqlc"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// RoleMenuRepository is the sqlc/pgx-backed persistence layer for the role<->menu
// association table (t_role_menu).
type RoleMenuRepository struct {
	q      sqlc.Querier
	logger lib.Logger
}

// NewRoleMenuRepository creates a new role menu repository bound to the pool-level
// Queries.
func NewRoleMenuRepository(q *sqlc.Queries, logger lib.Logger) RoleMenuRepository {
	return RoleMenuRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries. Used to
// share a single pgx transaction across repositories via lib.TxManager.RunInTx.
func (a RoleMenuRepository) WithTx(q *sqlc.Queries) RoleMenuRepository {
	a.q = q
	return a
}

// GetMenuIDsByRoleID returns the menu IDs associated with a role.
func (a RoleMenuRepository) GetMenuIDsByRoleID(roleID string) ([]string, error) {
	menuIDs, err := a.q.GetMenuIDsByRoleID(context.Background(), roleID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return menuIDs, nil
}

// BatchCreate inserts multiple associations atomically in a single statement.
func (a RoleMenuRepository) BatchCreate(roleMenus []*system.RoleMenu) error {
	if len(roleMenus) == 0 {
		return nil
	}

	roleIDs := make([]string, len(roleMenus))
	menuIDs := make([]string, len(roleMenus))
	for i, rm := range roleMenus {
		roleIDs[i] = rm.RoleID
		menuIDs[i] = rm.MenuID
	}

	err := a.q.BatchCreateRoleMenus(context.Background(), sqlc.BatchCreateRoleMenusParams{
		RoleIds: roleIDs,
		MenuIds: menuIDs,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByRoleID removes all associations for a role.
func (a RoleMenuRepository) DeleteByRoleID(roleID string) error {
	if err := a.q.DeleteRoleMenusByRoleID(context.Background(), roleID); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByMenuID removes all associations for a menu. Used by MenuService.Delete.
func (a RoleMenuRepository) DeleteByMenuID(menuID string) error {
	if err := a.q.DeleteRoleMenusByMenuID(context.Background(), menuID); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}
