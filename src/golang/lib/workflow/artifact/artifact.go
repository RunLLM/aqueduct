package artifact

import (
	"context"
	"encoding/json"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/storage"
	"github.com/aqueducthq/aqueduct/lib/workflow/preview_cache"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/xitongsys/parquet-go-source/buffer"
	"github.com/xitongsys/parquet-go/reader"
)

const sampleTableRow = 500

// Artifact is an interface for managing and inspect the lifecycle of an artifact
// produced by a workflow run.
type Artifact interface {
	ID() uuid.UUID
	Signature() uuid.UUID
	Type() shared.ArtifactType
	Name() string

	// InitializeResult initializes the artifact in the database.
	InitializeResult(ctx context.Context, dagResultID uuid.UUID) error

	// PersistResult updates the artifact result in the database.
	// Errors if InitializeResult() hasn't been called yet.
	PersistResult(ctx context.Context, execState *shared.ExecutionState) error

	// Finish is an end-of-lifecycle hook meant to do any final cleanup work.
	Finish(ctx context.Context)

	// Computed indicates whether this artifact's contents have been computed or not.
	// An artifact is only considered "computed" if its content has been written to storage.
	// This is *NOT* the same as having the operator's execution state == SUCCEEDED. For example,
	// for check operators, the artifact is computed even if the check operator does not pass
	// (returned false).
	Computed(ctx context.Context) bool

	// GetMetadata fetches the metadata for this artifact.
	// Errors if the artifact has not yet been computed.
	GetMetadata(ctx context.Context) (*shared.ArtifactResultMetadata, error)

	// GetContent fetches the content of this artifact.
	// Errors if the artifact has not yet been computed.
	GetContent(ctx context.Context) ([]byte, error)

	// SampleContent works similar to GetContent but takes only
	// a sample of data if it's too large to fit client.
	//
	// For now, it's primarily used for table artifact to limit
	// the number of rows sent to client.
	SampleContent(ctx context.Context) ([]byte, bool, error)
}

type ArtifactImpl struct {
	// This is the ID that will be stored in our database. It is the canonical handle
	// to this artifact throughout our system.
	id uuid.UUID

	// This is a more specific identifier than id, since it also encodes important structural/parameter
	// information about any upstream dependencies. It can be used as a unique handle to an artifact's
	// data, which is why it is used as the key in the preview artifact cache.
	signature uuid.UUID

	name         string
	description  string
	artifactType shared.ArtifactType

	execPaths *utils.ExecPaths

	repo           repos.Artifact
	resultRepo     repos.ArtifactResult
	resultID       uuid.UUID
	resultMetadata *shared.ArtifactResultMetadata

	// If this is not nil, this artifact should be written to the cache.
	// An artifact cannot be both cache-aware and persisted.
	previewCacheManager preview_cache.CacheManager
	resultsPersisted    bool

	storageConfig *shared.StorageConfig
	db            database.Database
}

func NewArtifact(
	signature uuid.UUID,
	dbArtifact models.Artifact,
	execPaths *utils.ExecPaths,
	artifactRepo repos.Artifact,
	artifactResultRepo repos.ArtifactResult,
	storageConfig *shared.StorageConfig,
	previewCacheManager preview_cache.CacheManager,
	db database.Database,
) (Artifact, error) {
	if previewCacheManager != nil && signature == uuid.Nil {
		return nil, errors.Newf("An Artifact signature must be provided for a cache-aware artifact.")
	}

	return &ArtifactImpl{
		id:                  dbArtifact.ID,
		signature:           signature,
		name:                dbArtifact.Name,
		description:         dbArtifact.Description,
		artifactType:        dbArtifact.Type,
		execPaths:           execPaths,
		repo:                artifactRepo,
		resultRepo:          artifactResultRepo,
		resultID:            uuid.Nil,
		resultMetadata:      nil,
		previewCacheManager: previewCacheManager,
		resultsPersisted:    false,
		storageConfig:       storageConfig,
		db:                  db,
	}, nil
}

