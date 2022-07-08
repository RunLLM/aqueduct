package _000012_drop_metadata_column

const downPostgresScript = `
ALTER TABLE op_result ADD COLUMN metadata JSONB;
`
