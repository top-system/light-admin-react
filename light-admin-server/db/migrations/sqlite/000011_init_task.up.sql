-- 000011_init_task.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000011_init_task.up.sql. See that file for
-- the shared-table semantics (pkg/queue soft-deletes via deleted_at; the admin
-- Job module reads/deletes ignoring it). BIGSERIAL maps to INTEGER PRIMARY KEY
-- AUTOINCREMENT (Go int64).

CREATE TABLE IF NOT EXISTS sys_tasks (
    id                       INTEGER       PRIMARY KEY AUTOINCREMENT,
    type                     VARCHAR(100)  NOT NULL DEFAULT '',
    status                   VARCHAR(50)   NOT NULL DEFAULT '',
    correlation_id           VARCHAR(36)   NOT NULL DEFAULT '',
    owner_id                 VARCHAR(32)   NOT NULL DEFAULT '',
    private_state            TEXT          NOT NULL DEFAULT '',
    public_retry_count       INTEGER       NOT NULL DEFAULT 0,
    public_executed_duration BIGINT        NOT NULL DEFAULT 0,
    public_error             TEXT          NOT NULL DEFAULT '',
    public_error_history     TEXT          NOT NULL DEFAULT '',
    public_resume_time       BIGINT        NOT NULL DEFAULT 0,
    created_at               DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at               DATETIME
);

CREATE INDEX IF NOT EXISTS idx_task_type           ON sys_tasks (type);
CREATE INDEX IF NOT EXISTS idx_task_status         ON sys_tasks (status);
CREATE INDEX IF NOT EXISTS idx_task_correlation_id ON sys_tasks (correlation_id);
CREATE INDEX IF NOT EXISTS idx_task_owner_id       ON sys_tasks (owner_id);
CREATE INDEX IF NOT EXISTS idx_task_deleted_at     ON sys_tasks (deleted_at);
