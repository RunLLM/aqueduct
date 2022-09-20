package _000018_add_dag_result_exec_state_column

const downPostgresScript = `
ALTER TABLE workflow_dag_result DROP COLUMN IF EXISTS execution_state;
`
