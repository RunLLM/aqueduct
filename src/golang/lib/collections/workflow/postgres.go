package workflow

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
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

func (r *postgresReaderImpl) GetWorkflowsWithLatestRunResult(
	ctx context.Context,
	organizationId string,
	db database.Database,
) ([]LatestWorkflowResponse, error) {
	// Get workflow metadata (id, name, description, creation time, last run time, and last run status)
	// for all workflows whose `organization_id` is `organizationId` ordered by when the workflow was created.
	// Get the last run DAG by getting the max created_at timestamp for all workflow DAGs associated with each
	// workflow in the organization.

	// We want to return 1 row for each workflow, so we use a LEFT JOIN between the workflow_dag
	// and workflow_dag_result tables. A LEFT JOIN outputs all rows in the left table even if there
	// is no match with a row in the right table. If there is no match, the columns of the right table
	// are NULL.
	// This means that `last_run_at` and `status` in the query output can be NULL.
	query := `
	SELECT * FROM (
		SELECT DISTINCT ON 
			(wf.id) wf.id AS id, wf.name AS name, 
		 	wf.description AS description, wf.created_at AS created_at, 
		 	wfdr.created_at AS last_run_at, wfdr.status as status, 
			 son_extract_path_text(wfd.engine_config, 'type') as engine
		FROM 
			workflow AS wf 
			INNER JOIN app_user ON wf.user_id = app_user.id
			INNER JOIN workflow_dag AS wfd ON wf.id = wfd.workflow_id
			LEFT JOIN workflow_dag_result AS wfdr ON wfd.id = wfdr.workflow_dag_id
		WHERE app_user.organization_id = $1 
		ORDER BY wf.id, wfdr.created_at DESC
	) AS temp
	ORDER BY created_at DESC;`

	var latestWorkflowResponse []LatestWorkflowResponse
	err := db.Query(ctx, &latestWorkflowResponse, query, organizationId)
	return latestWorkflowResponse, err
}
