package _000018_add_dag_result_exec_state_column

const sqliteAddColScript = `
ALTER TABLE workflow_dag_result 
ADD COLUMN execution_state BLOB;
`
