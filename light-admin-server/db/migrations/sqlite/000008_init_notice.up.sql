-- 000008_init_notice.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000008_init_notice.up.sql.
-- publish_time / revoke_time / read_time are nullable.

CREATE TABLE IF NOT EXISTS t_notice (
    id              VARCHAR(32)   NOT NULL,
    title           VARCHAR(50)   NOT NULL DEFAULT '',
    content         TEXT          NOT NULL DEFAULT '',
    type            INTEGER       NOT NULL DEFAULT 0,
    level           VARCHAR(5)    NOT NULL DEFAULT '',
    target_type     INTEGER       NOT NULL DEFAULT 0,
    target_user_ids VARCHAR(255)  NOT NULL DEFAULT '',
    publisher_id    VARCHAR(64)   NOT NULL DEFAULT '',
    publish_status  INTEGER       NOT NULL DEFAULT 0,
    publish_time    DATETIME,
    revoke_time     DATETIME,
    create_by       VARCHAR(64)   NOT NULL DEFAULT '',
    create_time     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by       VARCHAR(64)   NOT NULL DEFAULT '',
    update_time     DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted      INTEGER       NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_notice PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_publish_status ON t_notice (publish_status);

CREATE TABLE IF NOT EXISTS t_user_notice (
    id          VARCHAR(32)  NOT NULL,
    notice_id   VARCHAR(32)  NOT NULL DEFAULT '',
    user_id     VARCHAR(32)  NOT NULL DEFAULT '',
    is_read     INTEGER      NOT NULL DEFAULT 0,
    read_time   DATETIME,
    create_time DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted  INTEGER      NOT NULL DEFAULT 0,
    CONSTRAINT pk_t_user_notice PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_user_notice_user_id   ON t_user_notice (user_id);
CREATE INDEX IF NOT EXISTS idx_user_notice_notice_id ON t_user_notice (notice_id);
