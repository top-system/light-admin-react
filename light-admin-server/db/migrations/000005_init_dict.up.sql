-- 000005_init_dict.up.sql
-- Dictionary tables: t_dict (dictionaries) and t_dict_item (dictionary items).
-- Moves from GORM AutoMigrate to golang-migrate ownership as the Dict module
-- migrates to the sqlc data layer. `IF NOT EXISTS` keeps the migration idempotent
-- on databases where GORM AutoMigrate already created the tables.

CREATE TABLE IF NOT EXISTS t_dict (
    id          CHAR(32)      NOT NULL,
    dict_code   VARCHAR(100)  NOT NULL DEFAULT '',
    name        VARCHAR(100)  NOT NULL DEFAULT '',
    status      INTEGER       NOT NULL DEFAULT 1,
    remark      VARCHAR(255)  NOT NULL DEFAULT '',
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_dict PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_code        ON t_dict (dict_code);
CREATE INDEX        IF NOT EXISTS idx_dict_status     ON t_dict (status);
CREATE INDEX        IF NOT EXISTS idx_dict_is_deleted ON t_dict (is_deleted);

CREATE TABLE IF NOT EXISTS t_dict_item (
    id          CHAR(32)      NOT NULL,
    dict_code   VARCHAR(100)  NOT NULL DEFAULT '',
    label       VARCHAR(100)  NOT NULL DEFAULT '',
    value       VARCHAR(100)  NOT NULL DEFAULT '',
    tag_type    VARCHAR(50)   NOT NULL DEFAULT '',
    sort        INTEGER       NOT NULL DEFAULT 0,
    status      INTEGER       NOT NULL DEFAULT 1,
    remark      VARCHAR(255)  NOT NULL DEFAULT '',
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_dict_item PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_dict_code ON t_dict_item (dict_code);
