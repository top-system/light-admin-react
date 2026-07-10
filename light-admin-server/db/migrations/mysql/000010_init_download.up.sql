-- 000010_init_download.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000010_init_download.up.sql. BIGSERIAL maps
-- to BIGINT AUTO_INCREMENT. deleted_at is a plain nullable column, so deletes
-- are hard DELETEs.

CREATE TABLE IF NOT EXISTS sys_download_tasks (
    id             BIGINT        NOT NULL AUTO_INCREMENT,
    queue_task_id  BIGINT        NOT NULL DEFAULT 0,
    task_id        VARCHAR(100)  NOT NULL DEFAULT '',
    hash           VARCHAR(100)  NOT NULL DEFAULT '',
    name           VARCHAR(500)  NOT NULL DEFAULT '',
    url            TEXT          NOT NULL DEFAULT ('') ,
    downloader     VARCHAR(50)   NOT NULL DEFAULT '',
    status         VARCHAR(50)   NOT NULL DEFAULT '',
    total          BIGINT        NOT NULL DEFAULT 0,
    downloaded     BIGINT        NOT NULL DEFAULT 0,
    download_speed BIGINT        NOT NULL DEFAULT 0,
    uploaded       BIGINT        NOT NULL DEFAULT 0,
    upload_speed   BIGINT        NOT NULL DEFAULT 0,
    save_path      VARCHAR(500)  NOT NULL DEFAULT '',
    error_message  TEXT          NOT NULL DEFAULT ('') ,
    owner_id       CHAR(32)      NOT NULL DEFAULT '',
    created_at     DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at     DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    deleted_at     DATETIME(6)   NULL,
    PRIMARY KEY (id),
    KEY idx_download_queue_task_id (queue_task_id),
    KEY idx_download_task_id       (task_id),
    KEY idx_download_hash          (hash),
    KEY idx_download_status        (status),
    KEY idx_download_owner_id      (owner_id),
    KEY idx_download_deleted_at    (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
