-- name: CreateActivityLog :one
INSERT INTO activity_logs (
    organisation_id,
    user_id,
    action,
    entity_type,
    entity_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListActivityLogsByOrg :many
SELECT *
FROM activity_logs
WHERE organisation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
