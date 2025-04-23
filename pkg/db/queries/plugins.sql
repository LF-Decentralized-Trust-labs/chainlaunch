-- name: CreatePlugin :one
INSERT INTO plugins (
  name,
  api_version,
  kind,
  metadata,
  spec,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, NOW(), NOW()
) RETURNING *;

-- name: GetPluginByID :one
SELECT * FROM plugins
WHERE id = $1;

-- name: GetPluginByName :one
SELECT * FROM plugins
WHERE name = $1;

-- name: ListPlugins :many
SELECT * FROM plugins
ORDER BY name;

-- name: ListPluginsByKind :many
SELECT * FROM plugins
WHERE kind = $1
ORDER BY created_at DESC;

-- name: UpdatePlugin :one
UPDATE plugins
SET 
  api_version = $1,
  kind = $2,
  metadata = $3,
  spec = $4,
  updated_at = NOW()
WHERE name = $5
RETURNING *;

-- name: DeletePlugin :exec
DELETE FROM plugins
WHERE name = $1; 