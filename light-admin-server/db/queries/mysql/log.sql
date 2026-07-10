-- log.sql (MySQL dialect)
-- Derived from queries/postgres/log.sql. Every parameter is a named
-- sqlc.arg/narg (see config.sql header).

-- name: GetLog :one
SELECT * FROM sys_log WHERE id = sqlc.arg('id');

-- name: ListLogs :many
SELECT * FROM sys_log
WHERE (sqlc.narg('module')   IS NULL OR module = sqlc.narg('module'))
  AND (sqlc.narg('keywords') IS NULL OR (content LIKE sqlc.narg('keywords') OR request_uri LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from') IS NULL OR create_time >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')   IS NULL OR create_time <= sqlc.narg('create_to'))
ORDER BY create_time DESC
LIMIT  ?
OFFSET ?;

-- name: CountLogs :one
SELECT COUNT(*) FROM sys_log
WHERE (sqlc.narg('module')   IS NULL OR module = sqlc.narg('module'))
  AND (sqlc.narg('keywords') IS NULL OR (content LIKE sqlc.narg('keywords') OR request_uri LIKE sqlc.narg('keywords')))
  AND (sqlc.narg('create_from') IS NULL OR create_time >= sqlc.narg('create_from'))
  AND (sqlc.narg('create_to')   IS NULL OR create_time <= sqlc.narg('create_to'));

-- name: CreateLog :exec
INSERT INTO sys_log (
    id, module, request_method, request_params, response_content, content,
    request_uri, method, ip, province, city, execution_time,
    browser, browser_version, os, create_by, create_time
) VALUES (
    sqlc.arg('id'), sqlc.arg('module'), sqlc.arg('request_method'), sqlc.arg('request_params'),
    sqlc.arg('response_content'), sqlc.arg('content'),
    sqlc.arg('request_uri'), sqlc.arg('method'), sqlc.arg('ip'), sqlc.arg('province'),
    sqlc.arg('city'), sqlc.arg('execution_time'),
    sqlc.arg('browser'), sqlc.arg('browser_version'), sqlc.arg('os'), sqlc.arg('create_by'),
    sqlc.arg('now')
);

-- name: DeleteLog :exec
DELETE FROM sys_log WHERE id = sqlc.arg('id');

-- name: BatchDeleteLogs :exec
DELETE FROM sys_log WHERE id IN (sqlc.slice('ids'));
