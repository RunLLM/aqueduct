package _000006_add_retention_policy_column

const sqliteScript = `
ALTER TABLE workflow
ADD COLUMN retention_policy BLOB NOT NULL;
`
