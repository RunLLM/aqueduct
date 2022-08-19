package _000014_add_exec_state_column_to_artifact_result

const sqliteScript = `
ALTER TABLE artifact_result 
ADD COLUMN execution_state BLOB;
`
