package _000003_add_storage_column

const sqliteScript = `
ALTER TABLE workflow_dag
ADD COLUMN storage_config BLOB;
`
