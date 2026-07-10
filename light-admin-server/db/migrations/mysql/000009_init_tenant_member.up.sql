-- 000009_init_tenant_member.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000009_init_tenant_member.up.sql.

CREATE TABLE IF NOT EXISTS t_tenant (
    id          CHAR(32)      NOT NULL,
    code        VARCHAR(64)   NOT NULL DEFAULT '',
    name        VARCHAR(128)  NOT NULL DEFAULT '',
    status      INT           NOT NULL DEFAULT 1,
    create_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    create_by   VARCHAR(64)   NOT NULL DEFAULT '',
    update_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_by   VARCHAR(64)   NOT NULL DEFAULT '',
    is_deleted  INT           NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    UNIQUE KEY uniq_tenant_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS t_member (
    id              CHAR(32)      NOT NULL,
    tenant_id       CHAR(32)      NOT NULL DEFAULT '',
    username        VARCHAR(64)   NOT NULL DEFAULT '',
    email           VARCHAR(128)  NOT NULL DEFAULT '',
    mobile          VARCHAR(20)   NOT NULL DEFAULT '',
    password        VARCHAR(100)  NOT NULL DEFAULT '',
    nickname        VARCHAR(64)   NOT NULL DEFAULT '',
    avatar          VARCHAR(255)  NOT NULL DEFAULT '',
    gender          INT           NOT NULL DEFAULT 0,
    status          INT           NOT NULL DEFAULT 1,
    last_login_time DATETIME(6)   NULL,
    last_login_ip   VARCHAR(64)   NOT NULL DEFAULT '',
    create_time     DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_time     DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    is_deleted      INT           NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    KEY idx_member_tenant (tenant_id),
    UNIQUE KEY uniq_tenant_username (tenant_id, username),
    UNIQUE KEY uniq_tenant_email    (tenant_id, email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
