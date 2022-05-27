package artifact

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateArtifact(
	ctx context.Context,
	name string,
	description string,
	spec *Spec,
	db database.Database,
) (*Artifact, error) {
	insertColumns := []string{NameColumn, DescriptionColumn, SpecColumn}
	insertArtifactStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{name, description, spec}

	var artifact Artifact
	err := db.Query(ctx, &artifact, insertArtifactStmt, args...)
	return &artifact, err
}

func (r *standardReaderImpl) Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error) {
	return utils.IdExistsInTable(ctx, id, tableName, db)
}

func (r *standardReaderImpl) GetArtifact(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*Artifact, error) {
	artifacts, err := r.GetArtifacts(ctx, []uuid.UUID{id}, db)
	if err != nil {
		return nil, err
	}

	if len(artifacts) != 1 {
		return nil, errors.Newf("Expected 1 artifact, but got %d artifacts.", len(artifacts))
	}

	return &artifacts[0], nil
}

func (r *standardReaderImpl) GetArtifacts(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]Artifact, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	getArtifactsQuery := fmt.Sprintf(
		"SELECT %s FROM artifact WHERE id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var artifacts []Artifact
	err := db.Query(ctx, &artifacts, getArtifactsQuery, args...)
	return artifacts, err
}

func (r *standardReaderImpl) GetArtifactsByWorkflowDagId(
	ctx context.Context,
	workflowDagId uuid.UUID,
	db database.Database,
) ([]Artifact, error) {
	getArtifactsByWorkflowDagIdQuery := fmt.Sprintf(
		`SELECT %s FROM artifact WHERE id IN
		(SELECT from_id FROM workflow_dag_edge WHERE workflow_dag_id = $1 AND type = '%s' 
		UNION 
		SELECT to_id FROM workflow_dag_edge WHERE workflow_dag_id = $1 AND type = '%s')`,
		allColumns(),
		workflow_dag_edge.ArtifactToOperatorType,
		workflow_dag_edge.OperatorToArtifactType,
	)

	var artifacts []Artifact
	err := db.Query(ctx, &artifacts, getArtifactsByWorkflowDagIdQuery, workflowDagId)
	return artifacts, err
}

func (w *standardWriterImpl) UpdateArtifact(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*Artifact, error) {
	var artifact Artifact
	err := utils.UpdateRecordToDest(ctx, &artifact, changes, tableName, IdColumn, id, allColumns(), db)
	return &artifact, err
}

func (w *standardWriterImpl) DeleteArtifact(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return w.DeleteArtifacts(ctx, []uuid.UUID{id}, db)
}

func (w *standardWriterImpl) DeleteArtifacts(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	if len(ids) == 0 {
		return nil
	}

	deleteArtifactsStmt := fmt.Sprintf(
		"DELETE FROM artifact WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)
	return db.Execute(ctx, deleteArtifactsStmt, args...)
}

func (r *standardReaderImpl) ValidateArtifactOwnership(
	ctx context.Context,
	organizationId string,
	artifactId uuid.UUID,
	db database.Database,
) (bool, error) {
	return utils.ValidateNodeOwnership(ctx, organizationId, artifactId, db)
}
