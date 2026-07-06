-- task.sql
-- SQL for the admin Job module over sys_tasks. These queries intentionally do NOT
-- filter deleted_at (models/system.Task has no DeletedAt field), and Delete is a
-- hard DELETE — preserving the previous admin behaviour, which differs from the
-- pkg/queue engine's soft-delete view of the same table. Ordering is fixed to
-- created_at DESC; dynamic filters use sqlc.narg.

-- name: GetTask :one
SELECT * FROM sys_tasks WHERE id = $1;

-- name: ListTasks :many
SELECT * FROM sys_tasks
WHERE (sqlc.narg('type')::text           IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('status')::text         IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('correlation_id')::text IS NULL OR correlation_id = sqlc.narg('correlation_id'))
  AND (sqlc.narg('keywords')::text       IS NULL OR (
        type LIKE sqlc.narg('keywords') OR correlation_id LIKE sqlc.narg('keywords') OR public_error LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from')::timestamptz IS NULL OR created_at >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')::timestamptz   IS NULL OR created_at <= sqlc.narg('create_to'))
ORDER BY created_at DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountTasks :one
SELECT COUNT(*) FROM sys_tasks
WHERE (sqlc.narg('type')::text           IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('status')::text         IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('correlation_id')::text IS NULL OR correlation_id = sqlc.narg('correlation_id'))
  AND (sqlc.narg('keywords')::text       IS NULL OR (
        type LIKE sqlc.narg('keywords') OR correlation_id LIKE sqlc.narg('keywords') OR public_error LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from')::timestamptz IS NULL OR created_at >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')::timestamptz   IS NULL OR created_at <= sqlc.narg('create_to'));

-- name: DeleteTask :exec
DELETE FROM sys_tasks WHERE id = $1;

-- name: BatchDeleteTasks :exec
DELETE FROM sys_tasks WHERE id = ANY(@ids::bigint[]);

-- name: ListTaskTypes :many
SELECT DISTINCT type FROM sys_tasks;

-- name: ListTaskStatusCounts :many
SELECT status, COUNT(*) AS count FROM sys_tasks GROUP BY status;
