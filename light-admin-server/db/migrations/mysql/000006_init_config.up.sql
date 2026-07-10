-- 000006_init_config.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000006_init_config.up.sql.

CREATE TABLE IF NOT EXISTS t_config (
    id           CHAR(32)      NOT NULL,
    config_name  VARCHAR(50)   NOT NULL DEFAULT '',
    config_key   VARCHAR(50)   NOT NULL DEFAULT '',
    config_value VARCHAR(100)  NOT NULL DEFAULT '',
    remark       VARCHAR(255)  NOT NULL DEFAULT '',
    create_time  DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    create_by    VARCHAR(64)   NOT NULL DEFAULT '',
    update_time  DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_by    VARCHAR(64)   NOT NULL DEFAULT '',
    is_deleted   INT           NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
