-- dept.sql
-- All SQL for the department domain (t_dept). Dynamic filters use sqlc.narg
-- (nullable args) so one prepared statement serves every filter combination while
-- staying fully parameterized. Ordering is a fixed `sort ASC` (the previous GORM
-- default), eliminating any ORDER BY injection surface.

-- name: GetDept :one
SELECT * FROM t_dept
WHERE id = $1 AND is_deleted = 0;

-- name: GetDeptByCode :one
-- exclude_id lets callers ignore a specific row (used when validating a code on
-- update). A NULL exclude_id disables the exclusion.
SELECT * FROM t_dept
WHERE code = $1 AND is_deleted = 0
  AND (sqlc.narg('exclude_id')::text IS NULL OR id <> sqlc.narg('exclude_id'))
LIMIT 1;

-- name: ListDepts :many
SELECT * FROM t_dept
WHERE is_deleted = 0
  AND (sqlc.narg('keywords')::text IS NULL OR name LIKE sqlc.narg('keywords'))
  AND (sqlc.narg('status')::int    IS NULL OR status = sqlc.narg('status'))
ORDER BY sort ASC;

-- name: ListEnabledDepts :many
SELECT * FROM t_dept
WHERE is_deleted = 0 AND status = 1
ORDER BY sort ASC;

-- name: ListDeptsByIDs :many
SELECT * FROM t_dept
WHERE is_deleted = 0 AND id = ANY(@ids::text[]);

-- name: CreateDept :exec
INSERT INTO t_dept (
    id, name, code, parent_id, tree_path, sort, status,
    create_by, update_by, is_deleted, create_time, update_time
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, sqlc.arg('now'), sqlc.arg('now')
);

-- name: UpdateDept :exec
-- Mirrors the previous Select-scoped GORM update
-- (name, code, parent_id, tree_path, sort, status, update_by).
UPDATE t_dept SET
    name        = $2,
    code        = $3,
    parent_id   = $4,
    tree_path   = $5,
    sort        = $6,
    status      = $7,
    update_by   = $8,
    update_time = sqlc.arg('now')
WHERE id = $1;

-- name: SoftDeleteDept :exec
UPDATE t_dept SET is_deleted = 1, update_by = $2, update_time = sqlc.arg('now')
WHERE id = $1;

-- name: SoftDeleteDeptByTreePath :exec
-- Soft-deletes a department and all of its descendants. A node is a descendant
-- when its comma-wrapped tree_path contains ",<id>,". Matches the previous GORM
-- expression built from DBCompat.TreePathLike.
UPDATE t_dept SET is_deleted = 1, update_by = sqlc.arg('update_by'), update_time = sqlc.arg('now')
WHERE id = sqlc.arg('id')
   OR (',' || tree_path || ',') LIKE sqlc.arg('tree_path_like');
