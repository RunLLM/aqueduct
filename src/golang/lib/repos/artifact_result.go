package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/google/uuid"
)

// ArtifactResult defines all of the database operations that can be performed for an ArtifactResult.
type ArtifactResult interface {
	artifactResultReader
	artifactResultWriter
}

type artifactResultReader interface {
	// Get returns the ArtifactResult with id.
	Get(ctx context.Context, id uuid.UUID, db database.Database) (*models.ArtifactResult, error)

	// GetMultiple returns the ArtifactResults with ids.
	GetMultiple(ctx context.Context, ids []uuid.UUID, db database.Database) ([]models.ArtifactResult, error)

	// GetByArtifactId returns the ArtifactResult with id artifactId.
	GetByArtifactId(ctx context.Context, artifactId uuid.UUID, db database.Database) ([]models.ArtifactResult, error)

	// GetByArtifactId returns the ArtifactResult with artifact name name and workflow id workflowId .
	GetByArtifactNameAndWorkflowId(ctx context.Context, workflowId uuid.UUID, name string, db database.Database) ([]models.ArtifactResult, error)

	// GetByWorkflowDagResultIdAndArtifactId returns the ArtifactResult with artifact name name and workflow DAG id workflowDagResultId.
	GetByWorkflowDagResultIdAndArtifactId(ctx context.Context, workflowDagResultId uuid.UUID, artifactId uuid.UUID, db database.Database) ([]models.ArtifactResult, error)

	// GetByWorkflowDagResultIds returns the ArtifactResult from a workflow DAG with an id in the workflowDagResultIds list.
	GetByWorkflowDagResultIds(ctx context.Context, workflowDagResultIds []uuid.UUID, db database.Database) ([]models.ArtifactResult, error)
}

type artifactResultWriter interface {
	// Create inserts a new ArtifactResult with the specified fields.
	Create(
		ctx context.Context,
		workflowDagResultId uuid.UUID,
		artifactId uuid.UUID,
		contentPath string,
		db database.Database,
	) (*models.ArtifactResult, error)

	// CreateWithExecStateAndMetadata inserts a new ArtifactResult with the specified fields.
	CreateWithExecStateAndMetadata(
		ctx context.Context,
		workflowDagResultId uuid.UUID,
		artifactId uuid.UUID,
		contentPath string,
		execState *shared.ExecutionState,
		metadata *Metadata,
		db database.Database,
	) (*models.ArtifactResult, error)

	// Delete deletes the ArtifactResult with id.
	Delete(ctx context.Context, id uuid.UUID, db database.Database) error

	// DeleteMultiple deletes the ArtifactResult with id.
	DeleteMultiple(ctx context.Context, ids []uuid.UUID, db database.Database) error

	// Update applies changes to the ArtifactResult with id. It returns the updated ArtifactResult.
	Update(ctx context.Context, id uuid.UUID, changes map[string]interface{}, db database.Database) (*models.ArtifactResult, error)
}
