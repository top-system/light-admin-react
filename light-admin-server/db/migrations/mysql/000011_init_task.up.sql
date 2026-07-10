-- 000011_init_task.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000011_init_task.up.sql. See that file for
-- the shared-table semantics (pkg/queue soft-deletes via deleted_at; the admin
-- Job module reads/deletes ignoring it). BIGSERIAL maps to BIGINT
-- AUTO_INCREMENT, which also backs the :execlastid + re-read RETURNING
-- replacement in mysqlstore.

CREATE TABLE IF NOT EXISTS sys_tasks (
    id                       BIGINT        NOT NULL AUTO_INCREMENT,
    type                     VARCHAR(100)  NOT NULL DEFAULT '',
    status                   VARCHAR(50)   NOT NULL DEFAULT '',
    correlation_id           VARCHAR(36)   NOT NULL DEFAULT '',
    owner_id                 CHAR(32)      NOT NULL DEFAULT '',
    private_state            TEXT          NOT NULL DEFAULT ('') ,
    public_retry_count       INT           NOT NULL DEFAULT 0,
    public_executed_duration BIGINT        NOT NULL DEFAULT 0,
    public_error             TEXT          NOT NULL DEFAULT ('') ,
    public_error_history     TEXT          NOT NULL DEFAULT ('') ,
    public_resume_time       BIGINT        NOT NULL DEFAULT 0,
    created_at               DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at               DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    deleted_at               DATETIME(6)   NULL,
    PRIMARY KEY (id),
    KEY idx_task_type           (type),
    KEY idx_task_status         (status),
    KEY idx_task_correlation_id (correlation_id),
    KEY idx_task_owner_id       (owner_id),
    KEY idx_task_deleted_at     (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
