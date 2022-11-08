package _000020_add_execution_environment_table

const upPostgresScript = `
CREATE TABLE IF NOT EXISTS execution_environment (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    spec JSONB NOT NULL,
    hash VARCHAR NOT NULL UNIQUE
);

ALTER TABLE operator 
ADD COLUMN execution_environment_id UUID REFERENCES execution_environment (id);
`
