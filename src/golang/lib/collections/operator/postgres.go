package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type postgresReaderImpl struct {
	standardReaderImpl
}

type postgresWriterImpl struct {
	standardWriterImpl
}

func newPostgresReader() Reader {
	return &postgresReaderImpl{standardReaderImpl{}}
}

func newPostgresWriter() Writer {
	return &postgresWriterImpl{standardWriterImpl{}}
}

func (r *postgresReaderImpl) TableTouchedByWorkflow(
	ctx context.Context,
	workflowId string,
	integrationId string,
	tableName string,
	db database.Database,
) (bool, error) {
	query := fmt.Sprintf(`
	SELECT %s
	FROM (
		SELECT *
		FROM operator
		WHERE
			json_extract_path_text(spec, 'type') = 'load' AND 
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
					workflow_dag.workflow_id = $1
			)
	)
	WHERE
		json_extract_path_text(spec, 'load', 'integration_id')=$2 AND
		json_extract_path_text(spec, 'load', 'parameters', 'table')=$3;`, allColumns())

	var operators []DBOperator
	err := db.Query(ctx, &operators, query, workflowId, integrationId, tableName)

	touched := false
	if len(operators) > 0 {
		touched = true
	}

	return touched, err
}

func (r *postgresReaderImpl) GetOperatorsByIntegrationId(
	ctx context.Context,
	integrationId uuid.UUID,
	db database.Database,
) ([]DBOperator, error) {
	getOperatorsByIntegrationIdQuery := fmt.Sprintf(
		`SELECT %s FROM %s
		WHERE json_extract_text(spec, 'load', 'integration_id') = $1
		OR json_extract_text(spec, 'extract', 'integration_id') = $2`,
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

func (r *postgresReaderImpl) GetDistinctLoadOperatorsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]GetDistinctLoadOperatorsByWorkflowIdResponse, error) {
	query := `
	SELECT DISTINCT 
		name, 
		json_extract_path_text(spec, 'load', 'integration_id') AS integration_id,
		json_extract_path_text(spec, 'load', 'service') AS service, 
		json_extract_path_text(spec, 'load', 'parameters', 'table') AS table_name, 
		json_extract_path_text(spec, 'load', 'parameters', 'update_mode')  AS update_mode
	FROM operator WHERE (
		json_extract_path_text(spec, 'type') = 'load' AND 
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
				workflow_dag.workflow_id = $1
		)
	);`

	var workflowSpecs []GetDistinctLoadOperatorsByWorkflowIdResponse
	err := db.Query(ctx, &workflowSpecs, query, workflowId)
	return workflowSpecs, err
}
