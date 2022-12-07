package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/views"
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

	// GetByArtifactNameAndWorkflow returns the ArtifactResults for the given Workflow
	// where the associated Artifact is named artifactName.
	GetByArtifactNameAndWorkflow(ctx context.Context, artifactName string, workflowID uuid.UUID, DB database.Database) ([]models.ArtifactResult, error)

	// GetByArtifactAndDAGResult returns the ArtifactResult associated with the Artifact artifactID the DAGResult dagResultID.
	GetByArtifactAndDAGResult(ctx context.Context, artifactID uuid.UUID, dagResultID uuid.UUID, DB database.Database) (*models.ArtifactResult, error)

	// GetByDAGResults returns the ArtifactResult from a workflow DAG result with an ID in the dagResultIDs list.
	GetByDAGResults(ctx context.Context, dagResultIDs []uuid.UUID, DB database.Database) ([]models.ArtifactResult, error)

	// GetStatusByArtifactBatch returns an ArtifactResultStatus for each ArtifactResult associated
	// with an Artifact in artifactIDs.
	GetStatusByArtifactBatch(ctx context.Context, artifactIDs []uuid.UUID, DB database.Database) ([]views.ArtifactResultStatus, error)
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
		metadata *artifact_result.Metadata,
		DB database.Database,
	) (*models.ArtifactResult, error)

	// Delete deletes the ArtifactResult with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// DeleteBatch deletes the ArtifactResult with ID.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error

	// Update applies changes to the ArtifactResult with ID. It returns the updated ArtifactResult.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.ArtifactResult, error)
}
