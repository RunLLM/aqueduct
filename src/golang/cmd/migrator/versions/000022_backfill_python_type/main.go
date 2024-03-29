package _000022_backfill_python_type

import (
	"context"
	"os"
	"path/filepath"

	"github.com/aqueducthq/aqueduct/config"
	"github.com/aqueducthq/aqueduct/lib/database"
	log "github.com/sirupsen/logrus"
)

var confPath = filepath.Join(os.Getenv("HOME"), ".aqueduct", "server", "config", "config.yml")

func Up(ctx context.Context, db database.Database) error {
	artifactResults, err := getAllArtifactResults(ctx, db)
	if err != nil {
		return err
	}

	if err := config.Init(confPath); err != nil {
		return err
	}

	storageConfig := config.Storage()

	for _, artifactResult := range artifactResults {
		if !artifactResult.Metadata.IsNull {
			err = backfillPythonType(
				ctx,
				artifactResult.Id,
				&(artifactResult.Metadata.Metadata),
				artifactResult.ContentPath,
				&storageConfig,
				db,
			)
			if err != nil {
				log.Errorf("Error backfilling Python type for artifact result %s: %v", artifactResult.Id, err)
			}
		}
	}

	return nil
}

func Down(ctx context.Context, db database.Database) error {
	// We do not support down migration for our OSS version, so this implementation is left empty.
	return nil
}
