package _000008_delete_s3_config

const downPostgresScript = `
ALTER TABLE workflow_dag ADD COLUMN s3_config JSONB;
`
