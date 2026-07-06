-- tenant.sql
-- SQL for the tenant domain (t_tenant). Ordering is fixed to id DESC (the
-- previous GORM default); dynamic filters use sqlc.narg.

-- name: GetTenant :one
SELECT * FROM t_tenant WHERE id = $1 AND is_deleted = 0;

-- name: GetTenantByCode :one
SELECT * FROM t_tenant WHERE code = $1 AND is_deleted = 0;

-- name: ListTenants :many
SELECT * FROM t_tenant
WHERE is_deleted = 0
  AND (sqlc.narg('keywords')::text IS NULL OR (code LIKE sqlc.narg('keywords') OR name LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('status')::int    IS NULL OR status = sqlc.narg('status'))
ORDER BY id DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountTenants :one
SELECT COUNT(*) FROM t_tenant
WHERE is_deleted = 0
  AND (sqlc.narg('keywords')::text IS NULL OR (code LIKE sqlc.narg('keywords') OR name LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('status')::int    IS NULL OR status = sqlc.narg('status'));

-- name: CreateTenant :exec
INSERT INTO t_tenant (id, code, name, status, create_by, is_deleted, create_time, update_time)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW());

-- name: UpdateTenant :exec
UPDATE t_tenant SET code = $2, name = $3, status = $4, update_by = $5, update_time = NOW()
WHERE id = $1;

-- name: SoftDeleteTenant :exec
UPDATE t_tenant SET is_deleted = 1, update_time = NOW() WHERE id = $1;
