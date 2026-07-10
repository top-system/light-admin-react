-- 000005_init_dict.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000005_init_dict.up.sql.

CREATE TABLE IF NOT EXISTS t_dict (
    id          VARCHAR(32)   NOT NULL,
    dict_code   VARCHAR(100)  NOT NULL DEFAULT '',
    name        VARCHAR(100)  NOT NULL DEFAULT '',
    status      INTEGER       NOT NULL DEFAULT 1,
    remark      VARCHAR(255)  NOT NULL DEFAULT '',
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_dict PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_dict_code        ON t_dict (dict_code);
CREATE INDEX        IF NOT EXISTS idx_dict_status     ON t_dict (status);
CREATE INDEX        IF NOT EXISTS idx_dict_is_deleted ON t_dict (is_deleted);

CREATE TABLE IF NOT EXISTS t_dict_item (
    id          VARCHAR(32)   NOT NULL,
    dict_code   VARCHAR(100)  NOT NULL DEFAULT '',
    label       VARCHAR(100)  NOT NULL DEFAULT '',
    value       VARCHAR(100)  NOT NULL DEFAULT '',
    tag_type    VARCHAR(50)   NOT NULL DEFAULT '',
    sort        INTEGER       NOT NULL DEFAULT 0,
    status      INTEGER       NOT NULL DEFAULT 1,
    remark      VARCHAR(255)  NOT NULL DEFAULT '',
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_dict_item PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_dict_code ON t_dict_item (dict_code);
