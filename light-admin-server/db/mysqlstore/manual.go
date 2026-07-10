package mysqlstore

// Hand-written store.Store methods that gen-store cannot derive for MySQL:
//
//   - ListMenus: the ordering selector is split into two generated queries
//     (kept in lockstep with the sqlite dialect) and dispatched here.
//   - BatchCreate*: MySQL has no unnest(); the postgres single-statement bulk
//     inserts become loops over the generated single-row inserts. Callers that
//     need atomicity already run inside TxManager.RunInTx.
//   - CreateQueueTask / UpdateQueueTask: MySQL has no RETURNING; the write is
//     followed by a re-read on the same DBTX (pool connection or transaction),
//     with LastInsertId supplying the created id.
//
// The generated store.go pins `var _ store.Store = (*Store)(nil)`, so a
// missing or drifting method here is a compile error (multi-database-plan §8).

import (
	"context"

	"github.com/top-system/light-admin/db/mysqlgen"
	"github.com/top-system/light-admin/db/store"
)

// ListMenus dispatches on OrderBy: 1 orders by sort ASC, anything else keeps
// the id DESC default — the same whitelist the postgres CASE expression
// implements.
func (s *Store) ListMenus(ctx context.Context, arg store.ListMenusParams) ([]store.TMenu, error) {
	var (
		rows []mysqlgen.TMenu
		err  error
	)
	if arg.OrderBy == 1 {
		rows, err = s.q.ListMenusOrderBySort(ctx, mysqlgen.ListMenusOrderBySortParams{
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
		rows, err = s.q.ListMenusOrderByID(ctx, mysqlgen.ListMenusOrderByIDParams{
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

// CreateQueueTask inserts the task (:execlastid) and re-reads the created row
// on the same DBTX, replacing the postgres RETURNING *.
func (s *Store) CreateQueueTask(ctx context.Context, arg store.CreateQueueTaskParams) (store.SysTask, error) {
	id, err := s.q.CreateQueueTask(ctx, mysqlgen.CreateQueueTaskParams(arg))
	if err != nil {
		return store.SysTask{}, translateErr(err)
	}
	row, err := s.q.GetQueueTask(ctx, id)
	return store.SysTask(row), translateErr(err)
}

// UpdateQueueTask updates the live row and re-reads it on the same DBTX,
// replacing the postgres RETURNING *. The re-read also preserves the postgres
// behaviour of reporting ErrNoRows for a soft-deleted (or missing) task.
func (s *Store) UpdateQueueTask(ctx context.Context, arg store.UpdateQueueTaskParams) (store.SysTask, error) {
	// Field order differs from the postgres struct (id is the last named arg
	// here), so the params are mapped explicitly rather than converted.
	if err := s.q.UpdateQueueTask(ctx, mysqlgen.UpdateQueueTaskParams{
		Type:                   arg.Type,
		Status:                 arg.Status,
		CorrelationID:          arg.CorrelationID,
		OwnerID:                arg.OwnerID,
		PrivateState:           arg.PrivateState,
		PublicRetryCount:       arg.PublicRetryCount,
		PublicExecutedDuration: arg.PublicExecutedDuration,
		PublicError:            arg.PublicError,
		PublicErrorHistory:     arg.PublicErrorHistory,
		PublicResumeTime:       arg.PublicResumeTime,
		Now:                    arg.Now,
		ID:                     arg.ID,
	}); err != nil {
		return store.SysTask{}, translateErr(err)
	}
	row, err := s.q.GetQueueTask(ctx, arg.ID)
	return store.SysTask(row), translateErr(err)
}

// BatchCreateRoleMenus inserts each (role_id, menu_id) pair; INSERT IGNORE
// preserves the postgres ON CONFLICT DO NOTHING semantics.
func (s *Store) BatchCreateRoleMenus(ctx context.Context, arg store.BatchCreateRoleMenusParams) error {
	for i := range arg.RoleIds {
		if err := s.q.CreateRoleMenu(ctx, mysqlgen.CreateRoleMenuParams{
			RoleID: arg.RoleIds[i],
			MenuID: arg.MenuIds[i],
		}); err != nil {
			return translateErr(err)
		}
	}
	return nil
}

// BatchCreateUserRoles inserts each (user_id, role_id) pair; INSERT IGNORE
// preserves the postgres ON CONFLICT DO NOTHING semantics.
func (s *Store) BatchCreateUserRoles(ctx context.Context, arg store.BatchCreateUserRolesParams) error {
	for i := range arg.UserIds {
		if err := s.q.CreateUserRole(ctx, mysqlgen.CreateUserRoleParams{
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
		if err := s.q.CreateUserNotice(ctx, mysqlgen.CreateUserNoticeParams{
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
