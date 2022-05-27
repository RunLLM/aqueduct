package _000008_delete_s3_config

const upPostgresScript = `
ALTER TABLE workflow_dag
DROP COLUMN IF EXISTS s3_config;
`
