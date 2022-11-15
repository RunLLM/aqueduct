package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// ArtifactResult defines all of the database operations that can be performed for an ArtifactResult.
type ArtifactResult interface {
	artifactResultReader
	artifactResultWriter
}

type artifactResultReader interface {
	// Get returns the ArtifactResult with ID.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.ArtifactResult, error)

	// GetBatch returns the ArtifactResults with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.ArtifactResult, error)

	// GetByArtifact returns the ArtifactResults with IDs artifactID.
	GetByArtifact(ctx context.Context, artifactID uuid.UUID, DB database.Database) ([]models.ArtifactResult, error)

	// GetByArtifactAndWorkflow returns the ArtifactResult with artifact name artifactName and workflow ID workflowID.
	GetByArtifactAndWorkflow(ctx context.Context, workflowID uuid.UUID, artifactName string, DB database.Database) ([]models.ArtifactResult, error)

	// GetByArtifactAndDAGResult returns the ArtifactResult with artifact ID artifactID and workflow DAG ID dagResultID.
	GetByArtifactAndDAGResult(ctx context.Context, dagResultID uuid.UUID, artifactID uuid.UUID, DB database.Database) ([]models.ArtifactResult, error)

	// GetByDAGResults returns the ArtifactResult from a workflow DAG result with an ID in the dagResultIDs list.
	GetByDAGResults(ctx context.Context, dagResultIDs []uuid.UUID, DB database.Database) ([]models.ArtifactResult, error)
}

type artifactResultWriter interface {
	// Create inserts a new ArtifactResult with the specified fields.
	Create(
		ctx context.Context,
		dagResultID uuid.UUID,
		artifactID uuid.UUID,
		contentPath string,
		DB database.Database,
	) (*models.ArtifactResult, error)

	// CreateWithExecStateAndMetadata inserts a new ArtifactResult with the specified fields.
	CreateWithExecStateAndMetadata(
		ctx context.Context,
		dagResultID uuid.UUID,
		artifactID uuid.UUID,
		contentPath string,
		execState *shared.ExecutionState,
		metadata *shared.ArtifactResultMetadata,
		DB database.Database,
	) (*models.ArtifactResult, error)

	// Delete deletes the ArtifactResult with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// DeleteBatch deletes the ArtifactResult with ID.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error

	// Update applies changes to the ArtifactResult with ID. It returns the updated ArtifactResult.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.ArtifactResult, error)
}
