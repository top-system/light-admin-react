-- 000007_init_log.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000007_init_log.up.sql. TEXT columns use
-- expression defaults (('')), which require MySQL 8.0.13+.
-- sys_log has no soft-delete column: deletion is a hard DELETE.

CREATE TABLE IF NOT EXISTS sys_log (
    id               CHAR(32)      NOT NULL,
    module           VARCHAR(50)   NOT NULL DEFAULT '',
    request_method   VARCHAR(64)   NOT NULL DEFAULT '',
    request_params   TEXT          NOT NULL DEFAULT ('') ,
    response_content TEXT          NOT NULL DEFAULT ('') ,
    content          VARCHAR(255)  NOT NULL DEFAULT '',
    request_uri      VARCHAR(255)  NOT NULL DEFAULT '',
    method           VARCHAR(255)  NOT NULL DEFAULT '',
    ip               VARCHAR(45)   NOT NULL DEFAULT '',
    province         VARCHAR(100)  NOT NULL DEFAULT '',
    city             VARCHAR(100)  NOT NULL DEFAULT '',
    execution_time   BIGINT        NOT NULL DEFAULT 0,
    browser          VARCHAR(100)  NOT NULL DEFAULT '',
    browser_version  VARCHAR(100)  NOT NULL DEFAULT '',
    os               VARCHAR(100)  NOT NULL DEFAULT '',
    create_by        VARCHAR(64)   NOT NULL DEFAULT '',
    create_time      DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_module          (module),
    KEY idx_log_create_by   (create_by),
    KEY idx_log_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
