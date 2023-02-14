package _000004_storage_interface_backfill

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

type s3ConfigResponse struct {
	Id       uuid.UUID `db:"id" json:"id"`
	S3Config s3Config  `db:"s3_config" json:"s3_config"`
}

type operatorSpecResponse struct {
	Id   uuid.UUID `db:"id" json:"id"`
	Spec spec      `db:"spec" json:"spec"`
}

func getAllS3Configs(
	ctx context.Context,
	db database.Database,
) ([]s3ConfigResponse, error) {
	query := "SELECT id, s3_config FROM workflow_dag;"

	var response []s3ConfigResponse
	err := db.Query(ctx, &response, query)
	return response, err
}

func getAllOperatorSpecs(
	ctx context.Context,
	db database.Database,
) ([]operatorSpecResponse, error) {
	query := "SELECT id, spec FROM operator;"

	var response []operatorSpecResponse
	err := db.Query(ctx, &response, query)
	return response, err
}

func updateStorageConfig(
	ctx context.Context,
	id uuid.UUID,
	storageConfig storageConfig,
	db database.Database,
) error {
	changes := map[string]interface{}{
		"storage_config": &storageConfig,
	}
	return repos.UpdateRecord(ctx, changes, "workflow_dag", "id", id, db)
}

func updateOperatorSpec(
	ctx context.Context,
	id uuid.UUID,
	spec spec,
	db database.Database,
) error {
	changes := map[string]interface{}{
		"spec": &spec,
	}
	return repos.UpdateRecord(ctx, changes, "operator", "id", id, db)
}

func setStorageConfigToNull(
	ctx context.Context,
	db database.Database,
) error {
	query := "UPDATE workflow_dag SET storage_config = NULL;"
	return db.Execute(ctx, query)
}
