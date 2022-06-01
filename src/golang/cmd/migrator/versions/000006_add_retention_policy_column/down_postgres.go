package _000006_add_retention_policy_column

const downPostgresScript = `
ALTER TABLE workflow DROP COLUMN IF EXISTS retention_policy;
`
