package _000005_storage_interface_not_null

const downPostgresScript = `
ALTER TABLE workflow_dag DROP CONSTRAINT storage_config_not_null;
`
