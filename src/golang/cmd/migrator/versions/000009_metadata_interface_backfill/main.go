package _000009_metadata_interface_backfill

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
)

func Up(ctx context.Context, db database.Database) error {
	resultMetadata, err := getAllArtifactResultOldMetadata(ctx, db)
	if err != nil {
		return err
	}

	for _, artifactResultMetadata := range resultMetadata {
		if artifactResultMetadata.Metadata.IsNull {
			continue
		}
		err := updateArtifactResultAsNewMetadata(ctx, artifactResultMetadata.Id, artifactResultMetadata.Metadata.OldMetadata, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func Down(ctx context.Context, db database.Database) error {
	resultMetadata, err := getAllArtifactResultNewMetadata(ctx, db)
	if err != nil {
		return err
	}

	for _, artifactResultMetadata := range resultMetadata {
		if artifactResultMetadata.Metadata.IsNull {
			continue
		}
		err := updateArtifactResultAsOldMetadata(ctx, artifactResultMetadata.Id, artifactResultMetadata.Metadata.Metadata, db)
		if err != nil {
			return err
		}
	}

	return nil
}
