-- role.sql
-- All SQL for the role domain (t_role). Generated into db/sqlc by `make sqlc`.
-- Dynamic filters use sqlc.narg (nullable args): a NULL argument disables that
-- predicate, so a single prepared statement serves every filter combination
-- while remaining fully parameterized (no SQL injection surface). Ordering is a
-- fixed `id DESC` — the previous GORM default — rather than an interpolated
-- column, eliminating any ORDER BY injection surface.

-- name: GetRole :one
SELECT * FROM t_role
WHERE id = $1 AND is_deleted = 0;

-- name: GetRoleByCode :one
SELECT * FROM t_role
WHERE code = $1 AND is_deleted = 0
LIMIT 1;

-- name: ListRoles :many
SELECT * FROM t_role
WHERE is_deleted = 0
  AND (sqlc.narg('ids')::text[]  IS NULL OR id = ANY(sqlc.narg('ids')::text[]))
  AND (sqlc.narg('name')::text   IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('code')::text   IS NULL OR code = sqlc.narg('code'))
  AND (sqlc.narg('user_id')::text IS NULL OR id IN (
        SELECT role_id FROM t_user_role WHERE user_id = sqlc.narg('user_id')))
  AND (sqlc.narg('query_value')::text IS NULL OR (
        name LIKE sqlc.narg('query_value') OR code LIKE sqlc.narg('query_value')))
  AND (sqlc.narg('status')::int  IS NULL OR status = sqlc.narg('status'))
ORDER BY id DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountRoles :one
SELECT COUNT(*) FROM t_role
WHERE is_deleted = 0
  AND (sqlc.narg('ids')::text[]  IS NULL OR id = ANY(sqlc.narg('ids')::text[]))
  AND (sqlc.narg('name')::text   IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('code')::text   IS NULL OR code = sqlc.narg('code'))
  AND (sqlc.narg('user_id')::text IS NULL OR id IN (
        SELECT role_id FROM t_user_role WHERE user_id = sqlc.narg('user_id')))
  AND (sqlc.narg('query_value')::text IS NULL OR (
        name LIKE sqlc.narg('query_value') OR code LIKE sqlc.narg('query_value')))
  AND (sqlc.narg('status')::int  IS NULL OR status = sqlc.narg('status'));

-- name: CreateRole :exec
INSERT INTO t_role (
    id, name, code, sort, status, data_scope, create_by, update_by, is_deleted,
    create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateRole :exec
-- Mirrors the previous Select-scoped GORM update
-- (name, code, sort, status, data_scope, update_by).
UPDATE t_role SET
    name        = $2,
    code        = $3,
    sort        = $4,
    status      = $5,
    data_scope  = $6,
    update_by   = $7,
    update_time = sqlc.arg('now')
WHERE id = $1;

-- name: SoftDeleteRole :exec
UPDATE t_role SET is_deleted = 1, update_time = sqlc.arg('now') WHERE id = $1;

-- name: UpdateRoleStatus :exec
UPDATE t_role SET status = $2, update_time = sqlc.arg('now') WHERE id = $1;
