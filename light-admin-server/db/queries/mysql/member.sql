-- member.sql (MySQL dialect)
-- Derived from queries/postgres/member.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). All member access is tenant-scoped.

-- name: GetMemberByUsername :one
SELECT * FROM t_member
WHERE tenant_id = sqlc.arg('tenant_id') AND username = sqlc.arg('username') AND is_deleted = 0;

-- name: GetMemberByID :one
SELECT * FROM t_member
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id') AND is_deleted = 0;

-- name: CountMembersByUsername :one
SELECT COUNT(*) FROM t_member
WHERE tenant_id = sqlc.arg('tenant_id') AND username = sqlc.arg('username') AND is_deleted = 0;

-- name: CreateMember :exec
INSERT INTO t_member (
    id, tenant_id, username, email, mobile, password, nickname, avatar,
    gender, status, is_deleted, create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('tenant_id'), sqlc.arg('username'), sqlc.arg('email'),
    sqlc.arg('mobile'), sqlc.arg('password'), sqlc.arg('nickname'), sqlc.arg('avatar'),
    sqlc.arg('gender'), sqlc.arg('status'), sqlc.arg('is_deleted'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateMemberProfile :exec
UPDATE t_member SET
    nickname    = sqlc.arg('nickname'),
    avatar      = sqlc.arg('avatar'),
    gender      = sqlc.arg('gender'),
    mobile      = sqlc.arg('mobile'),
    email       = sqlc.arg('email'),
    update_time = sqlc.arg('now')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: UpdateMemberStatus :exec
UPDATE t_member SET status = sqlc.arg('status'), update_time = sqlc.arg('now')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: UpdateMemberPassword :exec
UPDATE t_member SET password = sqlc.arg('password'), update_time = sqlc.arg('now')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: UpdateMemberLoginInfo :exec
UPDATE t_member SET last_login_ip = sqlc.arg('last_login_ip'), last_login_time = sqlc.narg('now'), update_time = sqlc.narg('now')
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id');

-- name: ListMembers :many
SELECT * FROM t_member
WHERE is_deleted = 0
  AND (sqlc.narg('tenant_id') IS NULL OR tenant_id = sqlc.narg('tenant_id'))
  AND (sqlc.narg('keywords')  IS NULL OR (username LIKE sqlc.narg('keywords') OR nickname LIKE sqlc.narg('keywords') OR email LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('status')    IS NULL OR status = sqlc.narg('status'))
ORDER BY id DESC
LIMIT  ?
OFFSET ?;

-- name: CountMembers :one
SELECT COUNT(*) FROM t_member
WHERE is_deleted = 0
  AND (sqlc.narg('tenant_id') IS NULL OR tenant_id = sqlc.narg('tenant_id'))
  AND (sqlc.narg('keywords')  IS NULL OR (username LIKE sqlc.narg('keywords') OR nickname LIKE sqlc.narg('keywords') OR email LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('status')    IS NULL OR status = sqlc.narg('status'));
