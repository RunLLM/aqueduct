package _000021_add_gc_column_to_env_table

const upSqliteScript = `
ALTER TABLE execution_environment ADD COLUMN garbage_collected BOOL DEFAULT FALSE NOT NULL;
`
