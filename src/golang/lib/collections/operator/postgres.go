package operator

import (
	"fmt"
	"context"

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

func (r *postgresReaderImpl) GetDistinctLoadOperatorsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]DBOperator, error) {
	query := fmt.Sprintf(`
	SELECT %s
	FROM (
		SELECT DISTINCT *, 
			json_extract_path_text(spec, 'load', 'integration_id'),
			json_extract_path_text(spec, 'load', 'parameters', 'table'), 
			json_extract_path_text(spec, 'load', 'parameters', 'update_mode') 
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
		)
	);`, allColumns())

	var workflowSpecs []DBOperator
	err := db.Query(ctx, &workflowSpecs, query, workflowId)
	return workflowSpecs, err
}
