package sqlitestore

// Hand-written store.Store methods that gen-store cannot derive for SQLite:
//
//   - ListMenus: sqlc's sqlite engine does not rewrite parameters inside
//     ORDER BY CASE expressions, so the ordering selector is split into two
//     generated queries and dispatched here.
//   - BatchCreate*: SQLite has no unnest(); the postgres single-statement bulk
//     inserts become loops over the generated single-row inserts. Callers that
//     need atomicity already run inside TxManager.RunInTx.
//
// The generated store.go pins `var _ store.Store = (*Store)(nil)`, so a
// missing or drifting method here is a compile error (multi-database-plan §8).

import (
	"context"

	"github.com/top-system/light-admin/db/sqlitegen"
	"github.com/top-system/light-admin/db/store"
)

// ListMenus dispatches on OrderBy: 1 orders by sort ASC, anything else keeps
// the id DESC default — the same whitelist the postgres CASE expression
// implements.
func (s *Store) ListMenus(ctx context.Context, arg store.ListMenusParams) ([]store.TMenu, error) {
	var (
		rows []sqlitegen.TMenu
		err  error
	)
	if arg.OrderBy == 1 {
		rows, err = s.q.ListMenusOrderBySort(ctx, sqlitegen.ListMenusOrderBySortParams{
			Ids:            csvList(arg.Ids),
			Name:           arg.Name,
			ParentID:       arg.ParentID,
			PrefixTreePath: arg.PrefixTreePath,
			Type:           arg.Type,
			Visible:        arg.Visible,
			Keywords:       arg.Keywords,
			Limit:          limitOrAll(arg.Limit),
			Offset:         offsetOrZero(arg.Offset),
		})
	} else {
		rows, err = s.q.ListMenusOrderByID(ctx, sqlitegen.ListMenusOrderByIDParams{
			Ids:            csvList(arg.Ids),
			Name:           arg.Name,
			ParentID:       arg.ParentID,
			PrefixTreePath: arg.PrefixTreePath,
			Type:           arg.Type,
			Visible:        arg.Visible,
			Keywords:       arg.Keywords,
			Limit:          limitOrAll(arg.Limit),
			Offset:         offsetOrZero(arg.Offset),
		})
	}
	if err != nil {
		return nil, translateErr(err)
	}
	out := make([]store.TMenu, len(rows))
	for i := range rows {
		out[i] = store.TMenu(rows[i])
	}
	return out, nil
}

// BatchCreateRoleMenus inserts each (role_id, menu_id) pair; INSERT OR IGNORE
// preserves the postgres ON CONFLICT DO NOTHING semantics.
func (s *Store) BatchCreateRoleMenus(ctx context.Context, arg store.BatchCreateRoleMenusParams) error {
	for i := range arg.RoleIds {
		if err := s.q.CreateRoleMenu(ctx, sqlitegen.CreateRoleMenuParams{
			RoleID: arg.RoleIds[i],
			MenuID: arg.MenuIds[i],
		}); err != nil {
			return translateErr(err)
		}
	}
	return nil
}

// BatchCreateUserRoles inserts each (user_id, role_id) pair; INSERT OR IGNORE
// preserves the postgres ON CONFLICT DO NOTHING semantics.
func (s *Store) BatchCreateUserRoles(ctx context.Context, arg store.BatchCreateUserRolesParams) error {
	for i := range arg.UserIds {
		if err := s.q.CreateUserRole(ctx, sqlitegen.CreateUserRoleParams{
			UserID: arg.UserIds[i],
			RoleID: arg.RoleIds[i],
		}); err != nil {
			return translateErr(err)
		}
	}
	return nil
}

// BatchCreateUserNotices inserts one t_user_notice row per id triple.
func (s *Store) BatchCreateUserNotices(ctx context.Context, arg store.BatchCreateUserNoticesParams) error {
	for i := range arg.Ids {
		if err := s.q.CreateUserNotice(ctx, sqlitegen.CreateUserNoticeParams{
			ID:       arg.Ids[i],
			NoticeID: arg.NoticeIds[i],
			UserID:   arg.UserIds[i],
			IsRead:   0,
		}); err != nil {
			return translateErr(err)
		}
	}
	return nil
}
