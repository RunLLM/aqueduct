package _000003_add_storage_column

const downPostgresScript = `
ALTER TABLE workflow_dag DROP COLUMN IF EXISTS storage_config;
`
