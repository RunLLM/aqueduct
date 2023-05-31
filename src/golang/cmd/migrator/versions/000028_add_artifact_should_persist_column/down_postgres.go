package _000028_add_artifact_should_persist_column

const downPostgresScript = `
ALTER TABLE artifact DROP COLUMN IF EXISTS should_persist;
`
