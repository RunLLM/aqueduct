package _000012_drop_metadata_column

const sqliteScript = `
ALTER TABLE operator_result
DROP COLUMN metadata;
`
