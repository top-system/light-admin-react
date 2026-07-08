-- role_menu.sql
-- SQL for the role<->menu association (t_role_menu), scoped to the Role module's
-- access paths. The Menu module still reaches t_role_menu through GORM during the
-- transition, so its menu-side delete is intentionally NOT defined here.

-- name: GetMenuIDsByRoleID :many
SELECT menu_id FROM t_role_menu WHERE role_id = $1;

-- name: BatchCreateRoleMenus :exec
-- Single-statement bulk insert via array unnest, keeping the whole batch atomic.
INSERT INTO t_role_menu (role_id, menu_id)
SELECT unnest(@role_ids::text[]), unnest(@menu_ids::text[])
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- name: DeleteRoleMenusByRoleID :exec
DELETE FROM t_role_menu WHERE role_id = $1;

-- name: DeleteRoleMenusByMenuID :exec
DELETE FROM t_role_menu WHERE menu_id = $1;
