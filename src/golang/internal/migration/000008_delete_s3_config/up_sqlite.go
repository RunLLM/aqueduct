package _000008_delete_s3_config

const sqliteScript = `
ALTER TABLE workflow_dag
DROP COLUMN s3_config;
`
