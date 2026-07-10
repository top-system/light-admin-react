-- 000003_init_dept.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000003_init_dept.up.sql.

CREATE TABLE IF NOT EXISTS t_dept (
    id          CHAR(32)      NOT NULL,
    name        VARCHAR(100)  NOT NULL DEFAULT '',
    code        VARCHAR(100)  NOT NULL DEFAULT '',
    parent_id   CHAR(32)      NOT NULL DEFAULT '',
    tree_path   VARCHAR(255)  NOT NULL DEFAULT '',
    sort        INT           NOT NULL DEFAULT 0,
    status      INT           NOT NULL DEFAULT 1,
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    is_deleted  INT           NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_dept_code (code),
    KEY idx_dept_parent_id  (parent_id),
    KEY idx_dept_status     (status),
    KEY idx_dept_is_deleted (is_deleted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
