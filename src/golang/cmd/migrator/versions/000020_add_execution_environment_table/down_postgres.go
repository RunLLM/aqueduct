package _000020_add_execution_environment_table

const downPostgresScript = `
ALTER TABLE operator DROP COLUMN IF EXISTS execution_environment_id;

DROP TABLE IF EXISTS execution_environment;
`
