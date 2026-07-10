-- 000010_init_download.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000010_init_download.up.sql. BIGSERIAL maps
-- to INTEGER PRIMARY KEY AUTOINCREMENT (the rowid alias must be declared
-- INTEGER, not BIGINT; values are still Go int64). deleted_at is a plain
-- nullable column, so deletes are hard DELETEs.

CREATE TABLE IF NOT EXISTS sys_download_tasks (
    id             INTEGER       PRIMARY KEY AUTOINCREMENT,
    queue_task_id  BIGINT        NOT NULL DEFAULT 0,
    task_id        VARCHAR(100)  NOT NULL DEFAULT '',
    hash           VARCHAR(100)  NOT NULL DEFAULT '',
    name           VARCHAR(500)  NOT NULL DEFAULT '',
    url            TEXT          NOT NULL DEFAULT '',
    downloader     VARCHAR(50)   NOT NULL DEFAULT '',
    status         VARCHAR(50)   NOT NULL DEFAULT '',
    total          BIGINT        NOT NULL DEFAULT 0,
    downloaded     BIGINT        NOT NULL DEFAULT 0,
    download_speed BIGINT        NOT NULL DEFAULT 0,
    uploaded       BIGINT        NOT NULL DEFAULT 0,
    upload_speed   BIGINT        NOT NULL DEFAULT 0,
    save_path      VARCHAR(500)  NOT NULL DEFAULT '',
    error_message  TEXT          NOT NULL DEFAULT '',
    owner_id       VARCHAR(32)   NOT NULL DEFAULT '',
    created_at     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     DATETIME
);

CREATE INDEX IF NOT EXISTS idx_download_queue_task_id ON sys_download_tasks (queue_task_id);
CREATE INDEX IF NOT EXISTS idx_download_task_id       ON sys_download_tasks (task_id);
CREATE INDEX IF NOT EXISTS idx_download_hash          ON sys_download_tasks (hash);
CREATE INDEX IF NOT EXISTS idx_download_status        ON sys_download_tasks (status);
CREATE INDEX IF NOT EXISTS idx_download_owner_id      ON sys_download_tasks (owner_id);
CREATE INDEX IF NOT EXISTS idx_download_deleted_at    ON sys_download_tasks (deleted_at);
