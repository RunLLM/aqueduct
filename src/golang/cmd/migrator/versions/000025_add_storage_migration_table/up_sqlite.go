package _00025_add_storage_migration_table

const upSqliteScript = `
CREATE TABLE IF NOT EXISTS storage_migration (
	id BLOB NOT NULL PRIMARY KEY,
	src_integration_id BLOB REFERENCES integration (id),
	dest_integration_id BLOB REFERENCES integration (id),
	execution_state BLOB NOT NULL,
	current BOOL NOT NULL
);
`
