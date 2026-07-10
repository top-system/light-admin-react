-- 000002_init_role.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000002_init_role.up.sql (see 000001 header
-- for the dialect rules).

CREATE TABLE IF NOT EXISTS t_role (
    id          CHAR(32)      NOT NULL,
    name        VARCHAR(64)   NOT NULL DEFAULT '',
    code        VARCHAR(32)   NOT NULL DEFAULT '',
    sort        INT           NOT NULL DEFAULT 0,
    status      INT           NOT NULL DEFAULT 1,
    data_scope  INT           NOT NULL DEFAULT 0,
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    is_deleted  INT           NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uk_role_name      (name),
    UNIQUE KEY uk_role_code      (code),
    KEY idx_role_status     (status),
    KEY idx_role_is_deleted (is_deleted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Role <-> Menu association (t_role_menu). The (role_id, menu_id) pair is
-- unique, backing the INSERT IGNORE used by the batch insert.
CREATE TABLE IF NOT EXISTS t_role_menu (
    role_id CHAR(32) NOT NULL,
    menu_id CHAR(32) NOT NULL,
    UNIQUE KEY uk_roleid_menuid      (role_id, menu_id),
    KEY idx_role_menu_menu_id (menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
