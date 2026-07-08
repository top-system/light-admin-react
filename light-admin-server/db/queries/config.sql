-- config.sql
-- SQL for the system configuration domain (t_config). Dynamic filters use
-- sqlc.narg; ordering is fixed to id DESC (the previous GORM default).

-- name: GetConfig :one
SELECT * FROM t_config WHERE id = $1 AND is_deleted = 0;

-- name: GetConfigByKey :one
SELECT * FROM t_config WHERE config_key = $1 AND is_deleted = 0;

-- name: ListConfigs :many
SELECT * FROM t_config
WHERE is_deleted = 0
  AND (sqlc.narg('keywords')::text IS NULL OR (config_name LIKE sqlc.narg('keywords') OR config_key LIKE sqlc.narg('keywords')))
ORDER BY id DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountConfigs :one
SELECT COUNT(*) FROM t_config
WHERE is_deleted = 0
  AND (sqlc.narg('keywords')::text IS NULL OR (config_name LIKE sqlc.narg('keywords') OR config_key LIKE sqlc.narg('keywords')));

-- name: ListAllConfigs :many
SELECT * FROM t_config WHERE is_deleted = 0;

-- name: CountConfigsByKey :one
-- Counts non-deleted configs with the given key, optionally excluding one id.
SELECT COUNT(*) FROM t_config
WHERE config_key = $1 AND is_deleted = 0
  AND (sqlc.narg('exclude_id')::text IS NULL OR id <> sqlc.narg('exclude_id'));

-- name: CreateConfig :exec
INSERT INTO t_config (
    id, config_name, config_key, config_value, remark, create_by, update_by, is_deleted,
    create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateConfig :exec
UPDATE t_config SET
    config_name  = $2,
    config_key   = $3,
    config_value = $4,
    remark       = $5,
    update_by    = $6,
    update_time  = sqlc.arg('now')
WHERE id = $1;

-- name: SoftDeleteConfig :exec
UPDATE t_config SET is_deleted = 1, update_by = $2, update_time = sqlc.arg('now') WHERE id = $1;
