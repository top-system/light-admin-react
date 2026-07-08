-- dict.sql
-- SQL for the dictionary domain (t_dict). Dynamic filters use sqlc.narg so one
-- prepared statement serves every filter combination. Ordering is fixed
-- (create_time DESC) — the previous GORM default.

-- name: GetDict :one
SELECT * FROM t_dict WHERE id = $1 AND is_deleted = 0;

-- name: GetDictByCode :one
SELECT * FROM t_dict
WHERE dict_code = $1 AND is_deleted = 0
  AND (sqlc.narg('exclude_id')::text IS NULL OR id <> sqlc.narg('exclude_id'))
LIMIT 1;

-- name: ListDicts :many
SELECT * FROM t_dict
WHERE is_deleted = 0
  AND (sqlc.narg('keywords')::text IS NULL OR (name LIKE sqlc.narg('keywords') OR dict_code LIKE sqlc.narg('keywords')))
ORDER BY create_time DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountDicts :one
SELECT COUNT(*) FROM t_dict
WHERE is_deleted = 0
  AND (sqlc.narg('keywords')::text IS NULL OR (name LIKE sqlc.narg('keywords') OR dict_code LIKE sqlc.narg('keywords')));

-- name: ListEnabledDicts :many
SELECT * FROM t_dict WHERE is_deleted = 0 AND status = 1;

-- name: ListDictsByIDs :many
SELECT * FROM t_dict WHERE is_deleted = 0 AND id = ANY(@ids::text[]);

-- name: CreateDict :exec
INSERT INTO t_dict (
    id, dict_code, name, status, remark, create_by, update_by, is_deleted,
    create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateDict :exec
UPDATE t_dict SET
    dict_code   = $2,
    name        = $3,
    status      = $4,
    remark      = $5,
    update_by   = $6,
    update_time = sqlc.arg('now')
WHERE id = $1;

-- name: SoftDeleteDict :exec
UPDATE t_dict SET is_deleted = 1, update_by = $2, update_time = sqlc.arg('now') WHERE id = $1;

-- name: SoftDeleteDictsByIDs :exec
UPDATE t_dict SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id = ANY(sqlc.arg('ids')::text[]);
