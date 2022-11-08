package _000020_add_execution_environment_table

const postgresAddTableScript = `
CREATE TABLE IF NOT EXISTS execution_environment (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    spec JSONB NOT NULL,
    hash VARCHAR NOT NULL UNIQUE
);
`
