package _000012_drop_metadata_column

const upPostgresScript = `
ALTER TABLE operator_result
DROP COLUMN IF EXISTS metadata;
`
