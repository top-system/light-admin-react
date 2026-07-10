-- config.sql (SQLite dialect)
-- Derived from queries/postgres/config.sql per docs/multi-database-plan.md section 5:
-- every parameter is a named sqlc.arg/narg (mixing anonymous ? with named args
-- mis-numbers the placeholders under sqlc's sqlite engine). LIMIT/OFFSET are
-- non-null args; the adapter maps a nil store limit/offset to -1/0 (SQLite:
-- LIMIT -1 = no limit).

-- name: GetConfig :one
SELECT * FROM t_config WHERE id = sqlc.arg('id') AND is_deleted = 0;

-- name: GetConfigByKey :one
SELECT * FROM t_config WHERE config_key = sqlc.arg('config_key') AND is_deleted = 0;

-- name: ListConfigs :many
SELECT * FROM t_config
WHERE is_deleted = 0
  AND (sqlc.narg('keywords') IS NULL OR (config_name LIKE sqlc.narg('keywords') OR config_key LIKE sqlc.narg('keywords')))
ORDER BY id DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountConfigs :one
SELECT COUNT(*) FROM t_config
WHERE is_deleted = 0
  AND (sqlc.narg('keywords') IS NULL OR (config_name LIKE sqlc.narg('keywords') OR config_key LIKE sqlc.narg('keywords')));

-- name: ListAllConfigs :many
SELECT * FROM t_config WHERE is_deleted = 0;

-- name: CountConfigsByKey :one
-- Counts non-deleted configs with the given key, optionally excluding one id.
SELECT COUNT(*) FROM t_config
WHERE config_key = sqlc.arg('config_key') AND is_deleted = 0
  AND (sqlc.narg('exclude_id') IS NULL OR id <> sqlc.narg('exclude_id'));

-- name: CreateConfig :exec
INSERT INTO t_config (
    id, config_name, config_key, config_value, remark, create_by, update_by, is_deleted,
    create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('config_name'), sqlc.arg('config_key'), sqlc.arg('config_value'),
    sqlc.arg('remark'), sqlc.arg('create_by'), sqlc.arg('update_by'), sqlc.arg('is_deleted'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateConfig :exec
UPDATE t_config SET
    config_name  = sqlc.arg('config_name'),
    config_key   = sqlc.arg('config_key'),
    config_value = sqlc.arg('config_value'),
    remark       = sqlc.arg('remark'),
    update_by    = sqlc.arg('update_by'),
    update_time  = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: SoftDeleteConfig :exec
UPDATE t_config SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');
