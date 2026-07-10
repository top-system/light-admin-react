-- 000005_init_dict.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000005_init_dict.up.sql.

CREATE TABLE IF NOT EXISTS t_dict (
    id          CHAR(32)      NOT NULL,
    dict_code   VARCHAR(100)  NOT NULL DEFAULT '',
    name        VARCHAR(100)  NOT NULL DEFAULT '',
    status      INT           NOT NULL DEFAULT 1,
    remark      VARCHAR(255)  NOT NULL DEFAULT '',
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    is_deleted  INT           NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_dict_code (dict_code),
    KEY idx_dict_status     (status),
    KEY idx_dict_is_deleted (is_deleted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS t_dict_item (
    id          CHAR(32)      NOT NULL,
    dict_code   VARCHAR(100)  NOT NULL DEFAULT '',
    label       VARCHAR(100)  NOT NULL DEFAULT '',
    value       VARCHAR(100)  NOT NULL DEFAULT '',
    tag_type    VARCHAR(50)   NOT NULL DEFAULT '',
    sort        INT           NOT NULL DEFAULT 0,
    status      INT           NOT NULL DEFAULT 1,
    remark      VARCHAR(255)  NOT NULL DEFAULT '',
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    is_deleted  INT           NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    KEY idx_dict_code (dict_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
