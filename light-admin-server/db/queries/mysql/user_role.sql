-- user_role.sql (MySQL dialect)
-- Derived from queries/postgres/user_role.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). The nullable user-id list uses the
-- FIND_IN_SET membership test (see menu.sql header). MySQL has no
-- unnest(): BatchCreateUserRoles is implemented in the mysqlstore adapter as
-- a loop over CreateUserRole (INSERT IGNORE replaces ON CONFLICT).

-- name: ListUserRoles :many
-- Ordering is fixed (user_id DESC) rather than interpolated, eliminating any
-- ORDER BY injection surface. This matches the sole caller's default ordering.
SELECT user_id, role_id FROM t_user_role
WHERE (sqlc.narg('user_id')  IS NULL OR user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('user_ids') IS NULL OR FIND_IN_SET(user_id, sqlc.narg('user_ids')))
ORDER BY user_id DESC
LIMIT  ?
OFFSET ?;

-- name: CountUserRoles :one
SELECT COUNT(*) FROM t_user_role
WHERE (sqlc.narg('user_id')  IS NULL OR user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('user_ids') IS NULL OR FIND_IN_SET(user_id, sqlc.narg('user_ids')));

-- name: GetRoleIDsByUserID :many
SELECT role_id FROM t_user_role WHERE user_id = sqlc.arg('user_id');

-- name: GetUserIDsByRoleID :many
SELECT user_id FROM t_user_role WHERE role_id = sqlc.arg('role_id');

-- name: CreateUserRole :exec
INSERT IGNORE INTO t_user_role (user_id, role_id) VALUES (sqlc.arg('user_id'), sqlc.arg('role_id'));

-- name: DeleteUserRolesByUserID :exec
DELETE FROM t_user_role WHERE user_id = sqlc.arg('user_id');

-- name: DeleteUserRolesByRoleID :exec
DELETE FROM t_user_role WHERE role_id = sqlc.arg('role_id');
