-- 000001_init_user.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000001_init_user.up.sql per the type map in
-- docs/multi-database-plan.md §4: TIMESTAMPTZ -> DATETIME (UTC text storage),
-- CHAR -> VARCHAR (TEXT affinity). Timestamps are written by the application
-- (Go time.Time); the CURRENT_TIMESTAMP default is a parity fallback only.

CREATE TABLE IF NOT EXISTS t_user (
    id          VARCHAR(32)   NOT NULL,
    username    VARCHAR(64)   NOT NULL DEFAULT '',
    nickname    VARCHAR(64)   NOT NULL DEFAULT '',
    gender      INTEGER       NOT NULL DEFAULT 1,
    password    VARCHAR(100)  NOT NULL DEFAULT '',
    dept_id     VARCHAR(32)   NOT NULL DEFAULT '',
    avatar      VARCHAR(255)  NOT NULL DEFAULT '',
    mobile      VARCHAR(20)   NOT NULL DEFAULT '',
    status      INTEGER       NOT NULL DEFAULT 1,
    email       VARCHAR(128)  NOT NULL DEFAULT '',
    create_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    openid      VARCHAR(28)   NOT NULL DEFAULT '',
    CONSTRAINT pk_t_user PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_username         ON t_user (username);
CREATE INDEX IF NOT EXISTS idx_dept_id          ON t_user (dept_id);
CREATE INDEX IF NOT EXISTS idx_user_status      ON t_user (status);
CREATE INDEX IF NOT EXISTS idx_user_create_time ON t_user (create_time);
CREATE INDEX IF NOT EXISTS idx_user_is_deleted  ON t_user (is_deleted);

-- User <-> Role association (t_user_role). Composite primary key.
CREATE TABLE IF NOT EXISTS t_user_role (
    user_id VARCHAR(32) NOT NULL,
    role_id VARCHAR(32) NOT NULL,
    CONSTRAINT pk_t_user_role PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_role_role_id ON t_user_role (role_id);
