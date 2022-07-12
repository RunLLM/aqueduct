package operator

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
) ([]Operator, error) {
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

	var workflowSpecs []Operator
	err := db.Query(ctx, &workflowSpecs, query, workflowId)
	return workflowSpecs, err
}
