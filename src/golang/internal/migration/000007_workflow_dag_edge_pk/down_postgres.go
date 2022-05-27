package _000007_workflow_dag_edge_pk

const downPostgresScript = `ALTER TABLE workflow_dag_edge 
DROP CONSTRAINT workflow_dag_edge_pk;
`
