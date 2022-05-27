package _000007_workflow_dag_edge_pk

const upPostgresScript = `ALTER TABLE workflow_dag_edge
ADD CONSTRAINT workflow_dag_edge_pk PRIMARY KEY (workflow_dag_id, from_id, to_id);
`
