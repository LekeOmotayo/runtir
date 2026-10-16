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

-- name: ListDocumentsFiltered :many
SELECT *
FROM documents
WHERE organisation_id = $1
  AND is_archived = $2
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('mime_type')::text IS NULL OR mime_type = sqlc.narg('mime_type'))
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountDocumentsFiltered :one
SELECT COUNT(*)
FROM documents
WHERE organisation_id = $1
  AND is_archived = $2
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('mime_type')::text IS NULL OR mime_type = sqlc.narg('mime_type'));

-- name: UpdateDocumentStatus :one
UPDATE documents
SET 
    status = $3,
    error_message = $4,
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: UpdateDocumentTitle :one
UPDATE documents
SET 
    title = $3,
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: ArchiveDocument :one
UPDATE documents
SET 
    is_archived = true,
    archived_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: UnarchiveDocument :one
UPDATE documents
SET 
    is_archived = false,
    archived_at = NULL,
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

-- name: GetRecentDocuments :many
SELECT *
FROM documents
WHERE organisation_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: SearchDocuments :many
SELECT *
FROM documents
WHERE organisation_id = $1
  AND (
    title ILIKE '%' || $2 || '%'
    OR file_name ILIKE '%' || $2 || '%'
  )
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountSearchDocuments :one
SELECT COUNT(*)
FROM documents
WHERE organisation_id = $1
  AND (
    title ILIKE '%' || $2 || '%'
    OR file_name ILIKE '%' || $2 || '%'
  );

-- name: CreateDocumentMetadata :one
INSERT INTO document_metadata (
    document_id,
    key,
    value
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpsertDocumentMetadata :one
INSERT INTO document_metadata (
    document_id,
    key,
    value
) VALUES (
    $1, $2, $3
)
ON CONFLICT (document_id, key) DO UPDATE
SET value = EXCLUDED.value
RETURNING *;

-- name: GetDocumentMetadata :many
SELECT *
FROM document_metadata
WHERE document_id = $1;

-- name: DeleteDocumentMetadataByKey :exec
DELETE FROM document_metadata
WHERE document_id = $1 AND key = $2;

-- name: DeleteAllDocumentMetadata :exec
DELETE FROM document_metadata
WHERE document_id = $1;
