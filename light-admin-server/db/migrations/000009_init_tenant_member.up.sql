-- 000009_init_tenant_member.up.sql
-- Multi-tenant tables: t_tenant (tenants) and t_member (tenant-scoped C-side
-- users). Moves from GORM AutoMigrate to golang-migrate ownership as the
-- Tenant/Member module migrates to sqlc. `IF NOT EXISTS` keeps the migration
-- idempotent on databases where GORM already created the tables.

CREATE TABLE IF NOT EXISTS t_tenant (
    id          CHAR(32)      NOT NULL,
    code        VARCHAR(64)   NOT NULL DEFAULT '',
    name        VARCHAR(128)  NOT NULL DEFAULT '',
    status      INTEGER       NOT NULL DEFAULT 1,
    create_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    is_deleted  INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_tenant PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_code ON t_tenant (code);

CREATE TABLE IF NOT EXISTS t_member (
    id              CHAR(32)      NOT NULL,
    tenant_id       CHAR(32)      NOT NULL DEFAULT '',
    username        VARCHAR(64)   NOT NULL DEFAULT '',
    email           VARCHAR(128)  NOT NULL DEFAULT '',
    mobile          VARCHAR(20)   NOT NULL DEFAULT '',
    password        VARCHAR(100)  NOT NULL DEFAULT '',
    nickname        VARCHAR(64)   NOT NULL DEFAULT '',
    avatar          VARCHAR(255)  NOT NULL DEFAULT '',
    gender          INTEGER       NOT NULL DEFAULT 0,
    status          INTEGER       NOT NULL DEFAULT 1,
    last_login_time TIMESTAMPTZ,
    last_login_ip   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_time     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    is_deleted      INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_member PRIMARY KEY (id)
);

CREATE INDEX        IF NOT EXISTS idx_member_tenant     ON t_member (tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_username  ON t_member (tenant_id, username);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_email     ON t_member (tenant_id, email);
