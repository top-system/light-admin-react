-- 000002_init_role.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000002_init_role.up.sql (see 000001 header
-- for the dialect rules).

CREATE TABLE IF NOT EXISTS t_role (
    id          VARCHAR(32)   NOT NULL,
    name        VARCHAR(64)   NOT NULL DEFAULT '',
    code        VARCHAR(32)   NOT NULL DEFAULT '',
    sort        INTEGER       NOT NULL DEFAULT 0,
    status      INTEGER       NOT NULL DEFAULT 1,
    data_scope  INTEGER       NOT NULL DEFAULT 0,
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_role PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_role_name        ON t_role (name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_role_code        ON t_role (code);
CREATE INDEX        IF NOT EXISTS idx_role_status     ON t_role (status);
CREATE INDEX        IF NOT EXISTS idx_role_is_deleted ON t_role (is_deleted);

-- Role <-> Menu association (t_role_menu). The (role_id, menu_id) pair is
-- unique, backing the insert-or-ignore used by the batch insert.
CREATE TABLE IF NOT EXISTS t_role_menu (
    role_id VARCHAR(32) NOT NULL,
    menu_id VARCHAR(32) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_roleid_menuid      ON t_role_menu (role_id, menu_id);
CREATE INDEX        IF NOT EXISTS idx_role_menu_menu_id ON t_role_menu (menu_id);
