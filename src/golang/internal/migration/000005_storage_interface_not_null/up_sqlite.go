package _000005_storage_interface_not_null

const sqliteScript = `
BEGIN TRANSACTION;

ALTER TABLE workflow_dag RENAME TO tmp_workflow_dag;

CREATE TABLE workflow_dag (
    id BLOB NOT NULL PRIMARY KEY,
    workflow_id BLOB NOT NULL REFERENCES workflow (id),
    s3_config BLOB NOT NULL,
    created_at DATETIME NOT NULL,
    storage_config BLOB NOT NULL
);

INSERT INTO workflow_dag(id, workflow_id, s3_config, created_at, storage_config)
SELECT id, workflow_id, s3_config, created_at, storage_config
FROM tmp_workflow_dag;

DROP TABLE tmp_workflow_dag;

COMMIT;
`
