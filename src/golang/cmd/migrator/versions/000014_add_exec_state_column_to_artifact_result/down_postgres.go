package _000014_add_exec_state_column_to_artifact_result

const downPostgresScript = `
ALTER TABLE artifact_result DROP COLUMN IF EXISTS execution_state;
`
