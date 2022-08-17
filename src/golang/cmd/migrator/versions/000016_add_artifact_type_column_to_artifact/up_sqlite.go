package _000016_add_artifact_type_column_to_artifact

const sqliteAddColumn = `
ALTER TABLE artifact
ADD COLUMN type TEXT;
`

const sqliteDropColumn = `
ALTER TABLE artifact 
DROP COLUMN spec;
`
