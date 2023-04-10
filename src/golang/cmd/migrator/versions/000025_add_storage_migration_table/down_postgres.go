package _00025_add_storage_migration_table

const downPostgresScript = `
DROP TABLE IF EXISTS storage_migration;
`
