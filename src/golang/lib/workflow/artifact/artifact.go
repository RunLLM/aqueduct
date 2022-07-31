package artifact

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Artifact is an interface for managing and inspect the lifecycle of an artifact
// produced by a workflow run.
type Artifact interface {
	ID() uuid.UUID
	Type() artifact.Type
	Name() string

	// InitializeResult initializes the artifact in the database.
	InitializeResult(ctx context.Context, dagResultID uuid.UUID) error

	// PersistResult updates the artifact result in the database.
	// Errors if InitializeResult() hasn't been called yet.
	PersistResult(ctx context.Context, execState *shared.ExecutionState) error

	// Finish is an end-of-lifecycle hook meant to do any final cleanup work.
	Finish(ctx context.Context)

	// Computed indicates whether this artifact has been computed or not.
	// An artifact is only considered "computed" if its results have been written to storage.
	Computed(ctx context.Context) bool

	// GetMetadata fetches the metadata for this artifact.
	// Errors if the artifact has not yet been computed.
	GetMetadata(ctx context.Context) (*artifact_result.Metadata, error)

	// GetContent fetches the content of this artifact.
	// Errors if the artifact has not yet been computed.
	GetContent(ctx context.Context) ([]byte, error)
}

type ArtifactImpl struct {
	id           uuid.UUID
	name         string
	description  string
	artifactType artifact.Type

	contentPath  string
	metadataPath string

	resultWriter     artifact_result.Writer
	resultID         uuid.UUID
	resultsPersisted bool

	storageConfig *shared.StorageConfig
	db            database.Database
}

func NewArtifact(
	dbArtifact artifact.DBArtifact,
	contentPath string,
	metadataPath string,
	artifactResultWriter artifact_result.Writer,
	storageConfig *shared.StorageConfig,
	db database.Database,
) (Artifact, error) {
	return &ArtifactImpl{
		id:               dbArtifact.Id,
		name:             dbArtifact.Name,
		description:      dbArtifact.Description,
		artifactType:     dbArtifact.Spec.Type,
		contentPath:      contentPath,
		metadataPath:     metadataPath,
		resultWriter:     artifactResultWriter,
		resultID:         uuid.Nil,
		resultsPersisted: false,
		storageConfig:    storageConfig,
		db:               db,
	}, nil
}

func (a *ArtifactImpl) ID() uuid.UUID {
	return a.id
}

func (a *ArtifactImpl) Type() artifact.Type {
	return a.artifactType
}

func (a *ArtifactImpl) Name() string {
	return a.name
}

func (a *ArtifactImpl) Computed(ctx context.Context) bool {
	// An artifact is only considered computed if its metadata path has been populated.
	res := utils.ObjectExistsInStorage(
		ctx,
		a.storageConfig,
		a.metadataPath,
	)
	return res
}

func (a *ArtifactImpl) InitializeResult(ctx context.Context, dagResultID uuid.UUID) error {
	if a.resultWriter == nil {
		return errors.New("Artifact's result writer cannot be nil.")
	}

	artifactResult, err := a.resultWriter.CreateArtifactResult(
		ctx,
		dagResultID,
		a.ID(),
		a.contentPath,
		a.db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create operator result record.")
	}

	a.resultID = artifactResult.Id
	return nil
}

func (a *ArtifactImpl) updateArtifactResultAfterComputation(
	ctx context.Context,
	execState *shared.ExecutionState,
) {
	changes := map[string]interface{}{
		artifact_result.StatusColumn:    execState.Status,
		artifact_result.ExecStateColumn: execState,
		artifact_result.MetadataColumn:  nil,
	}

	if a.Computed(ctx) {
		var artifactResultMetadata artifact_result.Metadata
		err := utils.ReadFromStorage(
			ctx,
			a.storageConfig,
			a.metadataPath,
			&artifactResultMetadata,
		)
		if err != nil {
			log.Errorf("Unable to read artifact result metadata from storage and unmarshal: %v", err)
			return
		}
		changes[artifact_result.MetadataColumn] = &artifactResultMetadata
	}

	_, err := a.resultWriter.UpdateArtifactResult(
		ctx,
		a.resultID,
		changes,
		a.db,
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"changes": changes,
			},
		).Errorf("Unable to update artifact result metadata: %v", err)
	}
}

func (a *ArtifactImpl) PersistResult(ctx context.Context, execState *shared.ExecutionState) error {
	if a.resultsPersisted {
		return errors.Newf("Artifact %s was already persisted!", a.name)
	}
	if execState.Status != shared.FailedExecutionStatus && execState.Status != shared.SucceededExecutionStatus {
		return errors.Newf("Artifact %s has unexpected execution state: %s", a.Name(), execState.Status)
	}

	a.updateArtifactResultAfterComputation(ctx, execState)
	a.resultsPersisted = true
	return nil
}

func (a *ArtifactImpl) Finish(ctx context.Context) {
	// There is nothing to clean up if the artifact was never computed.
	if !a.Computed(ctx) {
		return
	}

	utils.CleanupStorageFile(ctx, a.storageConfig, a.metadataPath)

	// If the artifact was persisted to the DB, don't cleanup the storage content paths,
	// since we may need that data later.
	if !a.resultsPersisted {
		utils.CleanupStorageFile(ctx, a.storageConfig, a.contentPath)
	}
}

func (a *ArtifactImpl) GetMetadata(ctx context.Context) (*artifact_result.Metadata, error) {
	if !a.Computed(ctx) {
		return nil, errors.Newf("Cannot get metadata of Artifact %s, it has not yet been computed.", a.Name())
	}

	var metadata artifact_result.Metadata
	err := utils.ReadFromStorage(ctx, a.storageConfig, a.metadataPath, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (a *ArtifactImpl) GetContent(ctx context.Context) ([]byte, error) {
	if !a.Computed(ctx) {
		return nil, errors.Newf("Cannot get content of Artifact %s, it has not yet been computed.", a.Name())
	}
	content, err := storage.NewStorage(a.storageConfig).Get(ctx, a.contentPath)
	if err != nil {
		return nil, err
	}
	return content, nil
}
