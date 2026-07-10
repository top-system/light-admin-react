-- download.sql (MySQL dialect)
-- Derived from queries/postgres/download.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). MySQL has no RETURNING: the insert
-- uses :execlastid, which yields the AUTO_INCREMENT id directly.

-- name: GetDownloadTask :one
SELECT * FROM sys_download_tasks WHERE id = sqlc.arg('id');

-- name: GetDownloadTaskByTaskID :one
SELECT * FROM sys_download_tasks WHERE task_id = sqlc.arg('task_id');

-- name: GetDownloadTaskByQueueTaskID :one
SELECT * FROM sys_download_tasks WHERE queue_task_id = sqlc.arg('queue_task_id');

-- name: ListDownloadTasks :many
SELECT * FROM sys_download_tasks
WHERE (sqlc.narg('status')     IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('downloader') IS NULL OR downloader = sqlc.narg('downloader'))
  AND (sqlc.narg('keywords')   IS NULL OR (
        name LIKE sqlc.narg('keywords') OR url LIKE sqlc.narg('keywords') OR
        save_path LIKE sqlc.narg('keywords') OR hash LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from') IS NULL OR created_at >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')   IS NULL OR created_at <= sqlc.narg('create_to'))
ORDER BY created_at DESC
LIMIT  ?
OFFSET ?;

-- name: CountDownloadTasks :one
SELECT COUNT(*) FROM sys_download_tasks
WHERE (sqlc.narg('status')     IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('downloader') IS NULL OR downloader = sqlc.narg('downloader'))
  AND (sqlc.narg('keywords')   IS NULL OR (
        name LIKE sqlc.narg('keywords') OR url LIKE sqlc.narg('keywords') OR
        save_path LIKE sqlc.narg('keywords') OR hash LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from') IS NULL OR created_at >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')   IS NULL OR created_at <= sqlc.narg('create_to'));

-- name: CreateDownloadTask :execlastid
INSERT INTO sys_download_tasks (
    queue_task_id, task_id, hash, name, url, downloader, status,
    total, downloaded, download_speed, uploaded, upload_speed,
    save_path, error_message, owner_id, created_at, updated_at
) VALUES (
    sqlc.arg('queue_task_id'), sqlc.arg('task_id'), sqlc.arg('hash'), sqlc.arg('name'),
    sqlc.arg('url'), sqlc.arg('downloader'), sqlc.arg('status'),
    sqlc.arg('total'), sqlc.arg('downloaded'), sqlc.arg('download_speed'),
    sqlc.arg('uploaded'), sqlc.arg('upload_speed'),
    sqlc.arg('save_path'), sqlc.arg('error_message'), sqlc.arg('owner_id'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateDownloadTask :exec
UPDATE sys_download_tasks SET
    queue_task_id  = sqlc.arg('queue_task_id'),
    task_id        = sqlc.arg('task_id'),
    hash           = sqlc.arg('hash'),
    name           = sqlc.arg('name'),
    url            = sqlc.arg('url'),
    downloader     = sqlc.arg('downloader'),
    status         = sqlc.arg('status'),
    total          = sqlc.arg('total'),
    downloaded     = sqlc.arg('downloaded'),
    download_speed = sqlc.arg('download_speed'),
    uploaded       = sqlc.arg('uploaded'),
    upload_speed   = sqlc.arg('upload_speed'),
    save_path      = sqlc.arg('save_path'),
    error_message  = sqlc.arg('error_message'),
    owner_id       = sqlc.arg('owner_id'),
    updated_at     = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: UpdateDownloadTaskStatus :exec
UPDATE sys_download_tasks SET
    status         = sqlc.arg('status'),
    downloaded     = sqlc.arg('downloaded'),
    total          = sqlc.arg('total'),
    download_speed = sqlc.arg('download_speed'),
    uploaded       = sqlc.arg('uploaded'),
    upload_speed   = sqlc.arg('upload_speed'),
    error_message  = sqlc.arg('error_message'),
    updated_at     = sqlc.arg('now')
WHERE id = sqlc.arg('id');

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
DELETE FROM sys_download_tasks WHERE id = sqlc.arg('id');

-- name: BatchDeleteDownloadTasks :exec
DELETE FROM sys_download_tasks WHERE id IN (sqlc.slice('ids'));

-- name: ListDownloadStatusCounts :many
SELECT status, COUNT(*) AS count FROM sys_download_tasks GROUP BY status;

-- name: ListActiveDownloadTasks :many
SELECT id, queue_task_id, task_id, hash, downloader FROM sys_download_tasks
WHERE status IN ('downloading', 'seeding', 'unknown', 'queued');
