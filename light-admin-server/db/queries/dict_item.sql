-- dict_item.sql
-- SQL for dictionary items (t_dict_item). Dynamic filters use sqlc.narg so one
-- prepared statement serves every filter combination. Ordering is fixed
-- (sort ASC) — the previous GORM default.

-- name: GetDictItem :one
SELECT * FROM t_dict_item WHERE id = $1 AND is_deleted = 0;

-- name: ListDictItems :many
SELECT * FROM t_dict_item
WHERE is_deleted = 0
  AND (sqlc.narg('dict_code')::text IS NULL OR dict_code = sqlc.narg('dict_code'))
  AND (sqlc.narg('keywords')::text  IS NULL OR (label LIKE sqlc.narg('keywords') OR value LIKE sqlc.narg('keywords')))
ORDER BY sort ASC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountDictItems :one
SELECT COUNT(*) FROM t_dict_item
WHERE is_deleted = 0
  AND (sqlc.narg('dict_code')::text IS NULL OR dict_code = sqlc.narg('dict_code'))
  AND (sqlc.narg('keywords')::text  IS NULL OR (label LIKE sqlc.narg('keywords') OR value LIKE sqlc.narg('keywords')));

-- name: ListDictItemsByDictCode :many
SELECT * FROM t_dict_item
WHERE dict_code = $1 AND is_deleted = 0 AND status = 1
ORDER BY sort ASC;

-- name: CreateDictItem :exec
INSERT INTO t_dict_item (
    id, dict_code, label, value, tag_type, sort, status, remark,
    create_by, update_by, is_deleted, create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateDictItem :exec
UPDATE t_dict_item SET
    dict_code   = $2,
    label       = $3,
    value       = $4,
    tag_type    = $5,
    sort        = $6,
    status      = $7,
    remark      = $8,
    update_by   = $9,
    update_time = sqlc.arg('now')
WHERE id = $1;

-- name: SoftDeleteDictItem :exec
UPDATE t_dict_item SET is_deleted = 1, update_by = $2, update_time = sqlc.arg('now') WHERE id = $1;

-- name: SoftDeleteDictItemsByIDs :exec
UPDATE t_dict_item SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id = ANY(sqlc.arg('ids')::text[]);

-- name: SoftDeleteDictItemsByDictCodes :exec
UPDATE t_dict_item SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE dict_code = ANY(sqlc.arg('dict_codes')::text[]);

-- name: UpdateDictItemsDictCode :exec
-- Cascade a dictionary code rename onto its items. Called by DictRepository when a
-- dictionary's code changes.
UPDATE t_dict_item SET dict_code = sqlc.arg('new_code'), update_time = sqlc.arg('now')
WHERE dict_code = sqlc.arg('old_code');
