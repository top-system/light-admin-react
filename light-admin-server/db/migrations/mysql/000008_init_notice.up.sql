-- 000008_init_notice.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000008_init_notice.up.sql.
-- publish_time / revoke_time / read_time are nullable.

CREATE TABLE IF NOT EXISTS t_notice (
    id              CHAR(32)      NOT NULL,
    title           VARCHAR(50)   NOT NULL DEFAULT '',
    content         TEXT          NOT NULL DEFAULT ('') ,
    type            INT           NOT NULL DEFAULT 0,
    level           VARCHAR(5)    NOT NULL DEFAULT '',
    target_type     INT           NOT NULL DEFAULT 0,
    target_user_ids VARCHAR(255)  NOT NULL DEFAULT '',
    publisher_id    VARCHAR(64)   NOT NULL DEFAULT '',
    publish_status  INT           NOT NULL DEFAULT 0,
    publish_time    DATETIME(6)   NULL,
    revoke_time     DATETIME(6)   NULL,
    create_by       VARCHAR(64)   NOT NULL DEFAULT '',
    create_time     DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_by       VARCHAR(64)   NOT NULL DEFAULT '',
    update_time     DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    is_deleted      INT           NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    KEY idx_publish_status (publish_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS t_user_notice (
    id          CHAR(32)     NOT NULL,
    notice_id   CHAR(32)     NOT NULL DEFAULT '',
    user_id     CHAR(32)     NOT NULL DEFAULT '',
    is_read     INT          NOT NULL DEFAULT 0,
    read_time   DATETIME(6)  NULL,
    create_time DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_time DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    is_deleted  INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    KEY idx_user_notice_user_id   (user_id),
    KEY idx_user_notice_notice_id (notice_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
