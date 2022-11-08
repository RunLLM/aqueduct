package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type sqliteReaderImpl struct {
	standardReaderImpl
}

type sqliteWriterImpl struct {
	standardWriterImpl
}

func newSqliteReader() Reader {
	return &sqliteReaderImpl{standardReaderImpl{}}
}

func newSqliteWriter() Writer {
	return &sqliteWriterImpl{standardWriterImpl{}}
}

func (r *sqliteReaderImpl) GetOperatorsByIntegrationId(
	ctx context.Context,
	integrationId uuid.UUID,
	db database.Database,
) ([]DBOperator, error) {
	getOperatorsByIntegrationIdQuery := fmt.Sprintf(
		`SELECT %s FROM %s
		WHERE json_extract(spec, '$.load.integration_id') = $1
		OR json_extract(spec, '$.extract.integration_id') = $2`,
		allColumns(),
		tableName,
	)

	var operators []DBOperator
	err := db.Query(
		ctx,
		&operators,
		getOperatorsByIntegrationIdQuery,
		integrationId,
		integrationId,
	)
	return operators, err
}

func (w *sqliteWriterImpl) CreateOperator(
	ctx context.Context,
	name string,
	description string,
	spec *Spec,
	executionEnvironmentID *uuid.UUID,
	db database.Database,
) (*DBOperator, error) {
	insertColumns := []string{IdColumn, NameColumn, DescriptionColumn, SpecColumn, ExecutionEnvironmentIDColumn}
	insertOperatorStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{id, name, description, spec, executionEnvironmentID}

	var operator DBOperator
	err = db.Query(ctx, &operator, insertOperatorStmt, args...)
	return &operator, err
}

func (r *sqliteReaderImpl) GetLoadOperatorsForWorkflowAndIntegration(
	ctx context.Context,
	workflowId uuid.UUID,
	integrationId uuid.UUID,
	objectName string,
	db database.Database,
) ([]DBOperator, error) {
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
		EXISTS (
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
		);`, allColumns(), LoadType)

	var operators []DBOperator
	err := db.Query(ctx, &operators, query, objectName, integrationId, workflowId)

	return operators, err
}

func (r *sqliteReaderImpl) GetDistinctLoadOperatorsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]GetDistinctLoadOperatorsByWorkflowIdResponse, error) {
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

	var workflowSpecs []GetDistinctLoadOperatorsByWorkflowIdResponse
	err := db.Query(ctx, &workflowSpecs, query, workflowId)
	return workflowSpecs, err
}
