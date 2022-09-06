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
	query := `
	SELECT * FROM (
		SELECT DISTINCT ON 
			(wf.id) wf.id AS id, wf.name AS name, 
		 	wf.description AS description, wf.created_at AS created_at, 
		 	wfdr.created_at AS last_run_at, wfdr.status as status 
		FROM 
			workflow AS wf 
			INNER JOIN app_user ON wf.user_id = app_user.id
			LEFT JOIN workflow_dag AS wfd ON wf.id = wfd.workflow_id
			LEFT JOIN workflow_dag_result AS wfdr ON wfd.id = wfdr.dag_id
		WHERE app_user.organization_id = $1 
		ORDER BY wf.id, wfdr.created_at DESC
	) AS temp
	ORDER BY created_at DESC;`

	var latestWorkflowResponse []LatestWorkflowResponse
	err := db.Query(ctx, &latestWorkflowResponse, query, organizationId)
	return latestWorkflowResponse, err
}
