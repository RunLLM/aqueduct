package _00025_add_storage_migration_table

const upPostgresScript = `
CREATE TABLE IF NOT EXISTS storage_migration (
	id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	dest_integration_id UUID REFERENCES integration (id),
	execution_state JSONB NOT NULL,
	current BOOLEAN NOT NULL
);
`
