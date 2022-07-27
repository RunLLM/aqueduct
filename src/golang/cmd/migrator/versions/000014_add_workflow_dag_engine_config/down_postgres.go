package _000014_add_workflow_dag_engine_config

const downPostgresScript = `
ALTER TABLE workflow_dag
DROP COLUMN engine_config;`
