package _000021_add_gc_column_to_env_table

const upPostgresScript = `
DROP TABLE IF EXISTS execution_environment;

CREATE TABLE IF NOT EXISTS execution_environment (
    id BLOB NOT NULL PRIMARY KEY,
    spec BLOB NOT NULL,
    hash BLOB NOT NULL,
	garbage_collected BOOL DEFAULT FALSE NOT NULL
);
`
