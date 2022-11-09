package _000020_add_execution_environment_table

const upSqliteScript = `
CREATE TABLE IF NOT EXISTS execution_environment (
    id BLOB NOT NULL PRIMARY KEY,
    spec BLOB NOT NULL,
    hash BLOB NOT NULL UNIQUE
);

ALTER TABLE operator 
ADD COLUMN execution_environment_id BLOB REFERENCES execution_environment (id);
`
