-- queue_task.sql
-- SQL for the framework-agnostic pkg/queue engine over sys_tasks. Unlike the admin
-- Job module (task.sql), these queries honour soft-delete: reads filter
-- `deleted_at IS NULL`, Create/Update touch only live rows, and Delete is a soft
-- delete (sets deleted_at) — preserving the previous gorm.DeletedAt behaviour of
-- the engine's GORM TaskRepository. The sqlc implementation of queue.TaskRepository
-- lives in the application layer (lib) and is injected into pkg/queue.

-- name: CreateQueueTask :one
INSERT INTO sys_tasks (
    type, status, correlation_id, owner_id, private_state,
    public_retry_count, public_executed_duration, public_error,
    public_error_history, public_resume_time, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, sqlc.arg('now'), sqlc.arg('now')
)
RETURNING *;

-- name: UpdateQueueTask :one
UPDATE sys_tasks SET
    type                     = $2,
    status                   = $3,
    correlation_id           = $4,
    owner_id                 = $5,
    private_state            = $6,
    public_retry_count       = $7,
    public_executed_duration = $8,
    public_error             = $9,
    public_error_history     = $10,
    public_resume_time       = $11,
    updated_at               = sqlc.arg('now')
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: GetQueueTask :one
SELECT * FROM sys_tasks WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPendingQueueTasks :many
SELECT * FROM sys_tasks
WHERE deleted_at IS NULL
  AND status IN ('queued', 'processing', 'suspending')
  AND (sqlc.narg('types')::text[] IS NULL OR type = ANY(sqlc.narg('types')::text[]))
ORDER BY id;

-- name: SoftDeleteQueueTask :exec
UPDATE sys_tasks SET deleted_at = sqlc.arg('now')::timestamptz WHERE id = $1 AND deleted_at IS NULL;
