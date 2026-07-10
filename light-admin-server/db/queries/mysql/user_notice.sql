-- user_notice.sql (MySQL dialect)
-- Derived from queries/postgres/user_notice.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). MySQL has no unnest(): the postgres
-- BatchCreateUserNotices bulk insert is implemented in the mysqlstore adapter
-- as a loop over CreateUserNotice.

-- name: ListMyNotices :many
-- Published, non-deleted notices addressed to a user, newest first. Projects the
-- fields the UI needs (UserNoticePageVO).
SELECT un.id, un.notice_id, n.title, n.type, n.level, n.publish_time, un.is_read
FROM t_user_notice un
LEFT JOIN t_notice n ON un.notice_id = n.id
WHERE un.user_id = sqlc.arg('user_id')
  AND un.is_deleted = 0
  AND n.is_deleted = 0
  AND n.publish_status = 1
  AND (sqlc.narg('title') IS NULL OR n.title LIKE sqlc.narg('title'))
  AND (sqlc.narg('type')  IS NULL OR n.type = sqlc.narg('type'))
ORDER BY n.publish_time DESC
LIMIT  ?
OFFSET ?;

-- name: CountMyNotices :one
SELECT COUNT(*)
FROM t_user_notice un
LEFT JOIN t_notice n ON un.notice_id = n.id
WHERE un.user_id = sqlc.arg('user_id')
  AND un.is_deleted = 0
  AND n.is_deleted = 0
  AND n.publish_status = 1
  AND (sqlc.narg('title') IS NULL OR n.title LIKE sqlc.narg('title'))
  AND (sqlc.narg('type')  IS NULL OR n.type = sqlc.narg('type'));

-- name: CreateUserNotice :exec
INSERT INTO t_user_notice (id, notice_id, user_id, is_read)
VALUES (sqlc.arg('id'), sqlc.arg('notice_id'), sqlc.arg('user_id'), sqlc.arg('is_read'));

-- name: MarkUserNoticeRead :exec
UPDATE t_user_notice SET is_read = 1, read_time = sqlc.narg('now'), update_time = sqlc.narg('now')
WHERE notice_id = sqlc.arg('notice_id') AND user_id = sqlc.arg('user_id') AND is_read = 0;

-- name: MarkAllUserNoticesRead :exec
UPDATE t_user_notice SET is_read = 1, read_time = sqlc.narg('now'), update_time = sqlc.narg('now')
WHERE user_id = sqlc.arg('user_id') AND is_read = 0;

-- name: DeleteUserNoticesByNoticeID :exec
DELETE FROM t_user_notice WHERE notice_id = sqlc.arg('notice_id');

-- name: DeleteUserNoticesByNoticeIDs :exec
DELETE FROM t_user_notice WHERE notice_id IN (sqlc.slice('notice_ids'));
