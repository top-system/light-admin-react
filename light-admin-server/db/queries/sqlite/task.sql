-- task.sql (SQLite dialect)
-- Derived from queries/postgres/task.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). These queries intentionally do NOT
-- filter deleted_at, and Delete is a hard DELETE (see the postgres header).

-- name: GetTask :one
SELECT * FROM sys_tasks WHERE id = sqlc.arg('id');

-- name: ListTasks :many
SELECT * FROM sys_tasks
WHERE (sqlc.narg('type')           IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('status')         IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('correlation_id') IS NULL OR correlation_id = sqlc.narg('correlation_id'))
  AND (sqlc.narg('keywords')       IS NULL OR (
        type LIKE sqlc.narg('keywords') OR correlation_id LIKE sqlc.narg('keywords') OR public_error LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from') IS NULL OR created_at >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')   IS NULL OR created_at <= sqlc.narg('create_to'))
ORDER BY created_at DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountTasks :one
SELECT COUNT(*) FROM sys_tasks
WHERE (sqlc.narg('type')           IS NULL OR type = sqlc.narg('type'))
  AND (sqlc.narg('status')         IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('correlation_id') IS NULL OR correlation_id = sqlc.narg('correlation_id'))
  AND (sqlc.narg('keywords')       IS NULL OR (
        type LIKE sqlc.narg('keywords') OR correlation_id LIKE sqlc.narg('keywords') OR public_error LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from') IS NULL OR created_at >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')   IS NULL OR created_at <= sqlc.narg('create_to'));

-- name: DeleteTask :exec
DELETE FROM sys_tasks WHERE id = sqlc.arg('id');

-- name: BatchDeleteTasks :exec
DELETE FROM sys_tasks WHERE id IN (sqlc.slice('ids'));

-- name: ListTaskTypes :many
SELECT DISTINCT type FROM sys_tasks;

-- name: ListTaskStatusCounts :many
SELECT status, COUNT(*) AS count FROM sys_tasks GROUP BY status;
