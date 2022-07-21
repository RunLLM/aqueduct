package _000013_add_workflow_dag_engine_config

const sqliteScript = `
ALTER TABLE workflow_dag
ADD COLUMN engine_config BLOB NOT NULL
DEFAULT '{"type":"aqueduct", "aqueduct_config":{}}';
`
