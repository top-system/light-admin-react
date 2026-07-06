-- log.sql
-- SQL for the operation audit log (sys_log). Dynamic filters use sqlc.narg;
-- date-range bounds are passed as timestamptz (parsed in Go). Ordering is fixed
-- to create_time DESC — the previous GORM default.

-- name: GetLog :one
SELECT * FROM sys_log WHERE id = $1;

-- name: ListLogs :many
SELECT * FROM sys_log
WHERE (sqlc.narg('module')::text   IS NULL OR module = sqlc.narg('module'))
  AND (sqlc.narg('keywords')::text IS NULL OR (content LIKE sqlc.narg('keywords') OR request_uri LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from')::timestamptz IS NULL OR create_time >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')::timestamptz   IS NULL OR create_time <= sqlc.narg('create_to'))
ORDER BY create_time DESC
LIMIT  sqlc.narg('limit')
OFFSET sqlc.narg('offset');

-- name: CountLogs :one
SELECT COUNT(*) FROM sys_log
WHERE (sqlc.narg('module')::text   IS NULL OR module = sqlc.narg('module'))
  AND (sqlc.narg('keywords')::text IS NULL OR (content LIKE sqlc.narg('keywords') OR request_uri LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from')::timestamptz IS NULL OR create_time >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')::timestamptz   IS NULL OR create_time <= sqlc.narg('create_to'));

-- name: CreateLog :exec
INSERT INTO sys_log (
    id, module, request_method, request_params, response_content, content,
    request_uri, method, ip, province, city, execution_time,
    browser, browser_version, os, create_by, create_time
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11, $12,
    $13, $14, $15, $16, NOW()
);

-- name: DeleteLog :exec
DELETE FROM sys_log WHERE id = $1;

-- name: BatchDeleteLogs :exec
DELETE FROM sys_log WHERE id = ANY(@ids::text[]);
