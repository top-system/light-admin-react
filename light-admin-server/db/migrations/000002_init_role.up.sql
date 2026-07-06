-- 000002_init_role.up.sql
-- Role (t_role) and the Role<->Menu association (t_role_menu). These tables move
-- from GORM AutoMigrate to golang-migrate ownership as the Role module migrates
-- to the sqlc data layer. `IF NOT EXISTS` keeps the migration idempotent on
-- databases where GORM AutoMigrate already created these tables, so the schema
-- source of truth transfers to migrations without disrupting existing data.

CREATE TABLE IF NOT EXISTS t_role (
    id          CHAR(32)      NOT NULL,
    name        VARCHAR(64)   NOT NULL DEFAULT '',
    code        VARCHAR(32)   NOT NULL DEFAULT '',
    sort        INTEGER       NOT NULL DEFAULT 0,
    status      INTEGER       NOT NULL DEFAULT 1,
    data_scope  INTEGER       NOT NULL DEFAULT 0,
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_role PRIMARY KEY (id)
);

-- Mirrors the GORM model's uniqueIndex tags (uk_role_name / uk_role_code).
CREATE UNIQUE INDEX IF NOT EXISTS uk_role_name      ON t_role (name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_role_code      ON t_role (code);
CREATE INDEX        IF NOT EXISTS idx_role_status     ON t_role (status);
CREATE INDEX        IF NOT EXISTS idx_role_is_deleted ON t_role (is_deleted);

-- Role <-> Menu association (t_role_menu). No surrogate key; the (role_id, menu_id)
-- pair is unique, matching the GORM model's uniqueIndex uk_roleid_menuid and
-- backing the ON CONFLICT clause used by the batch insert.
CREATE TABLE IF NOT EXISTS t_role_menu (
    role_id CHAR(32) NOT NULL,
    menu_id CHAR(32) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_roleid_menuid      ON t_role_menu (role_id, menu_id);
CREATE INDEX        IF NOT EXISTS idx_role_menu_menu_id ON t_role_menu (menu_id);
