package _000016_add_artifact_type_column_to_artifact

const upPostgresScript = `
ALTER TABLE artifact 
ADD COLUMN type VARCHAR;

ALTER TABLE artifact 
DROP COLUMN spec;
`
