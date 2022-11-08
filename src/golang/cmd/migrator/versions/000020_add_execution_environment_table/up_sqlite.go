package _000020_add_execution_environment_table

const sqliteAddTableScript = `
CREATE TABLE IF NOT EXISTS execution_environment (
    id BLOB NOT NULL PRIMARY KEY,
    spec BLOB NOT NULL,
    hash TEXT NOT NULL UNIQUE
);
`
