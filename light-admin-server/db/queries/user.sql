-- user.sql
-- All SQL for the user domain. Generated into db/sqlc by `make sqlc`.
-- Dynamic filters use sqlc.narg (nullable args): a NULL argument disables that
-- predicate, so a single prepared statement serves every filter combination
-- while remaining fully parameterized (no SQL injection surface).

-- name: GetUser :one
SELECT * FROM t_user
WHERE id = $1 AND is_deleted = 0;

-- name: GetUserByUsername :one
SELECT * FROM t_user
WHERE username = $1 AND is_deleted = 0
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM t_user
WHERE is_deleted = 0
  AND (sqlc.narg('username')::text  IS NULL OR username = sqlc.narg('username'))
  AND (sqlc.narg('nickname')::text  IS NULL OR nickname = sqlc.narg('nickname'))
  AND (sqlc.narg('status')::int     IS NULL OR status   = sqlc.narg('status'))
  AND (sqlc.narg('dept_id')::text   IS NULL OR dept_id  = sqlc.narg('dept_id'))
  AND (sqlc.narg('keywords')::text  IS NULL OR (
        username ILIKE sqlc.narg('keywords') OR
        nickname ILIKE sqlc.narg('keywords') OR
        mobile   ILIKE sqlc.narg('keywords') OR
        email    ILIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from')::timestamptz IS NULL OR create_time >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')::timestamptz   IS NULL OR create_time <= sqlc.narg('create_to'))
  AND (sqlc.narg('role_ids')::text[] IS NULL OR id IN (
        SELECT user_id FROM t_user_role WHERE role_id = ANY(sqlc.narg('role_ids')::text[])))
ORDER BY create_time DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountUsers :one
SELECT COUNT(*) FROM t_user
WHERE is_deleted = 0
  AND (sqlc.narg('username')::text  IS NULL OR username = sqlc.narg('username'))
  AND (sqlc.narg('nickname')::text  IS NULL OR nickname = sqlc.narg('nickname'))
  AND (sqlc.narg('status')::int     IS NULL OR status   = sqlc.narg('status'))
  AND (sqlc.narg('dept_id')::text   IS NULL OR dept_id  = sqlc.narg('dept_id'))
  AND (sqlc.narg('keywords')::text  IS NULL OR (
        username ILIKE sqlc.narg('keywords') OR
        nickname ILIKE sqlc.narg('keywords') OR
        mobile   ILIKE sqlc.narg('keywords') OR
        email    ILIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from')::timestamptz IS NULL OR create_time >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')::timestamptz   IS NULL OR create_time <= sqlc.narg('create_to'))
  AND (sqlc.narg('role_ids')::text[] IS NULL OR id IN (
        SELECT user_id FROM t_user_role WHERE role_id = ANY(sqlc.narg('role_ids')::text[])));

-- name: CreateUser :exec
INSERT INTO t_user (
    id, username, nickname, gender, password, dept_id, avatar,
    mobile, status, email, create_by, update_by, is_deleted, openid,
    create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13, $14,
    NOW(), NOW()
);

-- name: UpdateUser :exec
UPDATE t_user SET
    username    = $2,
    nickname    = $3,
    gender      = $4,
    dept_id     = $5,
    avatar      = $6,
    mobile      = $7,
    status      = $8,
    email       = $9,
    update_by   = $10,
    update_time = NOW()
WHERE id = $1;

-- name: UpdateUserProfile :exec
UPDATE t_user SET
    nickname    = COALESCE(sqlc.narg('nickname'), nickname),
    gender      = COALESCE(sqlc.narg('gender'),   gender),
    avatar      = COALESCE(sqlc.narg('avatar'),   avatar),
    mobile      = COALESCE(sqlc.narg('mobile'),   mobile),
    email       = COALESCE(sqlc.narg('email'),    email),
    update_time = NOW()
WHERE id = sqlc.arg('id');

-- name: UpdateUserStatus :exec
UPDATE t_user SET status = $2, update_time = NOW() WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE t_user SET password = $2, update_time = NOW() WHERE id = $1;

-- name: SoftDeleteUser :exec
UPDATE t_user SET is_deleted = 1, update_time = NOW() WHERE id = $1;
