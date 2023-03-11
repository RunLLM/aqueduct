package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/errors"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/shared/operator"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
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
	return IDExistsInTable(ctx, ID, models.ArtifactTable, DB)
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
			(
				SELECT from_id FROM workflow_dag_edge 
					WHERE workflow_dag_id = $1 AND type = '%s' 
			UNION 
				SELECT to_id FROM workflow_dag_edge 
					WHERE workflow_dag_id = $1 AND type = '%s'
			)`,
		models.ArtifactCols(),
		shared.ArtifactToOperatorDAGEdge,
		shared.OperatorToArtifactDAGEdge,
	)
	args := []interface{}{dagID}

	return getArtifacts(ctx, DB, query, args...)
}

func (*artifactReader) GetIDsByDAGAndDownstreamOPBatch(
	ctx context.Context,
	dagIDs []uuid.UUID,
	operatorIDs []uuid.UUID,
	DB database.Database,
) ([]uuid.UUID, error) {
	// Get all the unique `artifact_id`s with an outgoing edge to an operator specified by `operatorIds`
	// from workflow DAGs specified by `workflowDagIds`.
	query := fmt.Sprintf(
		`SELECT DISTINCT from_id AS id 
		FROM workflow_dag_edge
		WHERE 
			workflow_dag_id IN (%s) 
		 	AND to_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(dagIDs), 1),
		stmt_preparers.GenerateArgsList(len(operatorIDs), len(dagIDs)+1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(dagIDs)
	args = append(args, stmt_preparers.CastIdsListToInterfaceList(operatorIDs)...)

	var objectIDs []views.ObjectID
	err := DB.Query(ctx, &objectIDs, query, args...)
	if err != nil {
		return nil, err
	}

	IDs := make([]uuid.UUID, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		IDs = append(IDs, objectID.ID)
	}

	return IDs, nil
}

func (*artifactReader) ValidateOrg(ctx context.Context, ID uuid.UUID, orgID string, DB database.Database) (bool, error) {
	return validateNodeOwnership(ctx, orgID, ID, DB)
}

func (*artifactReader) GetMetricsByUpstreamArtifactBatch(
	ctx context.Context,
	artifactIDs []uuid.UUID,
	DB database.Database,
) (map[uuid.UUID][]models.Artifact, error) {
	query := fmt.Sprintf(
		`SELECT DISTINCT
			%s,
			edge_artf_to_metrics_op.from_id as upstream_id
		FROM
			workflow_dag_edge edge_artf_to_metrics_op,
			workflow_dag_edge edge_metrics_op_to_artf,
			operator,
			artifact 
		WHERE 
			artifact.id = edge_metrics_op_to_artf.to_id
			AND edge_artf_to_metrics_op.to_id = operator.id 
			AND edge_metrics_op_to_artf.from_id = operator.id
			AND json_extract(operator.spec, '$.type') = '%s'
			AND edge_artf_to_metrics_op.from_id IN (%s);`,
		models.ArtifactColsWithPrefix(),
		operator.MetricType,
		stmt_preparers.GenerateArgsList(len(artifactIDs), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(artifactIDs)

	type artifactWithUpstreamID struct {
		// copy of artifact
		ID          uuid.UUID           `db:"id"`
		Name        string              `db:"name"`
		Description string              `db:"description"`
		Type        shared.ArtifactType `db:"type"`
		UpstreamID  uuid.UUID           `db:"upstream_id"`
	}

	var queryRows []artifactWithUpstreamID
	err := DB.Query(ctx, &queryRows, query, args...)
	if err != nil {
		return nil, err
	}

	results := make(map[uuid.UUID][]models.Artifact, len(queryRows))
	for _, queryRow := range queryRows {
		results[queryRow.UpstreamID] = append(results[queryRow.UpstreamID], models.Artifact{
			ID:          queryRow.ID,
			Name:        queryRow.Name,
			Description: queryRow.Description,
			Type:        queryRow.Type,
		})
	}

	return results, nil
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

	ID, err := GenerateUniqueUUID(ctx, models.ArtifactTable, DB)
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
	err := repos.UpdateRecordToDest(
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
