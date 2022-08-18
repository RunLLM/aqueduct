package _000016_add_artifact_type_column_to_artifact

const upPostgresAddColumn = `
ALTER TABLE artifact 
ADD COLUMN type VARCHAR NOT NULL DEFAULT 'untyped';
`

const upPostgresDropColumn = `
ALTER TABLE artifact 
DROP COLUMN spec;
`
