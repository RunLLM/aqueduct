package _000010_add_exec_state_column

const downPostgresScript = `
ALTER TABLE operator_result DROP COLUMN IF EXISTS execution_state;
`
