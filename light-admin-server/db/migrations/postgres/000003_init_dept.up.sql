-- 000003_init_dept.up.sql
-- Department table (t_dept). Moves from GORM AutoMigrate to golang-migrate
-- ownership as the Dept module migrates to the sqlc data layer. `IF NOT EXISTS`
-- keeps the migration idempotent on databases where GORM AutoMigrate already
-- created the table.

CREATE TABLE IF NOT EXISTS t_dept (
    id          CHAR(32)      NOT NULL,
    name        VARCHAR(100)  NOT NULL DEFAULT '',
    code        VARCHAR(100)  NOT NULL DEFAULT '',
    parent_id   CHAR(32)      NOT NULL DEFAULT '',
    tree_path   VARCHAR(255)  NOT NULL DEFAULT '',
    sort        INTEGER       NOT NULL DEFAULT 0,
    status      INTEGER       NOT NULL DEFAULT 1,
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_dept PRIMARY KEY (id)
);

-- Mirrors the GORM model's uniqueIndex tag (uk_dept_code).
CREATE UNIQUE INDEX IF NOT EXISTS uk_dept_code       ON t_dept (code);
CREATE INDEX        IF NOT EXISTS idx_dept_parent_id ON t_dept (parent_id);
CREATE INDEX        IF NOT EXISTS idx_dept_status    ON t_dept (status);
CREATE INDEX        IF NOT EXISTS idx_dept_is_deleted ON t_dept (is_deleted);
