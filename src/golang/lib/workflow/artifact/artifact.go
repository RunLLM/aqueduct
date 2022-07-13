package artifact

import (
	"context"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// This Artifact interface allows a caller to manage and inspect the lifecycle
// of a single artifact in a workflow run.
type Artifact interface {
	ID() uuid.UUID
	Type() artifact.Type
	Name() string

	// Indicates whether this artifact has been computed or not. An artifact is
	// only considered computed if the operator that generates it has completed
	// successfully.
	Computed() bool

	// Writes the data of this artifact to a backing store so it can be fetched later.
	// Errors if the artifact has not yet been computed.
	PersistResult(opStatus shared.ExecutionStatus) error

	Finish(ctx context.Context)
}

func initializeArtifactResultInDatabase(
	ctx context.Context,
	artifactID uuid.UUID,
	workflowDagResultID uuid.UUID,
	artifactResultWriter artifact_result.Writer,
	contentPath string,
	db database.Database,
) (uuid.UUID, error) {
	artifactResult, err := artifactResultWriter.CreateArtifactResult(
		ctx,
		workflowDagResultID,
		artifactID,
		contentPath,
		db,
	)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Failed to create operator result record.")
	}
	return artifactResult.Id, nil
}

type ArtifactImpl struct {
	ctx          context.Context
	id           uuid.UUID
	name         string
	description  string
	artifactType artifact.Type

	contentPath  string
	metadataPath string

	artifactResultWriter artifact_result.Writer
	artifactResultID     uuid.UUID

	//jobManager job.JobManager
	//vaultObject vault.Vault
	storageConfig *shared.StorageConfig
	db            database.Database

	persisted bool
}

func (a *ArtifactImpl) ID() uuid.UUID {
	// TODO(kenxu)
}

func (a *ArtifactImpl) Type() artifact.Type {
	// TODO(kenxu)
}

func (a *ArtifactImpl) Name() string {
	// TODO(kenxu)
}

func (a *ArtifactImpl) Computed() bool {
	// TODO(kenxu):
}

func (a *ArtifactImpl) PersistResult(opStatus shared.ExecutionStatus) error {
	if a.persisted {
		return errors.Newf("Artifact %s was already persisted!", a.name)
	}
	if !a.Computed() {
		return errors.Newf("Artifact %s cannot be persisted because it has not been computed.", a.name)
	}
	utils.UpdateArtifactResultAfterComputation(
		a.ctx,
		opStatus,
		a.storageConfig,
		a.metadataPath,
		a.artifactResultWriter,
		a.artifactResultID,
		a.db,
	)
	a.persisted = true
	return nil
}

func (a *ArtifactImpl) Finish(ctx context.Context) {
	utils.CleanupStorageFile(ctx, a.storageConfig, a.metadataPath)

	// If the artifact was persisted to the DB, don't cleanup the content paths,
	// since we may need that data later.
	if !a.persisted {
		utils.CleanupStorageFile(ctx, a.storageConfig, a.contentPath)
	}
}

func NewArtifact(
	ctx context.Context,
	dbArtifact artifact.DBArtifact,
	contentPath string,
	metadataPath string,
	// A nil value here means the artifact is not persisted.
	artifactResultWriter artifact_result.Writer,
	workflowDagResultID uuid.UUID,
	db database.Database,
) (Artifact, error) {
	var artifactResultID uuid.UUID

	canPersist := workflowDagResultID != uuid.Nil
	if canPersist {
		var err error
		artifactResultID, err = initializeArtifactResultInDatabase(ctx, dbArtifact.Id, workflowDagResultID, artifactResultWriter, contentPath, db)
		if err != nil {
			return nil, err
		}
	}

	return &ArtifactImpl{
		ctx:                  ctx,
		id:                   dbArtifact.Id,
		name:                 dbArtifact.Name,
		description:          dbArtifact.Description,
		artifactType:         dbArtifact.Spec.Type(),
		contentPath:          contentPath,
		metadataPath:         metadataPath,
		artifactResultID:     artifactResultID,
		artifactResultWriter: artifactResultWriter,
		persisted:            false,
		db:                   db,
	}, nil
}
