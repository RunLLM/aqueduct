package _000003_add_storage_column

const upPostgresScript = `
ALTER TABLE workflow_dag
ADD COLUMN storage_config JSONB;
`
