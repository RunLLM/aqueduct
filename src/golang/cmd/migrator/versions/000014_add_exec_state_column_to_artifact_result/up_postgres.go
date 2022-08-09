package _00014_add_exec_state_column_to_artifact_result

const upPostgresScript = `
ALTER TABLE artifact_result 
ADD COLUMN execution_state JSONB;
`
