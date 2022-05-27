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
) ([]latestWorkflowResponse, error) {
	query := `SELECT * FROM (SELECT DISTINCT ON (wf.id) wf.id AS id, wf.name AS name, 
		 wf.description AS description, wf.created_at AS created_at, 
		 wfdr.created_at AS last_run_at, wfdr.status as status 
		 FROM workflow AS wf, app_user, workflow_dag AS wfd, workflow_dag_result AS wfdr 
		 WHERE app_user.organization_id = $1 
		 AND wf.user_id = app_user.id AND wfd.workflow_id = wf.id AND wfdr.workflow_dag_id = wfd.id 
		 ORDER BY wf.id, wfdr.created_at DESC) AS temp
		 ORDER BY created_at DESC;`

	var latestWorkflowResponse []latestWorkflowResponse
	err := db.Query(ctx, &latestWorkflowResponse, query, organizationId)
	return latestWorkflowResponse, err
}
