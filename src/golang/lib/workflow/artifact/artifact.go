package artifact

import (
	"context"
	"fmt"
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
	if !a.Computed() {
		return errors.Newf(fmt.Sprintf("Artifact %s cannot be persisted because it has not been computed.", a.name))
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
	return nil
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
	if workflowDagResultID != uuid.Nil {
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
		db:                   db,
	}, nil
}
