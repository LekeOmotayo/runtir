-- name: CreateProcessingJob :one
INSERT INTO processing_jobs (
    document_id,
    organisation_id,
    job_type,
    status,
    max_attempts
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetProcessingJobByID :one
SELECT *
FROM processing_jobs
WHERE id = $1
LIMIT 1;

-- name: ListProcessingJobsByDoc :many
SELECT *
FROM processing_jobs
WHERE document_id = $1 AND organisation_id = $2
ORDER BY created_at ASC;

-- name: UpdateProcessingJobStatus :one
UPDATE processing_jobs
SET 
    status = $2,
    attempts = $3,
    error_message = $4,
    started_at = $5,
    completed_at = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ResetProcessingJob :one
UPDATE processing_jobs
SET 
    status = 'queued',
    error_message = NULL,
    started_at = NULL,
    completed_at = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
