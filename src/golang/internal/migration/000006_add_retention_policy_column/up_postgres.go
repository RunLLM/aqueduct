package _000006_add_retention_policy_column

const upPostgresScript = `
ALTER TABLE workflow
ADD COLUMN retention_policy JSONB NOT NULL DEFAULT '{"k_latest_runs": -1}'::json;
`
