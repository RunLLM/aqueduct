package _00027_rename_integrations_table

const downPostgresScript = `
ALTER TABLE resource RENAME TO integration;
`
