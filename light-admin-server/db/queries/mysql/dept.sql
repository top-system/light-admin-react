-- dept.sql (MySQL dialect)
-- Derived from queries/postgres/dept.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header). Non-null id lists use sqlc.slice;
-- string concatenation uses CONCAT().

-- name: GetDept :one
SELECT * FROM t_dept
WHERE id = sqlc.arg('id') AND is_deleted = 0;

-- name: GetDeptByCode :one
-- exclude_id lets callers ignore a specific row (used when validating a code on
-- update). A NULL exclude_id disables the exclusion.
SELECT * FROM t_dept
WHERE code = sqlc.arg('code') AND is_deleted = 0
  AND (sqlc.narg('exclude_id') IS NULL OR id <> sqlc.narg('exclude_id'))
LIMIT 1;

-- name: ListDepts :many
SELECT * FROM t_dept
WHERE is_deleted = 0
  AND (sqlc.narg('keywords') IS NULL OR name LIKE sqlc.narg('keywords'))
  AND (sqlc.narg('status')   IS NULL OR status = sqlc.narg('status'))
ORDER BY sort ASC;

-- name: ListEnabledDepts :many
SELECT * FROM t_dept
WHERE is_deleted = 0 AND status = 1
ORDER BY sort ASC;

-- name: ListDeptsByIDs :many
SELECT * FROM t_dept
WHERE is_deleted = 0 AND id IN (sqlc.slice('ids'));

-- name: CreateDept :exec
INSERT INTO t_dept (
    id, name, code, parent_id, tree_path, sort, status,
    create_by, update_by, is_deleted, create_time, update_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('name'), sqlc.arg('code'), sqlc.arg('parent_id'),
    sqlc.arg('tree_path'), sqlc.arg('sort'), sqlc.arg('status'),
    sqlc.arg('create_by'), sqlc.arg('update_by'), sqlc.arg('is_deleted'),
    sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateDept :exec
-- Mirrors the previous Select-scoped GORM update
-- (name, code, parent_id, tree_path, sort, status, update_by).
UPDATE t_dept SET
    name        = sqlc.arg('name'),
    code        = sqlc.arg('code'),
    parent_id   = sqlc.arg('parent_id'),
    tree_path   = sqlc.arg('tree_path'),
    sort        = sqlc.arg('sort'),
    status      = sqlc.arg('status'),
    update_by   = sqlc.arg('update_by'),
    update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: SoftDeleteDept :exec
UPDATE t_dept SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id');

-- name: SoftDeleteDeptByTreePath :exec
-- Soft-deletes a department and all of its descendants. A node is a descendant
-- when its comma-wrapped tree_path contains ",<id>,".
UPDATE t_dept SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id')
   OR CONCAT(',', tree_path, ',') LIKE sqlc.arg('tree_path_like');
