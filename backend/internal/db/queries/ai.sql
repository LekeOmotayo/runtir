-- name: CreateConversation :one
INSERT INTO ai_conversations (
    organisation_id,
    user_id,
    title
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetConversationByID :one
SELECT *
FROM ai_conversations
WHERE id = $1 AND organisation_id = $2
LIMIT 1;

-- name: ListConversationsByUser :many
SELECT *
FROM ai_conversations
WHERE organisation_id = $1 AND user_id = $2
ORDER BY updated_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateConversationTitle :one
UPDATE ai_conversations
SET
    title = $3,
    updated_at = NOW()
WHERE id = $1 AND organisation_id = $2
RETURNING *;

-- name: TouchConversation :one
UPDATE ai_conversations
SET updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteConversation :exec
DELETE FROM ai_conversations
WHERE id = $1 AND organisation_id = $2;

-- name: CreateMessage :one
INSERT INTO ai_messages (
    conversation_id,
    sender_type,
    content,
    token_count
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListMessagesByConversation :many
SELECT *
FROM ai_messages
WHERE conversation_id = $1
ORDER BY created_at ASC;

-- name: CreateMessageSource :one
INSERT INTO ai_message_sources (
    message_id,
    chunk_id,
    document_id,
    relevance_score,
    citation_text
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListMessageSources :many
SELECT
    ms.id,
    ms.message_id,
    ms.chunk_id,
    ms.document_id,
    ms.relevance_score,
    ms.citation_text,
    ms.created_at,
    d.title AS document_title,
    d.file_name AS document_file_name,
    dc.chunk_index,
    dc.page_number
FROM ai_message_sources ms
LEFT JOIN documents d ON ms.document_id = d.id
LEFT JOIN document_chunks dc ON ms.chunk_id = dc.id
WHERE ms.message_id = $1
ORDER BY ms.relevance_score DESC;
