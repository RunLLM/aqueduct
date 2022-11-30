package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type artifactRepo struct {
	artifactReader
	artifactWriter
}

type artifactReader struct{}

type artifactWriter struct{}

func NewArtifactRepo() repos.Artifact {
	return &artifactRepo{
		artifactReader: artifactReader{},
		artifactWriter: artifactWriter{},
	}
}

func (*artifactReader) Exists(ctx context.Context, ID uuid.UUID, DB database.Database) (bool, error) {
	return utils.IdExistsInTable(ctx, ID, models.ArtifactTable, DB)
}

func (*artifactReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Artifact, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact WHERE id = $1;`,
		models.ArtifactCols(),
	)
	args := []interface{}{ID}

	return getArtifact(ctx, DB, query, args...)
}

func (*artifactReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Artifact, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM artifact WHERE id IN (%s);`,
		models.ArtifactCols(),
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getArtifacts(ctx, DB, query, args...)
}

func (*artifactReader) GetByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]models.Artifact, error) {
	// Gets all artifacts that are a node with an incoming (id in `to_id`) or outgoing edge
	// (id in `from_id`) in the `workflow_dag_edge` for the specified DAG.
	query := fmt.Sprintf(
		`SELECT %s FROM artifact WHERE id IN
		(SELECT from_id FROM workflow_dag_edge WHERE workflow_dag_id = $1 AND type = '%s' 
		UNION 
		SELECT to_id FROM workflow_dag_edge WHERE workflow_dag_id = $1 AND type = '%s')`,
		models.ArtifactCols(),
		workflow_dag_edge.ArtifactToOperatorType,
		workflow_dag_edge.OperatorToArtifactType,
	)
	args := []interface{}{dagID}

	return getArtifacts(ctx, DB, query, args...)
}

func (*artifactReader) ValidateOrg(ctx context.Context, ID uuid.UUID, orgID string, DB database.Database) (bool, error) {
	return utils.ValidateNodeOwnership(ctx, orgID, ID, DB)
}

func (*artifactWriter) Create(
	ctx context.Context,
	name string,
	description string,
	artifactType shared.ArtifactType,
	DB database.Database,
) (*models.Artifact, error) {
	cols := []string{
		models.ArtifactID,
		models.ArtifactName,
		models.ArtifactDescription,
		models.ArtifactType,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.ArtifactTable, cols, models.ArtifactCols())

	ID, err := utils.GenerateUniqueUUID(ctx, models.ArtifactTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{ID, name, description, artifactType}
	return getArtifact(ctx, DB, query, args...)
}

func (*artifactWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return deleteArtifacts(ctx, DB, []uuid.UUID{ID})
}

func (*artifactWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error {
	return deleteArtifacts(ctx, DB, IDs)
}

func (*artifactWriter) Update(
	ctx context.Context,
	ID uuid.UUID,
	changes map[string]interface{},
	DB database.Database,
) (*models.Artifact, error) {
	var artifact models.Artifact
	err := utils.UpdateRecordToDest(
		ctx,
		&artifact,
		changes,
		models.ArtifactTable,
		models.ArtifactID,
		ID,
		models.ArtifactCols(),
		DB,
	)
	return &artifact, err
}

func getArtifacts(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.Artifact, error) {
	var artifacts []models.Artifact
	err := DB.Query(ctx, &artifacts, query, args...)
	return artifacts, err
}

func getArtifact(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.Artifact, error) {
	artifacts, err := getArtifacts(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(artifacts) == 0 {
		return nil, database.ErrNoRows
	}

	if len(artifacts) != 1 {
		return nil, errors.Newf("Expected 1 artifact but got %v", len(artifacts))
	}

	return &artifacts[0], nil
}

func deleteArtifacts(ctx context.Context, DB database.Database, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		`DELETE FROM artifact WHERE id IN (%s)`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return DB.Execute(ctx, query, args...)
}
