-- 000001_init_user.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000001_init_user.up.sql per the type map in
-- docs/multi-database-plan.md section 4: TIMESTAMPTZ -> DATETIME(6) (session
-- time zone pinned via the DSN loc parameter), NOW() -> CURRENT_TIMESTAMP(6).
-- Indexes are declared inside CREATE TABLE because MySQL's CREATE INDEX has no
-- IF NOT EXISTS. Requires MySQL 8.0.13+ (expression defaults on TEXT columns).

CREATE TABLE IF NOT EXISTS t_user (
    id          CHAR(32)      NOT NULL,
    username    VARCHAR(64)   NOT NULL DEFAULT '',
    nickname    VARCHAR(64)   NOT NULL DEFAULT '',
    gender      INT           NOT NULL DEFAULT 1,
    password    VARCHAR(100)  NOT NULL DEFAULT '',
    dept_id     VARCHAR(32)   NOT NULL DEFAULT '',
    avatar      VARCHAR(255)  NOT NULL DEFAULT '',
    mobile      VARCHAR(20)   NOT NULL DEFAULT '',
    status      INT           NOT NULL DEFAULT 1,
    email       VARCHAR(128)  NOT NULL DEFAULT '',
    create_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    is_deleted  INT           NOT NULL DEFAULT 0,
    openid      VARCHAR(28)   NOT NULL DEFAULT '',
    PRIMARY KEY (id),
    KEY idx_username         (username),
    KEY idx_dept_id          (dept_id),
    KEY idx_user_status      (status),
    KEY idx_user_create_time (create_time),
    KEY idx_user_is_deleted  (is_deleted)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- User <-> Role association (t_user_role). Composite primary key.
CREATE TABLE IF NOT EXISTS t_user_role (
    user_id CHAR(32) NOT NULL,
    role_id CHAR(32) NOT NULL,
    PRIMARY KEY (user_id, role_id),
    KEY idx_user_role_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
