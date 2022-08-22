package _000016_add_artifact_type_column_to_artifact

const downPostgresScript = `
ALTER TABLE artifact DROP COLUMN IF EXISTS type;

ALTER TABLE artifact ADD COLUMN spec JSONB NOT NULL;
`
