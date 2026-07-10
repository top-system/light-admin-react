-- 000003_init_dept.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000003_init_dept.up.sql.

CREATE TABLE IF NOT EXISTS t_dept (
    id          VARCHAR(32)   NOT NULL,
    name        VARCHAR(100)  NOT NULL DEFAULT '',
    code        VARCHAR(100)  NOT NULL DEFAULT '',
    parent_id   VARCHAR(32)   NOT NULL DEFAULT '',
    tree_path   VARCHAR(255)  NOT NULL DEFAULT '',
    sort        INTEGER       NOT NULL DEFAULT 0,
    status      INTEGER       NOT NULL DEFAULT 1,
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_dept PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_dept_code        ON t_dept (code);
CREATE INDEX        IF NOT EXISTS idx_dept_parent_id  ON t_dept (parent_id);
CREATE INDEX        IF NOT EXISTS idx_dept_status     ON t_dept (status);
CREATE INDEX        IF NOT EXISTS idx_dept_is_deleted ON t_dept (is_deleted);
