-- notice.sql
-- SQL for notices (t_notice). Dynamic filters use sqlc.narg; ordering is fixed to
-- create_time DESC — the previous GORM default. publish_time / revoke_time are
-- set conditionally via COALESCE so a single statement handles publish and revoke.

-- name: GetNotice :one
SELECT * FROM t_notice WHERE id = $1 AND is_deleted = 0;

-- name: ListNotices :many
SELECT * FROM t_notice
WHERE is_deleted = 0
  AND (sqlc.narg('title')::text IS NULL OR title LIKE sqlc.narg('title'))
  AND (sqlc.narg('type')::int   IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('publish_status')::int IS NULL OR publish_status = sqlc.narg('publish_status'))
ORDER BY create_time DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountNotices :one
SELECT COUNT(*) FROM t_notice
WHERE is_deleted = 0
  AND (sqlc.narg('title')::text IS NULL OR title LIKE sqlc.narg('title'))
  AND (sqlc.narg('type')::int   IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('publish_status')::int IS NULL OR publish_status = sqlc.narg('publish_status'));

-- name: CreateNotice :exec
INSERT INTO t_notice (
    id, title, content, type, level, target_type, target_user_ids,
    publish_status, create_by, is_deleted, create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateNotice :exec
UPDATE t_notice SET
    title           = $2,
    content         = $3,
    type            = $4,
    level           = $5,
    target_type     = $6,
    target_user_ids = $7,
    update_by       = $8,
    update_time     = sqlc.arg('now')
WHERE id = $1;

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
WHERE id = ANY(sqlc.arg('ids')::text[]);
