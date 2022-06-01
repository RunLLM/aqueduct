package _000005_storage_interface_not_null

const upPostgresScript = `
ALTER TABLE workflow_dag ADD CONSTRAINT storage_config_not_null
CHECK (storage_config is NOT NULL) NOT VALID;

ALTER TABLE workflow_dag VALIDATE CONSTRAINT storage_config_not_null;
`
