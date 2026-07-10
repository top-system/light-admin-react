-- user.sql (SQLite dialect)
-- Derived from queries/postgres/user.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). ILIKE becomes LIKE (SQLite LIKE is
-- already case-insensitive for ASCII). The nullable role-id list uses the
-- comma-wrapped instr() membership test (see menu.sql header).

-- name: GetUser :one
SELECT * FROM t_user
WHERE id = sqlc.arg('id') AND is_deleted = 0;

-- name: GetUserByUsername :one
SELECT * FROM t_user
WHERE username = sqlc.arg('username') AND is_deleted = 0
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM t_user
WHERE is_deleted = 0
  AND (sqlc.narg('username') IS NULL OR username = sqlc.narg('username'))
  AND (sqlc.narg('nickname') IS NULL OR nickname = sqlc.narg('nickname'))
  AND (sqlc.narg('status')   IS NULL OR status   = sqlc.narg('status'))
  AND (sqlc.narg('dept_id')  IS NULL OR dept_id  = sqlc.narg('dept_id'))
  AND (sqlc.narg('keywords') IS NULL OR (
        username LIKE sqlc.narg('keywords') OR
        nickname LIKE sqlc.narg('keywords') OR
        mobile   LIKE sqlc.narg('keywords') OR
        email    LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from') IS NULL OR create_time >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')   IS NULL OR create_time <= sqlc.narg('create_to'))
  AND (sqlc.narg('role_ids') IS NULL OR id IN (
        SELECT user_id FROM t_user_role WHERE instr(sqlc.narg('role_ids'), ',' || role_id || ',') > 0))
ORDER BY create_time DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountUsers :one
SELECT COUNT(*) FROM t_user
WHERE is_deleted = 0
  AND (sqlc.narg('username') IS NULL OR username = sqlc.narg('username'))
  AND (sqlc.narg('nickname') IS NULL OR nickname = sqlc.narg('nickname'))
  AND (sqlc.narg('status')   IS NULL OR status   = sqlc.narg('status'))
  AND (sqlc.narg('dept_id')  IS NULL OR dept_id  = sqlc.narg('dept_id'))
  AND (sqlc.narg('keywords') IS NULL OR (
        username LIKE sqlc.narg('keywords') OR
        nickname LIKE sqlc.narg('keywords') OR
        mobile   LIKE sqlc.narg('keywords') OR
        email    LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from') IS NULL OR create_time >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')   IS NULL OR create_time <= sqlc.narg('create_to'))
  AND (sqlc.narg('role_ids') IS NULL OR id IN (
        SELECT user_id FROM t_user_role WHERE instr(sqlc.narg('role_ids'), ',' || role_id || ',') > 0));

-- name: CreateUser :exec
INSERT INTO t_user (
    id, username, nickname, gender, password, dept_id, avatar,
    mobile, status, email, create_by, update_by, is_deleted, openid,
    create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('username'), sqlc.arg('nickname'), sqlc.arg('gender'),
    sqlc.arg('password'), sqlc.arg('dept_id'), sqlc.arg('avatar'),
    sqlc.arg('mobile'), sqlc.arg('status'), sqlc.arg('email'), sqlc.arg('create_by'),
    sqlc.arg('update_by'), sqlc.arg('is_deleted'), sqlc.arg('openid'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateUser :exec
UPDATE t_user SET
    username    = sqlc.arg('username'),
    nickname    = sqlc.arg('nickname'),
    gender      = sqlc.arg('gender'),
    dept_id     = sqlc.arg('dept_id'),
    avatar      = sqlc.arg('avatar'),
    mobile      = sqlc.arg('mobile'),
    status      = sqlc.arg('status'),
    email       = sqlc.arg('email'),
    update_by   = sqlc.arg('update_by'),
    update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: UpdateUserProfile :exec
UPDATE t_user SET
    nickname    = COALESCE(sqlc.narg('nickname'), nickname),
    gender      = COALESCE(sqlc.narg('gender'),   gender),
    avatar      = COALESCE(sqlc.narg('avatar'),   avatar),
    mobile      = COALESCE(sqlc.narg('mobile'),   mobile),
    email       = COALESCE(sqlc.narg('email'),    email),
    update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: UpdateUserStatus :exec
UPDATE t_user SET status = sqlc.arg('status'), update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');

-- name: UpdateUserPassword :exec
UPDATE t_user SET password = sqlc.arg('password'), update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');

-- name: SoftDeleteUser :exec
UPDATE t_user SET is_deleted = 1, update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');
