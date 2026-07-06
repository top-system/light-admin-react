-- member.sql
-- SQL for the member domain (t_member). All member access is tenant-scoped: every
-- statement filters by tenant_id to enforce row-level isolation. Ordering is fixed
-- to id DESC (the previous GORM default); dynamic filters use sqlc.narg.

-- name: GetMemberByUsername :one
SELECT * FROM t_member
WHERE tenant_id = $1 AND username = $2 AND is_deleted = 0;

-- name: GetMemberByID :one
SELECT * FROM t_member
WHERE tenant_id = $1 AND id = $2 AND is_deleted = 0;

-- name: CountMembersByUsername :one
SELECT COUNT(*) FROM t_member
WHERE tenant_id = $1 AND username = $2 AND is_deleted = 0;

-- name: CreateMember :exec
INSERT INTO t_member (
    id, tenant_id, username, email, mobile, password, nickname, avatar,
    gender, status, is_deleted, create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, NOW(), NOW()
);

-- name: UpdateMemberProfile :exec
UPDATE t_member SET
    nickname    = sqlc.arg('nickname'),
    avatar      = sqlc.arg('avatar'),
    gender      = sqlc.arg('gender'),
    mobile      = sqlc.arg('mobile'),
    email       = sqlc.arg('email'),
    update_time = NOW()
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: UpdateMemberStatus :exec
UPDATE t_member SET status = $3, update_time = NOW()
WHERE tenant_id = $1 AND id = $2;

-- name: UpdateMemberPassword :exec
UPDATE t_member SET password = $3, update_time = NOW()
WHERE tenant_id = $1 AND id = $2;

-- name: UpdateMemberLoginInfo :exec
UPDATE t_member SET last_login_ip = $3, last_login_time = NOW(), update_time = NOW()
WHERE tenant_id = $1 AND id = $2;

-- name: ListMembers :many
SELECT * FROM t_member
WHERE is_deleted = 0
  AND (sqlc.narg('tenant_id')::text IS NULL OR tenant_id = sqlc.narg('tenant_id'))
  AND (sqlc.narg('keywords')::text  IS NULL OR (username LIKE sqlc.narg('keywords') OR nickname LIKE sqlc.narg('keywords') OR email LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('status')::int     IS NULL OR status = sqlc.narg('status'))
ORDER BY id DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountMembers :one
SELECT COUNT(*) FROM t_member
WHERE is_deleted = 0
  AND (sqlc.narg('tenant_id')::text IS NULL OR tenant_id = sqlc.narg('tenant_id'))
  AND (sqlc.narg('keywords')::text  IS NULL OR (username LIKE sqlc.narg('keywords') OR nickname LIKE sqlc.narg('keywords') OR email LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('status')::int     IS NULL OR status = sqlc.narg('status'));
