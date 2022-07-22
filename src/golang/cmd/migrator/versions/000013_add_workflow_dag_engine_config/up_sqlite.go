package _000013_add_workflow_dag_engine_config

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

// setDefaultEngineConfig sets the `engine_config` column for the `workflow_dag` table
// to the default Aqueduct engine for all rows with a NULL `engine_config`.
func setDefaultEngineConfig(ctx context.Context, db database.Database) error {
	query := "UPDATE workflow_dag SET engine_config = $1 WHERE engine_config = NULL;"

	defaultEngineConfig := &EngineConfig{
		Type:           AqueductEngineType,
		AqueductConfig: &AqueductConfig{},
	}
	args := []interface{}{defaultEngineConfig}

	return db.Execute(ctx, query, args...)
}

// makeEngineConfigNotNull changes the `engine_config` column to be NOT NULL
func makeEngineConfigNotNull(ctx context.Context, db database.Database) error {
	query := `
	BEGIN TRANSACTION;

	ALTER TABLE workflow_dag RENAME TO tmp_workflow_dag;

	CREATE TABLE workflow_dag (
		id BLOB NOT NULL PRIMARY KEY,
		workflow_id BLOB NOT NULL REFERENCES workflow (id),
		s3_config BLOB NOT NULL,
		created_at DATETIME NOT NULL,
		storage_config BLOB NOT NULL,
		engine_config BLOB NOT NULL
	);

	INSERT INTO workflow_dag(id, workflow_id, s3_config, created_at, storage_config, engine_config)
	SELECT id, workflow_id, s3_config, created_at, storage_config, engine_config
	FROM tmp_workflow_dag;

	DROP TABLE tmp_workflow_dag;

	COMMIT;
	`

	return db.Execute(ctx, query)
}
