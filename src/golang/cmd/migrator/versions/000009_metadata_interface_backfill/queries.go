package _000009_metadata_interface_backfill

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/google/uuid"
)

type artifactResultOldMetadataResponse struct {
	Id       uuid.UUID       `db:"id" json:"id"`
	Metadata OldNullMetadata `db:"metadata" json:"metadata"`
}

type artifactResultNewMetadataResponse struct {
	Id       uuid.UUID    `db:"id" json:"id"`
	Metadata NullMetadata `db:"metadata" json:"metadata"`
}

func getAllArtifactResultOldMetadata(
	ctx context.Context,
	db database.Database,
) ([]artifactResultOldMetadataResponse, error) {
	query := "SELECT id, metadata FROM artifact_result;"

	var response []artifactResultOldMetadataResponse
	err := db.Query(ctx, &response, query)
	return response, err
}

func getAllArtifactResultNewMetadata(
	ctx context.Context,
	db database.Database,
) ([]artifactResultNewMetadataResponse, error) {
	query := "SELECT id, metadata FROM artifact_result;"

	var response []artifactResultNewMetadataResponse
	err := db.Query(ctx, &response, query)
	return response, err
}

func updateArtifactResultAsNewMetadata(
	ctx context.Context,
	id uuid.UUID,
	metadata OldMetadata,
	db database.Database,
) error {
	newMetadata := &Metadata{
		Schema:        metadata,
		SystemMetrics: make(map[string]string),
	}

	changes := map[string]interface{}{
		"metadata": newMetadata,
	}

	return repos.UpdateRecord(ctx, changes, "artifact_result", "id", id, db)
}

func updateArtifactResultAsOldMetadata(
	ctx context.Context,
	id uuid.UUID,
	metadata Metadata,
	db database.Database,
) error {
	changes := map[string]interface{}{
		"metadata": metadata.Schema,
	}

	return repos.UpdateRecord(ctx, changes, "artifact_result", "id", id, db)
}
