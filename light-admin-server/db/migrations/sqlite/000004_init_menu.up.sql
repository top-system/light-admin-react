-- 000004_init_menu.up.sql (SQLite dialect)
-- Derived from migrations/postgres/000004_init_menu.up.sql.
-- t_menu has no soft-delete column: deletion is a hard DELETE.

CREATE TABLE IF NOT EXISTS t_menu (
    id          VARCHAR(32)   NOT NULL,
    parent_id   VARCHAR(32)   NOT NULL DEFAULT '',
    tree_path   VARCHAR(255)  NOT NULL DEFAULT '',
    name        VARCHAR(64)   NOT NULL DEFAULT '',
    type        INTEGER       NOT NULL DEFAULT 0,
    route_name  VARCHAR(255)  NOT NULL DEFAULT '',
    route_path  VARCHAR(128)  NOT NULL DEFAULT '',
    component   VARCHAR(128)  NOT NULL DEFAULT '',
    perm        VARCHAR(128)  NOT NULL DEFAULT '',
    always_show INTEGER       NOT NULL DEFAULT 0,
    keep_alive  INTEGER       NOT NULL DEFAULT 0,
    visible     INTEGER       NOT NULL DEFAULT 1,
    sort        INTEGER       NOT NULL DEFAULT 0,
    icon        VARCHAR(64)   NOT NULL DEFAULT '',
    redirect    VARCHAR(128)  NOT NULL DEFAULT '',
    params      VARCHAR(255)  NOT NULL DEFAULT '',
    create_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_t_menu PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_menu_parent_id ON t_menu (parent_id);
CREATE INDEX IF NOT EXISTS idx_tree_path      ON t_menu (tree_path);
CREATE INDEX IF NOT EXISTS idx_type           ON t_menu (type);
