package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/integration"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag_edge"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

type GetDistinctLoadOperatorsByWorkflowIdResponse struct {
	OperatorName    string              `db:"operator_name" json:"operator_name"`
	ModifiedAt      time.Time           `db:"modified_at" json:"modified_at"`
	IntegrationName string              `db:"integration_name" json:"integration_name"`
	IntegrationId   uuid.UUID           `db:"integration_id" json:"integration_id"`
	Service         integration.Service `db:"service" json:"service"`
	TableName       string              `db:"table_name" json:"object_name"`
	UpdateMode      string              `db:"update_mode" json:"update_mode"`
}

func (w *standardWriterImpl) CreateOperator(
	ctx context.Context,
	name string,
	description string,
	spec *Spec,
	executionEnvironmentID *uuid.UUID,
	db database.Database,
) (*DBOperator, error) {
	insertColumns := []string{NameColumn, DescriptionColumn, SpecColumn, ExecutionEnvironmentIDColumn}
	insertOperatorStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{name, description, spec, executionEnvironmentID}

	var operator DBOperator
	err := db.Query(ctx, &operator, insertOperatorStmt, args...)
	return &operator, err
}

func (r *standardReaderImpl) Exists(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (bool, error) {
	return utils.IdExistsInTable(ctx, id, tableName, db)
}

func (r *standardReaderImpl) GetOperator(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*DBOperator, error) {
	operators, err := r.GetOperators(ctx, []uuid.UUID{id}, db)
	if err != nil {
		return nil, err
	}

	if len(operators) != 1 {
		return nil, errors.Newf("Expected 1 operator, but got %d operators.", len(operators))
	}

	return &operators[0], nil
}

func (r *standardReaderImpl) GetOperators(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]DBOperator, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	getOperatorsQuery := fmt.Sprintf(
		"SELECT %s FROM operator WHERE id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var operators []DBOperator
	err := db.Query(ctx, &operators, getOperatorsQuery, args...)
	return operators, err
}

func (r *standardReaderImpl) GetOperatorsByWorkflowDagId(
	ctx context.Context,
	workflowDagId uuid.UUID,
	db database.Database,
) ([]DBOperator, error) {
	// Gets all operators that are a node with an incoming (id in `to_id`) or outgoing edge
	// (id in `from_id`) in the `workflow_dag_edge` for the specified DAG.
	getOperatorsByWorkflowDagIdQuery := fmt.Sprintf(
		`SELECT %s FROM operator WHERE id IN
		(SELECT from_id FROM workflow_dag_edge WHERE workflow_dag_id = $1 AND type = '%s' 
		UNION 
		SELECT to_id FROM workflow_dag_edge WHERE workflow_dag_id = $1 AND type = '%s')`,
		allColumns(),
		workflow_dag_edge.OperatorToArtifactType,
		workflow_dag_edge.ArtifactToOperatorType,
	)

	var operators []DBOperator
	err := db.Query(ctx, &operators, getOperatorsByWorkflowDagIdQuery, workflowDagId)
	return operators, err
}

func (w *standardWriterImpl) UpdateOperator(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*DBOperator, error) {
	var operator DBOperator
	err := utils.UpdateRecordToDest(ctx, &operator, changes, tableName, IdColumn, id, allColumns(), db)
	return &operator, err
}

func (w *standardWriterImpl) DeleteOperator(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return w.DeleteOperators(ctx, []uuid.UUID{id}, db)
}

func (w *standardWriterImpl) DeleteOperators(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	if len(ids) == 0 {
		return nil
	}

	deleteOperatorsStmt := fmt.Sprintf(
		"DELETE FROM operator WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)
	return db.Execute(ctx, deleteOperatorsStmt, args...)
}

func (r *standardReaderImpl) ValidateOperatorOwnership(
	ctx context.Context,
	organizationId string,
	operatorId uuid.UUID,
	db database.Database,
) (bool, error) {
	return utils.ValidateNodeOwnership(ctx, organizationId, operatorId, db)
}
