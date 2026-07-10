-- 000006_init_config.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000006_init_config.up.sql.

CREATE TABLE IF NOT EXISTS t_config (
    id           VARCHAR(32)   NOT NULL,
    config_name  VARCHAR(50)   NOT NULL DEFAULT '',
    config_key   VARCHAR(50)   NOT NULL DEFAULT '',
    config_value VARCHAR(100)  NOT NULL DEFAULT '',
    remark       VARCHAR(255)  NOT NULL DEFAULT '',
    create_time  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by    VARCHAR(64)   NOT NULL DEFAULT '',
    update_time  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by    VARCHAR(64)   NOT NULL DEFAULT '',
    is_deleted   INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_config PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_config_key ON t_config (config_key);
