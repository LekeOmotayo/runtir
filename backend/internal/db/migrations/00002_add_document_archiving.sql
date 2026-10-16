-- +goose Up
-- +goose StatementBegin
ALTER TABLE documents 
ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_documents_org_archived ON documents(organisation_id, is_archived);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_documents_org_archived;
ALTER TABLE documents 
DROP COLUMN IF EXISTS archived_at,
DROP COLUMN IF EXISTS is_archived;
-- +goose StatementEnd
