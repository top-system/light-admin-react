-- 000007_init_log.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000007_init_log.up.sql.
-- sys_log has no soft-delete column: deletion is a hard DELETE.

CREATE TABLE IF NOT EXISTS sys_log (
    id               VARCHAR(32)   NOT NULL,
    module           VARCHAR(50)   NOT NULL DEFAULT '',
    request_method   VARCHAR(64)   NOT NULL DEFAULT '',
    request_params   TEXT          NOT NULL DEFAULT '',
    response_content TEXT          NOT NULL DEFAULT '',
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
    create_time      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_sys_log PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_module          ON sys_log (module);
CREATE INDEX IF NOT EXISTS idx_log_create_by   ON sys_log (create_by);
CREATE INDEX IF NOT EXISTS idx_log_create_time ON sys_log (create_time);
