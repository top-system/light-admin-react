-- role_menu.sql (SQLite dialect)
-- Derived from queries/postgres/role_menu.sql. Every parameter is a named
-- sqlc.arg (see config.sql header). SQLite has no unnest(): the postgres
-- BatchCreateRoleMenus bulk insert is implemented in the sqlitestore adapter
-- as a loop over CreateRoleMenu (INSERT OR IGNORE replaces ON CONFLICT
-- DO NOTHING).

-- name: GetMenuIDsByRoleID :many
SELECT menu_id FROM t_role_menu WHERE role_id = sqlc.arg('role_id');

-- name: CreateRoleMenu :exec
INSERT OR IGNORE INTO t_role_menu (role_id, menu_id) VALUES (sqlc.arg('role_id'), sqlc.arg('menu_id'));

-- name: DeleteRoleMenusByRoleID :exec
DELETE FROM t_role_menu WHERE role_id = sqlc.arg('role_id');

-- name: DeleteRoleMenusByMenuID :exec
DELETE FROM t_role_menu WHERE menu_id = sqlc.arg('menu_id');
