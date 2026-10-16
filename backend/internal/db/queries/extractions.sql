-- name: CreateExtraction :one
INSERT INTO document_extractions (
    document_id,
    raw_text,
    page_count,
    word_count,
    extraction_method
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetExtractionByDocumentID :one
SELECT *
FROM document_extractions
WHERE document_id = $1
LIMIT 1;

-- name: CreateChunk :one
INSERT INTO document_chunks (
    document_id,
    organisation_id,
    chunk_index,
    content,
    token_count,
    page_number
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListChunksByDocument :many
SELECT *
FROM document_chunks
WHERE document_id = $1 AND organisation_id = $2
ORDER BY chunk_index ASC;

-- name: SearchChunksByContent :many
SELECT *
FROM document_chunks
WHERE organisation_id = $1
  AND content ILIKE '%' || $2 || '%'
ORDER BY document_id, chunk_index
LIMIT $3 OFFSET $4;

-- name: SearchChunksWithDocument :many
SELECT 
    c.id AS chunk_id,
    c.document_id,
    c.chunk_index,
    c.token_count,
    c.page_number,
    c.content,
    d.title AS document_title,
    d.file_name AS document_file_name,
    d.mime_type AS document_mime_type
FROM document_chunks c
JOIN documents d ON c.document_id = d.id
WHERE c.organisation_id = $1
  AND d.is_archived = false
  AND c.content ILIKE '%' || $2 || '%'
ORDER BY c.document_id, c.chunk_index ASC
LIMIT $3 OFFSET $4;

-- name: CountSearchChunksWithDocument :one
SELECT COUNT(*)
FROM document_chunks c
JOIN documents d ON c.document_id = d.id
WHERE c.organisation_id = $1
  AND d.is_archived = false
  AND c.content ILIKE '%' || $2 || '%';

-- name: CountChunksByDocument :one
SELECT COUNT(*)
FROM document_chunks
WHERE document_id = $1;

-- name: DeleteChunksByDocument :exec
DELETE FROM document_chunks
WHERE document_id = $1;

-- name: CreateEmbedding :one
INSERT INTO document_embeddings (
    chunk_id,
    embedding_model,
    embedding
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetEmbeddingByChunkID :one
SELECT *
FROM document_embeddings
WHERE chunk_id = $1
LIMIT 1;
