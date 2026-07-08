-- user_notice.sql
-- SQL for per-user notice state (t_user_notice), including the "my notices" join.

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
  AND (sqlc.narg('title')::text IS NULL OR n.title LIKE sqlc.narg('title'))
  AND (sqlc.narg('type')::int   IS NULL OR n.type = sqlc.narg('type'))
ORDER BY n.publish_time DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountMyNotices :one
SELECT COUNT(*)
FROM t_user_notice un
LEFT JOIN t_notice n ON un.notice_id = n.id
WHERE un.user_id = sqlc.arg('user_id')
  AND un.is_deleted = 0
  AND n.is_deleted = 0
  AND n.publish_status = 1
  AND (sqlc.narg('title')::text IS NULL OR n.title LIKE sqlc.narg('title'))
  AND (sqlc.narg('type')::int   IS NULL OR n.type = sqlc.narg('type'));

-- name: BatchCreateUserNotices :exec
-- Single-statement bulk insert via array unnest.
INSERT INTO t_user_notice (id, notice_id, user_id, is_read)
SELECT unnest(@ids::text[]), unnest(@notice_ids::text[]), unnest(@user_ids::text[]), 0;

-- name: CreateUserNotice :exec
INSERT INTO t_user_notice (id, notice_id, user_id, is_read) VALUES ($1, $2, $3, $4);

-- name: MarkUserNoticeRead :exec
UPDATE t_user_notice SET is_read = 1, read_time = sqlc.arg('now')::timestamptz, update_time = sqlc.arg('now')::timestamptz
WHERE notice_id = $1 AND user_id = $2 AND is_read = 0;

-- name: MarkAllUserNoticesRead :exec
UPDATE t_user_notice SET is_read = 1, read_time = sqlc.arg('now')::timestamptz, update_time = sqlc.arg('now')::timestamptz
WHERE user_id = $1 AND is_read = 0;

-- name: DeleteUserNoticesByNoticeID :exec
DELETE FROM t_user_notice WHERE notice_id = $1;

-- name: DeleteUserNoticesByNoticeIDs :exec
DELETE FROM t_user_notice WHERE notice_id = ANY(@notice_ids::text[]);
