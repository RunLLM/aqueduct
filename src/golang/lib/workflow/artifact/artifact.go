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
	LogicalID() uuid.UUID
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
	// This is *NOT* the same as having the operator's execution state == SUCCEEDED. For example,
	// for check operators, the artifact could have been computed even if the check operator did not
	// pass (returned false).
	Computed(ctx context.Context) bool

	// GetMetadata fetches the metadata for this artifact.
	// Errors if the artifact has not yet been computed.
	GetMetadata(ctx context.Context) (*artifact_result.Metadata, error)

	// GetContent fetches the content of this artifact.
	// Errors if the artifact has not yet been computed.
	GetContent(ctx context.Context) ([]byte, error)
}

type ArtifactImpl struct {
	// This is the ID that will be stored in our database. It is the canonical handle
	// to this artifact throughout our system.
	id uuid.UUID

	// This is a special ID that also encodes important structural/parameter information.
	// This can be used as a unique handle to a specific artifact **for a specific dag execution**.
	logicalID uuid.UUID

	name         string
	description  string
	artifactType artifact.Type

	execPaths *utils.ExecPaths

	resultWriter artifact_result.Writer
	resultID     uuid.UUID

	// If this is not nil, this artifact should be written to the cache.
	// An artifact cannot be both cache-aware and persisted.
	previewArtifactCacheManager PreviewCacheManager
	resultsPersisted            bool

	storageConfig *shared.StorageConfig
	db            database.Database
}

func NewArtifact(
	logicalID uuid.UUID,
	dbArtifact artifact.DBArtifact,
	execPaths *utils.ExecPaths,
	artifactResultWriter artifact_result.Writer,
	storageConfig *shared.StorageConfig,
	artifactCacheManager PreviewCacheManager,
	db database.Database,
) (Artifact, error) {
	if artifactCacheManager != nil && logicalID == uuid.Nil {
		return nil, errors.Newf("Logical ID must be provided for a cache-aware artifact.")
	}

	return &ArtifactImpl{
		id:                          dbArtifact.Id,
		logicalID:                   logicalID,
		name:                        dbArtifact.Name,
		description:                 dbArtifact.Description,
		artifactType:                dbArtifact.Spec.Type(),
		execPaths:                   execPaths,
		resultWriter:                artifactResultWriter,
		resultID:                    uuid.Nil,
		previewArtifactCacheManager: artifactCacheManager,
		resultsPersisted:            false,
		storageConfig:               storageConfig,
		db:                          db,
	}, nil
}

func (a *ArtifactImpl) ID() uuid.UUID {
	return a.id
}

func (a *ArtifactImpl) LogicalID() uuid.UUID {
	return a.logicalID
}

func (a *ArtifactImpl) Type() artifact.Type {
	return a.artifactType
}

func (a *ArtifactImpl) Name() string {
	return a.name
}

func (a *ArtifactImpl) Computed(ctx context.Context) bool {
	// An artifact is only considered computed if its results have been written.
	res := utils.ObjectExistsInStorage(
		ctx,
		a.storageConfig,
		a.execPaths.ArtifactMetadataPath,
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
		a.execPaths.ArtifactContentPath,
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
			a.execPaths.ArtifactMetadataPath,
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
	if a.previewArtifactCacheManager != nil {
		return errors.Newf("Artifact %s is cache-aware, so it cannot be persisted.", a.Name())
	}

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
	// There is nothing to do if the artifact was never even computed.
	if !a.Computed(ctx) {
		return
	}

	// Do not update the cache or clean anything up if the artifact result was persisted.
	if a.resultsPersisted {
		return
	}

	// Update the artifact cache, performing any necessary deletions.
	if a.previewArtifactCacheManager != nil {
		err := a.previewArtifactCacheManager.Put(context.TODO(), a.LogicalID(), a.execPaths)
		if err != nil {
			log.Errorf("Error when updating the result of artifact %s: %v", a.ID(), err)
		}
	}
}

func (a *ArtifactImpl) GetMetadata(ctx context.Context) (*artifact_result.Metadata, error) {
	if !a.Computed(ctx) {
		return nil, errors.Newf("Cannot get metadata of Artifact %s, it has not yet been computed.", a.Name())
	}

	var metadata artifact_result.Metadata
	err := utils.ReadFromStorage(ctx, a.storageConfig, a.execPaths.ArtifactMetadataPath, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (a *ArtifactImpl) GetContent(ctx context.Context) ([]byte, error) {
	if !a.Computed(ctx) {
		return nil, errors.Newf("Cannot get content of Artifact %s, it has not yet been computed.", a.Name())
	}
	content, err := storage.NewStorage(a.storageConfig).Get(ctx, a.execPaths.ArtifactContentPath)
	if err != nil {
		return nil, err
	}
	return content, nil
}
