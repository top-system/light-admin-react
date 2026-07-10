-- queue_task.sql (MySQL dialect)
-- Derived from queries/postgres/queue_task.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). MySQL has no RETURNING: Create uses
-- :execlastid and Update plain :exec; the mysqlstore adapter re-reads the row
-- on the same connection/transaction. The nullable type-list filter uses the
-- FIND_IN_SET membership test (see menu.sql header).

-- name: CreateQueueTask :execlastid
INSERT INTO sys_tasks (
    type, status, correlation_id, owner_id, private_state,
    public_retry_count, public_executed_duration, public_error,
    public_error_history, public_resume_time, created_at, updated_at
) VALUES (
    sqlc.arg('type'), sqlc.arg('status'), sqlc.arg('correlation_id'), sqlc.arg('owner_id'),
    sqlc.arg('private_state'), sqlc.arg('public_retry_count'), sqlc.arg('public_executed_duration'),
    sqlc.arg('public_error'), sqlc.arg('public_error_history'), sqlc.arg('public_resume_time'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateQueueTask :exec
UPDATE sys_tasks SET
    type                     = sqlc.arg('type'),
    status                   = sqlc.arg('status'),
    correlation_id           = sqlc.arg('correlation_id'),
    owner_id                 = sqlc.arg('owner_id'),
    private_state            = sqlc.arg('private_state'),
    public_retry_count       = sqlc.arg('public_retry_count'),
    public_executed_duration = sqlc.arg('public_executed_duration'),
    public_error             = sqlc.arg('public_error'),
    public_error_history     = sqlc.arg('public_error_history'),
    public_resume_time       = sqlc.arg('public_resume_time'),
    updated_at               = sqlc.arg('now')
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetQueueTask :one
SELECT * FROM sys_tasks WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListPendingQueueTasks :many
SELECT * FROM sys_tasks
WHERE deleted_at IS NULL
  AND status IN ('queued', 'processing', 'suspending')
  AND (sqlc.narg('types') IS NULL OR FIND_IN_SET(type, sqlc.narg('types')))
ORDER BY id;

-- name: SoftDeleteQueueTask :exec
UPDATE sys_tasks SET deleted_at = sqlc.arg('now') WHERE id = sqlc.arg('id') AND deleted_at IS NULL;
