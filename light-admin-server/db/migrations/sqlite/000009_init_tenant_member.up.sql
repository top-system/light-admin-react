-- 000009_init_tenant_member.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000009_init_tenant_member.up.sql.

CREATE TABLE IF NOT EXISTS t_tenant (
    id          VARCHAR(32)   NOT NULL,
    code        VARCHAR(64)   NOT NULL DEFAULT '',
    name        VARCHAR(128)  NOT NULL DEFAULT '',
    status      INTEGER       NOT NULL DEFAULT 1,
    create_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_tenant PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_code ON t_tenant (code);

CREATE TABLE IF NOT EXISTS t_member (
    id              VARCHAR(32)   NOT NULL,
    tenant_id       VARCHAR(32)   NOT NULL DEFAULT '',
    username        VARCHAR(64)   NOT NULL DEFAULT '',
    email           VARCHAR(128)  NOT NULL DEFAULT '',
    mobile          VARCHAR(20)   NOT NULL DEFAULT '',
    password        VARCHAR(100)  NOT NULL DEFAULT '',
    nickname        VARCHAR(64)   NOT NULL DEFAULT '',
    avatar          VARCHAR(255)  NOT NULL DEFAULT '',
    gender          INTEGER       NOT NULL DEFAULT 0,
    status          INTEGER       NOT NULL DEFAULT 1,
    last_login_time DATETIME,
    last_login_ip   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted      INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_member PRIMARY KEY (id)
);

CREATE INDEX        IF NOT EXISTS idx_member_tenant    ON t_member (tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_username ON t_member (tenant_id, username);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_email    ON t_member (tenant_id, email);
