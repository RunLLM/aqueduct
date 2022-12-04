package _000021_add_gc_column_to_env_table

const downPostgresScript = `
ALTER TABLE execution_environment DROP COLUMN IF EXISTS garbage_collected;
`
