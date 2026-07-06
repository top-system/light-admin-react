-- 000006_init_config.up.sql
-- System configuration table (t_config). Moves from GORM AutoMigrate to
-- golang-migrate ownership as the Config module migrates to sqlc. `IF NOT EXISTS`
-- keeps the migration idempotent on databases where GORM already created it.

CREATE TABLE IF NOT EXISTS t_config (
    id           CHAR(32)      NOT NULL,
    config_name  VARCHAR(50)   NOT NULL DEFAULT '',
    config_key   VARCHAR(50)   NOT NULL DEFAULT '',
    config_value VARCHAR(100)  NOT NULL DEFAULT '',
    remark       VARCHAR(255)  NOT NULL DEFAULT '',
    create_time  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    create_by    VARCHAR(64)   NOT NULL DEFAULT '',
    update_time  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_by    VARCHAR(64)   NOT NULL DEFAULT '',
    is_deleted   INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_config PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_config_key ON t_config (config_key);
