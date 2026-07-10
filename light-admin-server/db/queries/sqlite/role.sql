-- role.sql (SQLite dialect)
-- Derived from queries/postgres/role.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). The nullable id-list filter uses the
-- comma-wrapped instr() membership test (see menu.sql header).

-- name: GetRole :one
SELECT * FROM t_role
WHERE id = sqlc.arg('id') AND is_deleted = 0;

-- name: GetRoleByCode :one
SELECT * FROM t_role
WHERE code = sqlc.arg('code') AND is_deleted = 0
LIMIT 1;

-- name: ListRoles :many
SELECT * FROM t_role
WHERE is_deleted = 0
  AND (sqlc.narg('ids')     IS NULL OR instr(sqlc.narg('ids'), ',' || id || ',') > 0)
  AND (sqlc.narg('name')    IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('code')    IS NULL OR code = sqlc.narg('code'))
  AND (sqlc.narg('user_id') IS NULL OR id IN (
        SELECT role_id FROM t_user_role WHERE user_id = sqlc.narg('user_id')))
  AND (sqlc.narg('query_value') IS NULL OR (
        name LIKE sqlc.narg('query_value') OR code LIKE sqlc.narg('query_value')))
  AND (sqlc.narg('status')  IS NULL OR status = sqlc.narg('status'))
ORDER BY id DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountRoles :one
SELECT COUNT(*) FROM t_role
WHERE is_deleted = 0
  AND (sqlc.narg('ids')     IS NULL OR instr(sqlc.narg('ids'), ',' || id || ',') > 0)
  AND (sqlc.narg('name')    IS NULL OR name = sqlc.narg('name'))
  AND (sqlc.narg('code')    IS NULL OR code = sqlc.narg('code'))
  AND (sqlc.narg('user_id') IS NULL OR id IN (
        SELECT role_id FROM t_user_role WHERE user_id = sqlc.narg('user_id')))
  AND (sqlc.narg('query_value') IS NULL OR (
        name LIKE sqlc.narg('query_value') OR code LIKE sqlc.narg('query_value')))
  AND (sqlc.narg('status')  IS NULL OR status = sqlc.narg('status'));

-- name: CreateRole :exec
INSERT INTO t_role (
    id, name, code, sort, status, data_scope, create_by, update_by, is_deleted,
    create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('name'), sqlc.arg('code'), sqlc.arg('sort'),
    sqlc.arg('status'), sqlc.arg('data_scope'), sqlc.arg('create_by'),
    sqlc.arg('update_by'), sqlc.arg('is_deleted'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateRole :exec
-- Mirrors the previous Select-scoped GORM update
-- (name, code, sort, status, data_scope, update_by).
UPDATE t_role SET
    name        = sqlc.arg('name'),
    code        = sqlc.arg('code'),
    sort        = sqlc.arg('sort'),
    status      = sqlc.arg('status'),
    data_scope  = sqlc.arg('data_scope'),
    update_by   = sqlc.arg('update_by'),
    update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: SoftDeleteRole :exec
UPDATE t_role SET is_deleted = 1, update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');

-- name: UpdateRoleStatus :exec
UPDATE t_role SET status = sqlc.arg('status'), update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');
