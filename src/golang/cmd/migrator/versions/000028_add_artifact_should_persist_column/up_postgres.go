package _000028_add_artifact_should_persist_column

const upPostgresScript = `
ALTER TABLE artifact 
ADD COLUMN should_persist BOOL DEFAULT TRUE NOT NULL;
`
