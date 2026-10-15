-- name: CreateDocument :one
INSERT INTO documents (
    organisation_id,
    uploaded_by,
    title,
    file_name,
    file_path,
    file_size,
    mime_type,
    file_hash,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetDocumentByID :one
SELECT *
FROM documents
WHERE id = $1 AND organisation_id = $2
LIMIT 1;

-- name: ListDocumentsByOrg :many
SELECT *
FROM documents
WHERE organisation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateDocumentStatus :one
UPDATE documents
SET 
    status = $3,
    error_message = $4,
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: DeleteDocument :exec
DELETE FROM documents
WHERE id = $1 AND organisation_id = $2;

-- name: CountDocumentsByOrg :one
SELECT COUNT(*)
FROM documents
WHERE organisation_id = $1;

-- name: CountDocumentsByStatus :many
SELECT status, COUNT(*) as count
FROM documents
WHERE organisation_id = $1
GROUP BY status;

-- name: CreateDocumentMetadata :one
INSERT INTO document_metadata (
    document_id,
    key,
    value
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetDocumentMetadata :many
SELECT *
FROM document_metadata
WHERE document_id = $1;
