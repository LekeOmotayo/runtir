-- name: CreateOrganisation :one
INSERT INTO organisations (
    name,
    slug
) VALUES (
    $1, $2
)
RETURNING *;

-- name: GetOrganisationByID :one
SELECT *
FROM organisations
WHERE id = $1
LIMIT 1;

-- name: GetOrganisationBySlug :one
SELECT *
FROM organisations
WHERE slug = $1
LIMIT 1;

-- name: ListUserOrganisations :many
SELECT 
    o.id,
    o.name,
    o.slug,
    o.created_at,
    o.updated_at,
    m.role
FROM organisations o
JOIN memberships m ON o.id = m.organisation_id
WHERE m.user_id = $1
ORDER BY o.name ASC;