func NewArtifactFromDBObjects(
	signature uuid.UUID,
	dbArtifact *models.Artifact,
	dbArtifactResult *models.ArtifactResult,
	artifactRepo repos.Artifact,
	artifactResultRepo repos.ArtifactResult,
	storageConfig *shared.StorageConfig,
	previewCacheManager preview_cache.CacheManager,
	DB database.Database,
) Artifact {
	var resultMetadata *shared.ArtifactResultMetadata
	if !dbArtifactResult.Metadata.IsNull {
		resultMetadata = &dbArtifactResult.Metadata.ArtifactResultMetadata
	}

	return &ArtifactImpl{
		id:           dbArtifact.ID,
		signature:    signature,
		name:         dbArtifact.Name,
		description:  dbArtifact.Description,
		artifactType: dbArtifact.Type,
		execPaths: &utils.ExecPaths{
			ArtifactContentPath: dbArtifactResult.ContentPath,
		},
		repo:                artifactRepo,
		resultRepo:          artifactResultRepo,
		resultID:            dbArtifactResult.ID,
		resultMetadata:      resultMetadata,
		previewCacheManager: previewCacheManager,
		resultsPersisted:    true,
		storageConfig:       storageConfig,
		db:                  DB,
	}
}

func (a *ArtifactImpl) ID() uuid.UUID {
	return a.id
}

func (a *ArtifactImpl) Signature() uuid.UUID {
	return a.signature
}

func (a *ArtifactImpl) Type() shared.ArtifactType {
	return a.artifactType
}

func (a *ArtifactImpl) Name() string {
	return a.name
}

func (a *ArtifactImpl) Computed(ctx context.Context) bool {
	// An artifact is only considered computed if its contents have been written.
	res := utils.ObjectExistsInStorage(
		ctx,
		a.storageConfig,
		a.execPaths.ArtifactContentPath,
	)
	return res
}

func (a *ArtifactImpl) InitializeResult(ctx context.Context, dagResultID uuid.UUID) error {
	if a.resultRepo == nil {
		return errors.New("Artifact's result writer cannot be nil.")
	}

	artifactResult, err := a.resultRepo.Create(
		ctx,
		dagResultID,
		a.ID(),
		a.execPaths.ArtifactContentPath,
		a.db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create artifact result record.")
	}

	a.resultID = artifactResult.ID
	return nil
}

