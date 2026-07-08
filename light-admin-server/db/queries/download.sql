-- download.sql
-- SQL for download tasks (sys_download_tasks). Dynamic filters use sqlc.narg;
-- date-range bounds are passed as timestamptz. Ordering is fixed to created_at
-- DESC — the previous GORM default.

-- name: GetDownloadTask :one
SELECT * FROM sys_download_tasks WHERE id = $1;

-- name: GetDownloadTaskByTaskID :one
SELECT * FROM sys_download_tasks WHERE task_id = $1;

-- name: GetDownloadTaskByQueueTaskID :one
SELECT * FROM sys_download_tasks WHERE queue_task_id = $1;

-- name: ListDownloadTasks :many
SELECT * FROM sys_download_tasks
WHERE (sqlc.narg('status')::text     IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('downloader')::text IS NULL OR downloader = sqlc.narg('downloader'))
  AND (sqlc.narg('keywords')::text   IS NULL OR (
        name LIKE sqlc.narg('keywords') OR url LIKE sqlc.narg('keywords') OR
        save_path LIKE sqlc.narg('keywords') OR hash LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from')::timestamptz IS NULL OR created_at >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')::timestamptz   IS NULL OR created_at <= sqlc.narg('create_to'))
ORDER BY created_at DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountDownloadTasks :one
SELECT COUNT(*) FROM sys_download_tasks
WHERE (sqlc.narg('status')::text     IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('downloader')::text IS NULL OR downloader = sqlc.narg('downloader'))
  AND (sqlc.narg('keywords')::text   IS NULL OR (
        name LIKE sqlc.narg('keywords') OR url LIKE sqlc.narg('keywords') OR
        save_path LIKE sqlc.narg('keywords') OR hash LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from')::timestamptz IS NULL OR created_at >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')::timestamptz   IS NULL OR created_at <= sqlc.narg('create_to'));

-- name: CreateDownloadTask :one
INSERT INTO sys_download_tasks (
    queue_task_id, task_id, hash, name, url, downloader, status,
    total, downloaded, download_speed, uploaded, upload_speed,
    save_path, error_message, owner_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12,
    $13, $14, $15, sqlc.arg('now'), sqlc.arg('now')
)
RETURNING id;

-- name: UpdateDownloadTask :exec
UPDATE sys_download_tasks SET
    queue_task_id  = $2,
    task_id        = $3,
    hash           = $4,
    name           = $5,
    url            = $6,
    downloader     = $7,
    status         = $8,
    total          = $9,
    downloaded     = $10,
    download_speed = $11,
    uploaded       = $12,
    upload_speed   = $13,
    save_path      = $14,
    error_message  = $15,
    owner_id       = $16,
    updated_at     = sqlc.arg('now')
WHERE id = $1;

-- name: UpdateDownloadTaskStatus :exec
UPDATE sys_download_tasks SET
    status         = $2,
    downloaded     = $3,
    total          = $4,
    download_speed = $5,
    uploaded       = $6,
    upload_speed   = $7,
    error_message  = $8,
    updated_at     = sqlc.arg('now')
WHERE id = $1;

-- name: UpdateDownloadTaskFromDownloader :exec
-- Updates progress and, only when a non-empty value is supplied, the identity
-- fields (task_id, hash, name, save_path). NULLIF(arg,'') + COALESCE keeps the
-- existing value for empty arguments, mirroring the previous conditional map.
UPDATE sys_download_tasks SET
    status         = sqlc.arg('status'),
    downloaded     = sqlc.arg('downloaded'),
    total          = sqlc.arg('total'),
    download_speed = sqlc.arg('download_speed'),
    uploaded       = sqlc.arg('uploaded'),
    upload_speed   = sqlc.arg('upload_speed'),
    error_message  = sqlc.arg('error_message'),
    task_id        = COALESCE(NULLIF(sqlc.arg('task_id'), ''), task_id),
    hash           = COALESCE(NULLIF(sqlc.arg('hash'), ''), hash),
    name           = COALESCE(NULLIF(sqlc.arg('name'), ''), name),
    save_path      = COALESCE(NULLIF(sqlc.arg('save_path'), ''), save_path),
    updated_at     = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: DeleteDownloadTask :exec
DELETE FROM sys_download_tasks WHERE id = $1;

-- name: BatchDeleteDownloadTasks :exec
DELETE FROM sys_download_tasks WHERE id = ANY(@ids::bigint[]);

-- name: ListDownloadStatusCounts :many
SELECT status, COUNT(*) AS count FROM sys_download_tasks GROUP BY status;

-- name: ListActiveDownloadTasks :many
SELECT id, queue_task_id, task_id, hash, downloader FROM sys_download_tasks
WHERE status = ANY(ARRAY['downloading', 'seeding', 'unknown', 'queued']);
