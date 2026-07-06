-- 000008_init_notice.up.sql
-- Notice (t_notice) and per-user notice state (t_user_notice). Moves from GORM
-- AutoMigrate to golang-migrate ownership as the Notice module migrates to sqlc.
-- `IF NOT EXISTS` keeps the migration idempotent on databases where GORM already
-- created the tables. publish_time / revoke_time / read_time are nullable.

CREATE TABLE IF NOT EXISTS t_notice (
    id              CHAR(32)      NOT NULL,
    title           VARCHAR(50)   NOT NULL DEFAULT '',
    content         TEXT          NOT NULL DEFAULT '',
    type            INTEGER       NOT NULL DEFAULT 0,
    level           VARCHAR(5)    NOT NULL DEFAULT '',
    target_type     INTEGER       NOT NULL DEFAULT 0,
    target_user_ids VARCHAR(255)  NOT NULL DEFAULT '',
    publisher_id    VARCHAR(64)   NOT NULL DEFAULT '',
    publish_status  INTEGER       NOT NULL DEFAULT 0,
    publish_time    TIMESTAMPTZ,
    revoke_time     TIMESTAMPTZ,
    create_by       VARCHAR(64)   NOT NULL DEFAULT '',
    create_time     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_by       VARCHAR(64)   NOT NULL DEFAULT '',
    update_time     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    is_deleted      INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_notice PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_publish_status ON t_notice (publish_status);

CREATE TABLE IF NOT EXISTS t_user_notice (
    id          CHAR(32)     NOT NULL,
    notice_id   CHAR(32)     NOT NULL DEFAULT '',
    user_id     CHAR(32)     NOT NULL DEFAULT '',
    is_read     INTEGER      NOT NULL DEFAULT 0,
    read_time   TIMESTAMPTZ,
    create_time TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted  INTEGER      NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_user_notice PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_user_notice_user_id   ON t_user_notice (user_id);
CREATE INDEX IF NOT EXISTS idx_user_notice_notice_id ON t_user_notice (notice_id);
