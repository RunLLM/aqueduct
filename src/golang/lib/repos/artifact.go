package repos

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// Artifact defines all of the database operations that can be performed for an Artifact.
type Artifact interface {
	artifactReader
	artifactWriter
}

type artifactReader interface {
	// Exists returns whether an Artifact with ID exists.
	Exists(ctx context.Context, ID uuid.UUID, DB database.Database) (bool, error)

	// Get returns the Artifact with ID.
	Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Artifact, error)

	// GetBatch returns the Artifacts with IDs.
	GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Artifact, error)

	// GetByDAG returns the Artifacts created by the workflow DAG with ID dagID.
	GetByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]models.Artifact, error)

	// GetIDsByDAGAndDownstreamOPBatch returns a list of Artifact IDs belonging to a DAG
	// in dagIDs if it is connected via a DAGEdge to an operator in operatorIDs.
	GetIDsByDAGAndDownstreamOPBatch(
		ctx context.Context,
		dagIDs []uuid.UUID,
		operatorIDs []uuid.UUID,
		DB database.Database,
	) ([]uuid.UUID, error)

	// GetMetricsByUpstreamArtifactBatch returns a map of metrics Artifacts if they
	// are direct downstream of any artifact whose ID belongs to the given artifactIDs.
	// The returned map is keyed by the upstream artifact ID in the artifactIDs list.
	GetMetricsByUpstreamArtifactBatch(
		ctx context.Context,
		artifactIDs []uuid.UUID,
		DB database.Database,
	) (map[uuid.UUID][]models.Artifact, error)

	// ValidateOrg returns whether the Artifact was created by a user in orgID.
	ValidateOrg(ctx context.Context, ID uuid.UUID, orgID string, DB database.Database) (bool, error)
}

type artifactWriter interface {
	// Create inserts a new Artifact with the specified fields.
	Create(
		ctx context.Context,
		name string,
		description string,
		artifactType shared.ArtifactType,
		DB database.Database,
	) (*models.Artifact, error)

	// Update applies changes to the Artifact with ID. It returns the updated Artifact.
	Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Artifact, error)

	// Delete deletes the Artifact with ID.
	Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error

	// DeleteBatch deletes the Artifacts with IDs.
	DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error
}
