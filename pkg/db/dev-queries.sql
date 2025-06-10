-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (name, description, boilerplate, slug) VALUES (?, ?, ?, ?) RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;

-- name: GetProject :one
SELECT * FROM projects WHERE id = ?;

-- name: GetProjectBySlug :one
SELECT * FROM projects WHERE slug = ?;

-- name: CreateConversation :one
INSERT INTO conversations (project_id) VALUES (?) RETURNING *;

-- name: GetDefaultConversationForProject :one
SELECT * FROM conversations WHERE project_id = ? ORDER BY started_at ASC LIMIT 1;

-- name: InsertMessage :one
INSERT INTO messages (conversation_id, parent_id, sender, content) VALUES (?, ?, ?, ?) RETURNING *;

-- name: ListMessagesForConversation :many
SELECT * FROM messages WHERE conversation_id = ? ORDER BY created_at ASC;

-- name: ListConversationsForProject :many
SELECT * FROM conversations WHERE project_id = ? ORDER BY started_at ASC;

-- name: InsertToolCall :one
INSERT INTO tool_calls (message_id, tool_name, arguments, result, error)
VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: ListToolCallsForMessage :many
SELECT * FROM tool_calls WHERE message_id = ? ORDER BY created_at ASC;

-- name: ListToolCallsForConversation :many
SELECT tc.* FROM tool_calls tc
JOIN messages m ON tc.message_id = m.id
WHERE m.conversation_id = ?
ORDER BY tc.created_at ASC;

-- name: UpdateProjectContainerInfo :exec
UPDATE projects
SET
  container_id = ?,
  container_name = ?,
  status = ?,
  last_started_at = ?,
  last_stopped_at = ?,
  container_port = ?
WHERE id = ?;
