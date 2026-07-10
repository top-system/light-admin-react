-- dict.sql (SQLite dialect)
-- Derived from queries/postgres/dict.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header).

-- name: GetDict :one
SELECT * FROM t_dict WHERE id = sqlc.arg('id') AND is_deleted = 0;

-- name: GetDictByCode :one
SELECT * FROM t_dict
WHERE dict_code = sqlc.arg('dict_code') AND is_deleted = 0
  AND (sqlc.narg('exclude_id') IS NULL OR id <> sqlc.narg('exclude_id'))
LIMIT 1;

-- name: ListDicts :many
SELECT * FROM t_dict
WHERE is_deleted = 0
  AND (sqlc.narg('keywords') IS NULL OR (name LIKE sqlc.narg('keywords') OR dict_code LIKE sqlc.narg('keywords')))
ORDER BY create_time DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountDicts :one
SELECT COUNT(*) FROM t_dict
WHERE is_deleted = 0
  AND (sqlc.narg('keywords') IS NULL OR (name LIKE sqlc.narg('keywords') OR dict_code LIKE sqlc.narg('keywords')));

-- name: ListEnabledDicts :many
SELECT * FROM t_dict WHERE is_deleted = 0 AND status = 1;

-- name: ListDictsByIDs :many
SELECT * FROM t_dict WHERE is_deleted = 0 AND id IN (sqlc.slice('ids'));

-- name: CreateDict :exec
INSERT INTO t_dict (
    id, dict_code, name, status, remark, create_by, update_by, is_deleted,
    create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('dict_code'), sqlc.arg('name'), sqlc.arg('status'),
    sqlc.arg('remark'), sqlc.arg('create_by'), sqlc.arg('update_by'), sqlc.arg('is_deleted'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateDict :exec
UPDATE t_dict SET
    dict_code   = sqlc.arg('dict_code'),
    name        = sqlc.arg('name'),
    status      = sqlc.arg('status'),
    remark      = sqlc.arg('remark'),
    update_by   = sqlc.arg('update_by'),
    update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: SoftDeleteDict :exec
UPDATE t_dict SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');

-- name: SoftDeleteDictsByIDs :exec
UPDATE t_dict SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id IN (sqlc.slice('ids'));