func (a *ArtifactImpl) updateArtifactResultAfterComputation(
	ctx context.Context,
	execState *shared.ExecutionState,
) {
	changes := map[string]interface{}{
		models.ArtifactResultMetadata:  nil,
		models.ArtifactResultStatus:    execState.Status,
		models.ArtifactResultExecState: execState,
	}

	metadataExists := utils.ObjectExistsInStorage(ctx, a.storageConfig, a.execPaths.ArtifactMetadataPath)

	if a.Computed(ctx) && metadataExists {
		var artifactResultMetadata shared.ArtifactResultMetadata
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
		changes[models.ArtifactResultMetadata] = &artifactResultMetadata
	}

	_, err := a.resultRepo.Update(
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

// For lazily published workflows, we will need to update the artifact type in the database
// to something more specific than UNTYPED, so that we can enforce types in the future.
// Errors are ignored, since this update is meant to be best-effort.
func (a *ArtifactImpl) updateArtifactTypeAfterComputation(
	ctx context.Context,
) {
	if a.artifactType != shared.UntypedArtifact {
		return
	}

	if !a.Computed(ctx) {
		return
	}

	metadata, err := a.GetMetadata(ctx)
	if err != nil {
		log.Errorf("Error when fetching artifact metadata: %v", err)
		return
	}

	changes := map[string]interface{}{
		models.ArtifactType: metadata.ArtifactType,
	}

	_, err = a.repo.Update(ctx, a.ID(), changes, a.db)
	if err != nil {
		log.WithFields(
			log.Fields{
				"changes": changes,
			},
		).Errorf("Unable to update the artifact type: %v", err)
	}
}

func (a *ArtifactImpl) PersistResult(ctx context.Context, execState *shared.ExecutionState) error {
	if a.previewCacheManager != nil {
		return errors.Newf("Artifact %s is cache-aware, so it cannot be persisted.", a.Name())
	}

	if a.resultsPersisted {
		return errors.Newf("Artifact %s was already persisted!", a.name)
	}
	if !execState.Terminated() {
		return errors.Newf("Artifact %s has unexpected execution state: %s", a.Name(), execState.Status)
	}

	a.updateArtifactResultAfterComputation(ctx, execState)
	a.updateArtifactTypeAfterComputation(ctx)

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
	if a.previewCacheManager != nil {
		err := a.previewCacheManager.Put(context.TODO(), a.Signature(), a.execPaths)
		if err != nil {
			log.Errorf("Error when updating the result of artifact %s: %v", a.ID(), err)
		}
	}
}

func (a *ArtifactImpl) GetMetadata(ctx context.Context) (*shared.ArtifactResultMetadata, error) {
	if a.resultMetadata == nil {
		if !a.Computed(ctx) {
			// metadata is not ready yet.
			return nil, nil
		}

		// If the path is not available, we assume the data is not available.
		if !utils.ObjectExistsInStorage(ctx, a.storageConfig, a.execPaths.ArtifactMetadataPath) {
			return nil, nil
		}

		var metadata shared.ArtifactResultMetadata
		err := utils.ReadFromStorage(ctx, a.storageConfig, a.execPaths.ArtifactMetadataPath, &metadata)
		if err != nil {
			return nil, err
		}
		a.resultMetadata = &metadata
	}

	return a.resultMetadata, nil
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

func (a *ArtifactImpl) SampleContent(ctx context.Context) ([]byte, bool, error) {
	metadata, err := a.GetMetadata(ctx)
	if err != nil {
		return nil, false, err
	}

	// Ignore if artifact is not computed.
	// Ideally we should use a.Computed() but that involves a potential API call to storage.
	// So here we use metadata which is also sufficient.
	if metadata == nil {
		return nil, false, nil
	}

	// For 'compiled' types, we simply ignore the content as they won't be used by client.
	if metadata.SerializationType == shared.BytesSerialization || metadata.SerializationType == shared.PicklableSerialization {
		return nil, false, nil
	}

	content, err := a.GetContent(ctx)
	if err != nil {
		return nil, false, err
	}

	// For table types, we returns a down-sampled table when possible
	if metadata.SerializationType == shared.TableSerialization {
		type table struct {
			Schema map[string]interface{} `json:"schema"`
			Data   []interface{}          `json:"data"`
		}
		var t table
		err := json.Unmarshal(content, &t)

		// If a table artifact can be deserialized by json.Unmarshal(),
		// it is then an outdated Json artifact. We proceed with old method.
		if err == nil {
			if len(t.Data) <= sampleTableRow {
				// If the table is small, return the original content
				return content, false, nil
			}

			t.Data = t.Data[:sampleTableRow]
			downsampledContent, err := json.Marshal(t)
			if err != nil {
				return nil, false, err
			}

			return downsampledContent, true, nil
		}

		readBuffer := buffer.NewBufferFileFromBytes(content)
		parquetReader, err := reader.NewParquetReader(
			readBuffer,
			nil,
			// Number of parallel read.
			4,
		)
		if err != nil {
			return nil, false, err
		}
		parquetTable, err := parquetReader.ReadByNumber(sampleTableRow)
		if err != nil {
			return nil, false, err
		}
		schema := make(map[string]interface{})

		schemaFromParquet, _ := parquetReader.SchemaHandler.GetType(parquetReader.SchemaHandler.GetRootInName())

		for i := 0; i < schemaFromParquet.NumField(); i++ {
			schema[schemaFromParquet.Field(i).Name] = schemaFromParquet.Field(i).Type.String()[1:]
		}

		t.Data = parquetTable
		t.Schema = schema

		jsonTable, err := json.Marshal(t)
		if err != nil {
			return nil, false, err
		}
		if int(parquetReader.GetNumRows()) < sampleTableRow {
			return jsonTable, false, nil
		}
		return jsonTable, true, nil
	}

	if metadata.SerializationType == shared.BsonTableSerialization {
		// The over-simplified type for record orient.
		var t []interface{}

		err := json.Unmarshal(content, &t)
		if err != nil {
			return nil, false, err
		}

		if len(t) <= sampleTableRow {
			// If the table is small, return the original content
			return content, false, nil
		}

		t = t[:sampleTableRow]
		downsampledContent, err := json.Marshal(t)
		if err != nil {
			return nil, false, err
		}

		return downsampledContent, true, nil
	}

	return content, false, nil
}
