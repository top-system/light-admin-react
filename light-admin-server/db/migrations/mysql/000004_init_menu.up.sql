-- 000004_init_menu.up.sql (MySQL dialect)
-- Derived from migrations/postgres/000004_init_menu.up.sql.
-- t_menu has no soft-delete column: deletion is a hard DELETE.

CREATE TABLE IF NOT EXISTS t_menu (
    id          CHAR(32)      NOT NULL,
    parent_id   CHAR(32)      NOT NULL DEFAULT '',
    tree_path   VARCHAR(255)  NOT NULL DEFAULT '',
    name        VARCHAR(64)   NOT NULL DEFAULT '',
    type        INT           NOT NULL DEFAULT 0,
    route_name  VARCHAR(255)  NOT NULL DEFAULT '',
    route_path  VARCHAR(128)  NOT NULL DEFAULT '',
    component   VARCHAR(128)  NOT NULL DEFAULT '',
    perm        VARCHAR(128)  NOT NULL DEFAULT '',
    always_show INT           NOT NULL DEFAULT 0,
    keep_alive  INT           NOT NULL DEFAULT 0,
    visible     INT           NOT NULL DEFAULT 1,
    sort        INT           NOT NULL DEFAULT 0,
    icon        VARCHAR(64)   NOT NULL DEFAULT '',
    redirect    VARCHAR(128)  NOT NULL DEFAULT '',
    params      VARCHAR(255)  NOT NULL DEFAULT '',
    create_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    update_time DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_menu_parent_id (parent_id),
    KEY idx_tree_path      (tree_path),
    KEY idx_type           (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
