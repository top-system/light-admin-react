-- menu.sql
-- All SQL for the menu domain (t_menu). Dynamic filters use sqlc.narg (nullable
-- args) so one prepared statement serves every filter combination while staying
-- fully parameterized.
--
-- Ordering is whitelisted, not interpolated: the bound order_by selects between
-- the only two orderings the application uses — sort ASC (order_by = 1) and the
-- default id DESC (order_by = 0). Each ORDER BY term is active in exactly one
-- mode, so there is no ORDER BY injection surface.

-- name: ListMenus :many
SELECT * FROM t_menu
WHERE (sqlc.narg('ids')::text[]           IS NULL OR id = ANY(sqlc.narg('ids')::text[]))
  AND (sqlc.narg('name')::text            IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('parent_id')::text       IS NULL OR parent_id = sqlc.narg('parent_id'))
  AND (sqlc.narg('prefix_tree_path')::text IS NULL OR tree_path LIKE sqlc.narg('prefix_tree_path'))
  AND (sqlc.narg('type')::int             IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('visible')::int          IS NULL OR visible = sqlc.narg('visible'))
  AND (sqlc.narg('keywords')::text        IS NULL OR name LIKE sqlc.narg('keywords'))
ORDER BY
  CASE WHEN sqlc.arg('order_by')::int = 1 THEN sort END ASC,
  CASE WHEN sqlc.arg('order_by')::int = 0 THEN id END DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountMenus :one
SELECT COUNT(*) FROM t_menu
WHERE (sqlc.narg('ids')::text[]           IS NULL OR id = ANY(sqlc.narg('ids')::text[]))
  AND (sqlc.narg('name')::text            IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('parent_id')::text       IS NULL OR parent_id = sqlc.narg('parent_id'))
  AND (sqlc.narg('prefix_tree_path')::text IS NULL OR tree_path LIKE sqlc.narg('prefix_tree_path'))
  AND (sqlc.narg('type')::int             IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('visible')::int          IS NULL OR visible = sqlc.narg('visible'))
  AND (sqlc.narg('keywords')::text        IS NULL OR name LIKE sqlc.narg('keywords'));

-- name: GetMenu :one
SELECT * FROM t_menu WHERE id = $1;

-- name: CreateMenu :exec
INSERT INTO t_menu (
    id, parent_id, tree_path, name, type, route_name, route_path, component,
    perm, always_show, keep_alive, visible, sort, icon, redirect, params,
    create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14, $15, $16,
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateMenu :exec
UPDATE t_menu SET
    parent_id   = $2,
    tree_path   = $3,
    name        = $4,
    type        = $5,
    route_name  = $6,
    route_path  = $7,
    component   = $8,
    perm        = $9,
    always_show = $10,
    keep_alive  = $11,
    visible     = $12,
    sort        = $13,
    icon        = $14,
    redirect    = $15,
    params      = $16,
    update_time = sqlc.arg('now')
WHERE id = $1;

-- name: DeleteMenu :exec
DELETE FROM t_menu WHERE id = $1;

-- name: UpdateMenuVisible :exec
UPDATE t_menu SET visible = $2, update_time = sqlc.arg('now') WHERE id = $1;

-- name: UpdateMenuTreePath :exec
UPDATE t_menu SET tree_path = $2, update_time = sqlc.arg('now') WHERE id = $1;

-- name: ListMenusByRoleIDs :many
-- Menus assigned to any of the given roles, ordered by sort ASC.
SELECT * FROM t_menu
WHERE id IN (SELECT DISTINCT menu_id FROM t_role_menu WHERE role_id = ANY(@role_ids::text[]))
ORDER BY sort ASC;

-- name: ListButtonPermsByRoleIDs :many
-- Non-empty button (type = 4) permission strings for the given roles.
SELECT perm FROM t_menu
WHERE id IN (SELECT DISTINCT menu_id FROM t_role_menu WHERE role_id = ANY(@role_ids::text[]))
  AND type = 4
  AND perm <> '';
