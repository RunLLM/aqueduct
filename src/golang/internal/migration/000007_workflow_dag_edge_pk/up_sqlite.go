package _000007_workflow_dag_edge_pk

const sqliteScript = `
BEGIN TRANSACTION;

ALTER TABLE workflow_dag_edge RENAME TO tmp_workflow_dag_edge;

CREATE TABLE workflow_dag_edge (
    workflow_dag_id BLOB NOT NULL REFERENCES workflow_dag (id),
    type TEXT NOT NULL,
    from_id BLOB NOT NULL,
    to_id BLOB NOT NULL,
    idx INTEGER NOT NULL,
	PRIMARY KEY (workflow_dag_id, from_id, to_id)
);

INSERT INTO workflow_dag_edge(workflow_dag_id, type, from_id, to_id, idx)
SELECT workflow_dag_id, type, from_id, to_id, idx
FROM tmp_workflow_dag_edge;

DROP TABLE tmp_workflow_dag_edge;

COMMIT;
`
