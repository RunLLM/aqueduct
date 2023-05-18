package _00027_rename_integrations_table

const sqliteScript = `
ALTER TABLE integration RENAME TO resource;
`
