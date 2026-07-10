-- menu.sql (MySQL dialect)
-- Derived from queries/postgres/menu.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). Two engine-specific deviations:
--
--  * The nullable id-list filter has no array type here: the adapter passes
--    the ids as a comma-separated string ("a,b") and FIND_IN_SET tests
--    membership; NULL disables the filter (same semantics as the postgres
--    narg text[]).
--  * The postgres ListMenus ORDER BY CASE selector is split into two queries
--    (kept in lockstep with the sqlite dialect); the mysqlstore adapter
--    dispatches on OrderBy.

-- name: ListMenusOrderBySort :many
SELECT * FROM t_menu
WHERE (sqlc.narg('ids')              IS NULL OR FIND_IN_SET(id, sqlc.narg('ids')))
  AND (sqlc.narg('name')             IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('parent_id')        IS NULL OR parent_id = sqlc.narg('parent_id'))
  AND (sqlc.narg('prefix_tree_path') IS NULL OR tree_path LIKE sqlc.narg('prefix_tree_path'))
  AND (sqlc.narg('type')             IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('visible')          IS NULL OR visible = sqlc.narg('visible'))
  AND (sqlc.narg('keywords')         IS NULL OR name LIKE sqlc.narg('keywords'))
ORDER BY sort ASC
LIMIT  ?
OFFSET ?;

-- name: ListMenusOrderByID :many
SELECT * FROM t_menu
WHERE (sqlc.narg('ids')              IS NULL OR FIND_IN_SET(id, sqlc.narg('ids')))
  AND (sqlc.narg('name')             IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('parent_id')        IS NULL OR parent_id = sqlc.narg('parent_id'))
  AND (sqlc.narg('prefix_tree_path') IS NULL OR tree_path LIKE sqlc.narg('prefix_tree_path'))
  AND (sqlc.narg('type')             IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('visible')          IS NULL OR visible = sqlc.narg('visible'))
  AND (sqlc.narg('keywords')         IS NULL OR name LIKE sqlc.narg('keywords'))
ORDER BY id DESC
LIMIT  ?
OFFSET ?;

-- name: CountMenus :one
SELECT COUNT(*) FROM t_menu
WHERE (sqlc.narg('ids')              IS NULL OR FIND_IN_SET(id, sqlc.narg('ids')))
  AND (sqlc.narg('name')             IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('parent_id')        IS NULL OR parent_id = sqlc.narg('parent_id'))
  AND (sqlc.narg('prefix_tree_path') IS NULL OR tree_path LIKE sqlc.narg('prefix_tree_path'))
  AND (sqlc.narg('type')             IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('visible')          IS NULL OR visible = sqlc.narg('visible'))
  AND (sqlc.narg('keywords')         IS NULL OR name LIKE sqlc.narg('keywords'));

-- name: GetMenu :one
SELECT * FROM t_menu WHERE id = sqlc.arg('id');

-- name: CreateMenu :exec
INSERT INTO t_menu (
    id, parent_id, tree_path, name, type, route_name, route_path, component,
    perm, always_show, keep_alive, visible, sort, icon, redirect, params,
    create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('parent_id'), sqlc.arg('tree_path'), sqlc.arg('name'),
    sqlc.arg('type'), sqlc.arg('route_name'), sqlc.arg('route_path'), sqlc.arg('component'),
    sqlc.arg('perm'), sqlc.arg('always_show'), sqlc.arg('keep_alive'), sqlc.arg('visible'),
    sqlc.arg('sort'), sqlc.arg('icon'), sqlc.arg('redirect'), sqlc.arg('params'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateMenu :exec
UPDATE t_menu SET
    parent_id   = sqlc.arg('parent_id'),
    tree_path   = sqlc.arg('tree_path'),
    name        = sqlc.arg('name'),
    type        = sqlc.arg('type'),
    route_name  = sqlc.arg('route_name'),
    route_path  = sqlc.arg('route_path'),
    component   = sqlc.arg('component'),
    perm        = sqlc.arg('perm'),
    always_show = sqlc.arg('always_show'),
    keep_alive  = sqlc.arg('keep_alive'),
    visible     = sqlc.arg('visible'),
    sort        = sqlc.arg('sort'),
    icon        = sqlc.arg('icon'),
    redirect    = sqlc.arg('redirect'),
    params      = sqlc.arg('params'),
    update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: DeleteMenu :exec
DELETE FROM t_menu WHERE id = sqlc.arg('id');

-- name: UpdateMenuVisible :exec
UPDATE t_menu SET visible = sqlc.arg('visible'), update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');

-- name: UpdateMenuTreePath :exec
UPDATE t_menu SET tree_path = sqlc.arg('tree_path'), update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');

-- name: ListMenusByRoleIDs :many
-- Menus assigned to any of the given roles, ordered by sort ASC.
SELECT * FROM t_menu
WHERE id IN (SELECT DISTINCT menu_id FROM t_role_menu WHERE role_id IN (sqlc.slice('role_ids')))
ORDER BY sort ASC;

-- name: ListButtonPermsByRoleIDs :many
-- Non-empty button (type = 4) permission strings for the given roles.
SELECT perm FROM t_menu
WHERE id IN (SELECT DISTINCT menu_id FROM t_role_menu WHERE role_id IN (sqlc.slice('role_ids')))
  AND type = 4
  AND perm <> '';
