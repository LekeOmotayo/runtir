-- name: CreateMembership :one
INSERT INTO memberships (
    organisation_id,
    user_id,
    role
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetMembership :one
SELECT *
FROM memberships
WHERE organisation_id = $1 AND user_id = $2
LIMIT 1;

-- name: ListMembershipsByOrg :many
SELECT 
    m.id,
    m.organisation_id,
    m.user_id,
    m.role,
    m.created_at,
    m.updated_at,
    u.email,
    u.full_name,
    u.avatar_url
FROM memberships m
JOIN users u ON m.user_id = u.id
WHERE m.organisation_id = $1
ORDER BY m.created_at ASC;

-- name: UpdateMembershipRole :one
UPDATE memberships
SET 
    role = $3,
    updated_at = NOW()
WHERE organisation_id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteMembership :exec
DELETE FROM memberships
WHERE organisation_id = $1 AND user_id = $2;
