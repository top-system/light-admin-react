-- 000004_init_menu.up.sql
-- Menu table (t_menu). Moves from GORM AutoMigrate to golang-migrate ownership as
-- the Menu module migrates to the sqlc data layer. `IF NOT EXISTS` keeps the
-- migration idempotent on databases where GORM AutoMigrate already created it.
-- t_menu has no soft-delete column: deletion is a hard DELETE.

CREATE TABLE IF NOT EXISTS t_menu (
    id          CHAR(32)      NOT NULL,
    parent_id   CHAR(32)      NOT NULL DEFAULT '',
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
    create_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_t_menu PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_menu_parent_id ON t_menu (parent_id);
CREATE INDEX IF NOT EXISTS idx_tree_path      ON t_menu (tree_path);
CREATE INDEX IF NOT EXISTS idx_type           ON t_menu (type);
