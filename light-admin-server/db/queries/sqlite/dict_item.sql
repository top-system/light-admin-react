-- dict_item.sql (SQLite dialect)
-- Derived from queries/postgres/dict_item.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header).

-- name: GetDictItem :one
SELECT * FROM t_dict_item WHERE id = sqlc.arg('id') AND is_deleted = 0;

-- name: ListDictItems :many
SELECT * FROM t_dict_item
WHERE is_deleted = 0
  AND (sqlc.narg('dict_code') IS NULL OR dict_code = sqlc.narg('dict_code'))
  AND (sqlc.narg('keywords')  IS NULL OR (label LIKE sqlc.narg('keywords') OR value LIKE sqlc.narg('keywords')))
ORDER BY sort ASC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountDictItems :one
SELECT COUNT(*) FROM t_dict_item
WHERE is_deleted = 0
  AND (sqlc.narg('dict_code') IS NULL OR dict_code = sqlc.narg('dict_code'))
  AND (sqlc.narg('keywords')  IS NULL OR (label LIKE sqlc.narg('keywords') OR value LIKE sqlc.narg('keywords')));

-- name: ListDictItemsByDictCode :many
SELECT * FROM t_dict_item
WHERE dict_code = sqlc.arg('dict_code') AND is_deleted = 0 AND status = 1
ORDER BY sort ASC;

-- name: CreateDictItem :exec
INSERT INTO t_dict_item (
    id, dict_code, label, value, tag_type, sort, status, remark,
    create_by, update_by, is_deleted, create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('dict_code'), sqlc.arg('label'), sqlc.arg('value'),
    sqlc.arg('tag_type'), sqlc.arg('sort'), sqlc.arg('status'), sqlc.arg('remark'),
    sqlc.arg('create_by'), sqlc.arg('update_by'), sqlc.arg('is_deleted'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateDictItem :exec
UPDATE t_dict_item SET
    dict_code   = sqlc.arg('dict_code'),
    label       = sqlc.arg('label'),
    value       = sqlc.arg('value'),
    tag_type    = sqlc.arg('tag_type'),
    sort        = sqlc.arg('sort'),
    status      = sqlc.arg('status'),
    remark      = sqlc.arg('remark'),
    update_by   = sqlc.arg('update_by'),
    update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: SoftDeleteDictItem :exec
UPDATE t_dict_item SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now') WHERE id = sqlc.arg('id');

-- name: SoftDeleteDictItemsByIDs :exec
UPDATE t_dict_item SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id IN (sqlc.slice('ids'));

-- name: SoftDeleteDictItemsByDictCodes :exec
UPDATE t_dict_item SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE dict_code IN (sqlc.slice('dict_codes'));

-- name: UpdateDictItemsDictCode :exec
-- Cascade a dictionary code rename onto its items. Called by DictRepository when a
-- dictionary's code changes.
UPDATE t_dict_item SET dict_code = sqlc.arg('new_code'), update_time = sqlc.arg('now')
WHERE dict_code = sqlc.arg('old_code');
