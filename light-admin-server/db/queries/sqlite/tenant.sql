-- tenant.sql (SQLite dialect)
-- Derived from queries/postgres/tenant.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header).

-- name: GetTenant :one
SELECT * FROM t_tenant WHERE id = sqlc.arg('id') AND is_deleted = 0;

-- name: GetTenantByCode :one
SELECT * FROM t_tenant WHERE code = sqlc.arg('code') AND is_deleted = 0;

-- name: ListTenants :many
SELECT * FROM t_tenant
WHERE is_deleted = 0
  AND (sqlc.narg('keywords') IS NULL OR (code LIKE sqlc.narg('keywords') OR name LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('status')   IS NULL OR status = sqlc.narg('status'))
ORDER BY id DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountTenants :one
SELECT COUNT(*) FROM t_tenant
WHERE is_deleted = 0
  AND (sqlc.narg('keywords') IS NULL OR (code LIKE sqlc.narg('keywords') OR name LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('status')   IS NULL OR status = sqlc.narg('status'));

-- name: CreateTenant :exec
INSERT INTO t_tenant (id, code, name, status, create_by, is_deleted, create_time, update_time)
VALUES (sqlc.arg('id'), sqlc.arg('code'), sqlc.arg('name'), sqlc.arg('status'),
        sqlc.arg('create_by'), sqlc.arg('is_deleted'), sqlc.arg('now'), sqlc.arg('now'));

-- name: UpdateTenant :exec
UPDATE t_tenant SET code = sqlc.arg('code'), name = sqlc.arg('name'), status = sqlc.arg('status'),
    update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: SoftDeleteTenant :exec
UPDATE t_tenant SET is_deleted = 1, update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');
