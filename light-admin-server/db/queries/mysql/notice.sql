-- notice.sql (MySQL dialect)
-- Derived from queries/postgres/notice.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header).

-- name: GetNotice :one
SELECT * FROM t_notice WHERE id = sqlc.arg('id') AND is_deleted = 0;

-- name: ListNotices :many
SELECT * FROM t_notice
WHERE is_deleted = 0
  AND (sqlc.narg('title') IS NULL OR title LIKE sqlc.narg('title'))
  AND (sqlc.narg('type')  IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('publish_status') IS NULL OR publish_status = sqlc.narg('publish_status'))
ORDER BY create_time DESC
LIMIT  ?
OFFSET ?;

-- name: CountNotices :one
SELECT COUNT(*) FROM t_notice
WHERE is_deleted = 0
  AND (sqlc.narg('title') IS NULL OR title LIKE sqlc.narg('title'))
  AND (sqlc.narg('type')  IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('publish_status') IS NULL OR publish_status = sqlc.narg('publish_status'));

-- name: CreateNotice :exec
INSERT INTO t_notice (
    id, title, content, type, level, target_type, target_user_ids,
    publish_status, create_by, is_deleted, create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('title'), sqlc.arg('content'), sqlc.arg('type'),
    sqlc.arg('level'), sqlc.arg('target_type'), sqlc.arg('target_user_ids'),
    sqlc.arg('publish_status'), sqlc.arg('create_by'), sqlc.arg('is_deleted'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateNotice :exec
UPDATE t_notice SET
    title           = sqlc.arg('title'),
    content         = sqlc.arg('content'),
    type            = sqlc.arg('type'),
    level           = sqlc.arg('level'),
    target_type     = sqlc.arg('target_type'),
    target_user_ids = sqlc.arg('target_user_ids'),
    update_by       = sqlc.arg('update_by'),
    update_time     = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: UpdateNoticeStatus :exec
UPDATE t_notice SET
    publish_status = sqlc.arg('publish_status'),
    publisher_id   = sqlc.arg('publisher_id'),
    publish_time   = COALESCE(sqlc.narg('publish_time'), publish_time),
    revoke_time    = COALESCE(sqlc.narg('revoke_time'), revoke_time),
    update_time    = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: SoftDeleteNoticesByIDs :exec
UPDATE t_notice SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id IN (sqlc.slice('ids'));
