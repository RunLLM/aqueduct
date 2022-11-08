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
	Get(ctx context.Context, ID uuid.UUID, db database.Database) (*models.ArtifactResult, error)

	// GetBatch returns the ArtifactResults with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, db database.Database) ([]models.ArtifactResult, error)

	// GetByArtifact returns the ArtifactResult with ID artifactID.
	GetByArtifact(ctx context.Context, artifactID uuid.UUID, db database.Database) ([]models.ArtifactResult, error)

	// GetByArtifactAndWorkflow returns the ArtifactResult with artifact name artifactName and workflow ID workflowID.
	GetByArtifactAndWorkflow(ctx context.Context, workflowID uuid.UUID, artifactName string, db database.Database) ([]models.ArtifactResult, error)

	// GetByDAGAndArtifact returns the ArtifactResult with artifact ID artifactID and workflow DAG ID workflowDAGResultID.
	GetByDAGAndArtifact(ctx context.Context, workflowDAGResultID uuid.UUID, artifactID uuid.UUID, db database.Database) ([]models.ArtifactResult, error)

	// GetByDAGs returns the ArtifactResult from a workflow DAG with an ID in the workflowDAGResultIDs list.
	GetByDAGs(ctx context.Context, workflowDAGResultIDs []uuid.UUID, db database.Database) ([]models.ArtifactResult, error)
}

type artifactResultWriter interface {
	// Create inserts a new ArtifactResult with the specified fields.
	Create(
		ctx context.Context,
		workflowDAGResultID uuid.UUID,
		artifactID uuid.UUID,
		contentPath string,
		db database.Database,
	) (*models.ArtifactResult, error)

	// CreateWithExecStateAndMetadata inserts a new ArtifactResult with the specified fields.
	CreateWithExecStateAndMetadata(
		ctx context.Context,
		workflowDAGResultID uuid.UUID,
		artifactID uuid.UUID,
		contentPath string,
		execState *shared.ExecutionState,
		metadata *shared.Metadata,
		db database.Database,
	) (*models.ArtifactResult, error)

	// Delete deletes the ArtifactResult with ID.
	Delete(ctx context.Context, ID uuid.UUID, db database.Database) error

	// DeleteBatch deletes the ArtifactResult with ID.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, db database.Database) error

	// Update applies changes to the ArtifactResult with ID. It returns the updated ArtifactResult.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, db database.Database) (*models.ArtifactResult, error)
}
