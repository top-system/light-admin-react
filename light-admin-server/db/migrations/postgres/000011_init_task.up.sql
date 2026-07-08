-- 000011_init_task.up.sql
-- Task queue table (sys_tasks). This table is shared: the framework-agnostic
-- pkg/queue engine persists task lifecycle state through its own TaskRepository
-- interface (GORM implementation, which soft-deletes via deleted_at), while the
-- admin Job module reads/deletes it through the sqlc data layer. Moving the schema
-- to golang-migrate lets sqlc generate against it; the queue engine keeps writing
-- to the same table. `IF NOT EXISTS` keeps the migration idempotent on databases
-- where GORM AutoMigrate already created it.
--
-- The embedded public state (pkg/queue TaskPublicState) is stored with a public_
-- column prefix. deleted_at is a nullable gorm.DeletedAt column used by the engine;
-- the admin side intentionally neither filters nor sets it (models/system.Task has
-- no DeletedAt field), preserving the previous hard-delete admin behaviour.

CREATE TABLE IF NOT EXISTS sys_tasks (
    id                       BIGSERIAL     NOT NULL,
    type                     VARCHAR(100)  NOT NULL DEFAULT '',
    status                   VARCHAR(50)   NOT NULL DEFAULT '',
    correlation_id           VARCHAR(36)   NOT NULL DEFAULT '',
    owner_id                 CHAR(32)      NOT NULL DEFAULT '',
    private_state            TEXT          NOT NULL DEFAULT '',
    public_retry_count       INTEGER       NOT NULL DEFAULT 0,
    public_executed_duration BIGINT        NOT NULL DEFAULT 0,
    public_error             TEXT          NOT NULL DEFAULT '',
    public_error_history     TEXT          NOT NULL DEFAULT '',
    public_resume_time       BIGINT        NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ,
    CONSTRAINT pk_sys_tasks PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_task_type           ON sys_tasks (type);
CREATE INDEX IF NOT EXISTS idx_task_status         ON sys_tasks (status);
CREATE INDEX IF NOT EXISTS idx_task_correlation_id ON sys_tasks (correlation_id);
CREATE INDEX IF NOT EXISTS idx_task_owner_id       ON sys_tasks (owner_id);
CREATE INDEX IF NOT EXISTS idx_task_deleted_at     ON sys_tasks (deleted_at);
