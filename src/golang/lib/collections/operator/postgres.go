package operator

import (
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
