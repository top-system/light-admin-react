-- user_role.sql
-- SQL for the user<->role association (t_user_role).

-- name: ListUserRoles :many
-- Ordering is fixed (user_id DESC) rather than interpolated, eliminating any
-- ORDER BY injection surface. This matches the sole caller's default ordering.
SELECT user_id, role_id FROM t_user_role
WHERE (sqlc.narg('user_id')::text   IS NULL OR user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('user_ids')::text[] IS NULL OR user_id = ANY(sqlc.narg('user_ids')::text[]))
ORDER BY user_id DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountUserRoles :one
SELECT COUNT(*) FROM t_user_role
WHERE (sqlc.narg('user_id')::text   IS NULL OR user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('user_ids')::text[] IS NULL OR user_id = ANY(sqlc.narg('user_ids')::text[]));

-- name: GetRoleIDsByUserID :many
SELECT role_id FROM t_user_role WHERE user_id = $1;

-- name: GetUserIDsByRoleID :many
SELECT user_id FROM t_user_role WHERE role_id = $1;

-- name: CreateUserRole :exec
INSERT INTO t_user_role (user_id, role_id) VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: BatchCreateUserRoles :exec
-- Single-statement bulk insert via array unnest, keeping the whole batch atomic.
INSERT INTO t_user_role (user_id, role_id)
SELECT unnest(@user_ids::text[]), unnest(@role_ids::text[])
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: DeleteUserRolesByUserID :exec
DELETE FROM t_user_role WHERE user_id = $1;

-- name: DeleteUserRolesByRoleID :exec
DELETE FROM t_user_role WHERE role_id = $1;
