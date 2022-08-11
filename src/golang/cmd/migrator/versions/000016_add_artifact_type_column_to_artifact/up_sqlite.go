package _000016_add_artifact_type_column_to_artifact

const sqliteScript = `
ALTER TABLE artifact
ADD COLUMN type TEXT;

ALTER TABLE artifact 
DROP COLUMN spec;
`
