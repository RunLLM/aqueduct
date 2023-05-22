package _00027_rename_integrations_table

const upPostgresScript = `
ALTER TABLE integration RENAME TO resource;
`
