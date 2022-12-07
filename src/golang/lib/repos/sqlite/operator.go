package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type operatorRepo struct {
	operatorReader
	operatorWriter
}

type operatorReader struct{}

type operatorWriter struct{}

func NewOperatorRepo() repos.Operator {
	return &operatorRepo{
		operatorReader: operatorReader{},
		operatorWriter: operatorWriter{},
	}
}

func (*operatorReader) Exists(ctx context.Context, ID uuid.UUID, DB database.Database) (bool, error) {
	return utils.IdExistsInTable(ctx, ID, models.OperatorTable, DB)
}

func (*operatorReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Operator, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM operator WHERE id = $1;`,
		models.OperatorCols(),
	)
	args := []interface{}{ID}

	return getOperator(ctx, DB, query, args...)
}

func (*operatorReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Operator, error) {
	if len(IDs) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	query := fmt.Sprintf(
		`SELECT %s FROM operator WHERE id IN (%s);`,
		models.OperatorCols(),
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) GetByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) ([]models.Operator, error) {
	// Gets all operators that are a node with an incoming (id in `to_id`) or outgoing edge
	// (id in `from_id`) in the `workflow_dag_edge` for the specified DAG.
	query := fmt.Sprintf(
		`SELECT %s FROM operator WHERE id IN
		(
			SELECT from_id 
			FROM workflow_dag_edge 
			WHERE workflow_dag_id = $1 AND type = '%s' 
			UNION 
			SELECT to_id 
			FROM workflow_dag_edge 
			WHERE workflow_dag_id = $1 AND type = '%s'
		)`,
		models.OperatorCols(),
		shared.OperatorToArtifactDAGEdge,
		shared.ArtifactToOperatorDAGEdge,
	)
	args := []interface{}{dagID}

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) GetDistinctLoadOPsByWorkflow(
	ctx context.Context,
	workflowID uuid.UUID,
	DB database.Database,
) ([]views.LoadOperator, error) {
	// Get all unique load operator (defined as a unique combination of integration,
	// table, and update mode) that has an edge (in `from_id` or `to_id`) in a DAG
	// belonging to the specified workflow in order of when the operator was last modified.
	query := `
	SELECT
		operator.name AS operator_name, 
		workflow_dag.created_at AS modified_at,
		integration.name AS integration_name, 
		json_extract(operator.spec, '$.load.integration_id') AS integration_id, 
		json_extract(operator.spec, '$.load.service') AS service, 
		json_extract(operator.spec, '$.load.parameters.table') AS table_name, 
		json_extract(operator.spec, '$.load.parameters.update_mode') AS update_mode
	FROM 
		operator, integration, workflow_dag_edge, workflow_dag
	WHERE (
		json_extract(spec, '$.type')='load' AND 
		integration.id = json_extract(operator.spec, '$.load.integration_id') AND
		( 
			workflow_dag_edge.from_id = operator.id OR 
			workflow_dag_edge.to_id = operator.id 
		) AND 
		workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
		workflow_dag.workflow_id = $1
	)
	GROUP BY
		integration.name, 
		json_extract(operator.spec, '$.load.integration_id'), 
		json_extract(operator.spec, '$.load.service'), 
		json_extract(operator.spec, '$.load.parameters.table'), 
		json_extract(operator.spec, '$.load.parameters.update_mode')
	ORDER BY modified_at DESC;`

	args := []interface{}{workflowID}

	var operators []views.LoadOperator
	err := DB.Query(ctx, &operators, query, args...)
	return operators, err
}

func (*operatorReader) GetLoadOPsByWorkflowAndIntegration(
	ctx context.Context,
	workflowID uuid.UUID,
	integrationID uuid.UUID,
	objectName string,
	DB database.Database,
) ([]models.Operator, error) {
	// Get all load operators where table=objectName & integration_id=integrationId
	// and has an edge (in `from_id` or `to_id`) in a DAG belonging to the specified
	// workflow.
	query := fmt.Sprintf(`
	SELECT %s
	FROM operator
	WHERE
		json_extract(spec, '$.type') = '%s' AND 
		json_extract(spec, '$.load.parameters.table')=$1 AND
		json_extract(spec, '$.load.integration_id')=$2 AND
		EXISTS 
		(
			SELECT 1 
			FROM 
				workflow_dag_edge, workflow_dag 
			WHERE 
			( 
				workflow_dag_edge.from_id = operator.id OR 
				workflow_dag_edge.to_id = operator.id 
			) AND 
			workflow_dag_edge.workflow_dag_id = workflow_dag.id AND 
			workflow_dag.workflow_id = $4
		);`,
		models.OperatorCols(),
		shared.LoadType,
	)

	return getOperators(ctx, DB, query)
}

func (*operatorReader) GetLoadOPsByIntegration(
	ctx context.Context,
	integrationID uuid.UUID,
	objectName string,
	DB database.Database,
) ([]models.Operator, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM operator
		WHERE 
			json_extract(spec, '$.load.integration_id') = $1
			OR json_extract(spec, '$.extract.integration_id') = $2`,
		models.OperatorCols(),
	)
	args := []interface{}{integrationID, integrationID}

	return getOperators(ctx, DB, query, args...)
}

func (*operatorReader) ValidateOrg(ctx context.Context, operatorId uuid.UUID, orgID string, DB database.Database) (bool, error) {
	return utils.ValidateNodeOwnership(ctx, orgID, operatorId, DB)
}

func (*operatorWriter) Create(
	ctx context.Context,
	name string,
	description string,
	spec *operator.Spec,
	executionEnvironmentID *uuid.UUID,
	DB database.Database,
) (*models.Operator, error) {
	cols := []string{
		models.OperatorID,
		models.OperatorName,
		models.OperatorDescription,
		models.OperatorSpec,
		models.OperatorExecutionEnvironmentID,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.OperatorTable, cols, models.OperatorCols())

	ID, err := utils.GenerateUniqueUUID(ctx, models.OperatorTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		name,
		description,
		spec,
		executionEnvironmentID,
	}

	return getOperator(ctx, DB, query, args...)
}

func (*operatorWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	return deleteOperators(ctx, DB, []uuid.UUID{ID})
}

func (*operatorWriter) DeleteBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) error {
	return deleteOperators(ctx, DB, IDs)
}

func (*operatorWriter) Update(
	ctx context.Context,
	ID uuid.UUID,
	changes map[string]interface{},
	DB database.Database,
) (*models.Operator, error) {
	var operator models.Operator
	err := utils.UpdateRecordToDest(
		ctx,
		&operator,
		changes,
		models.OperatorTable,
		models.OperatorID,
		ID,
		models.OperatorCols(),
		DB,
	)
	return &operator, err
}

func getOperators(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.Operator, error) {
	var operators []models.Operator
	err := DB.Query(ctx, &operators, query, args...)
	return operators, err
}

func getOperator(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.Operator, error) {
	operators, err := getOperators(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(operators) == 0 {
		return nil, database.ErrNoRows
	}

	if len(operators) != 1 {
		return nil, errors.Newf("Expected 1 Operator but got %v", len(operators))
	}

	return &operators[0], nil
}

func deleteOperators(ctx context.Context, DB database.Database, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		`DELETE FROM operator WHERE id IN (%s)`,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return DB.Execute(ctx, query, args...)
}
