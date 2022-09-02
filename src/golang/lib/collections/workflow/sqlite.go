package workflow

import (
	"context"
	"time"

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

func (w *sqliteWriterImpl) CreateWorkflow(
	ctx context.Context,
	userId uuid.UUID,
	name string,
	description string,
	schedule *Schedule,
	retentionPolicy *RetentionPolicy,
	db database.Database,
) (*Workflow, error) {
	insertColumns := []string{IdColumn, UserIdColumn, NameColumn, DescriptionColumn, ScheduleColumn, CreatedAtColumn, RetentionColumn}
	insertWorkflowStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	id, err := utils.GenerateUniqueUUID(ctx, tableName, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{id, userId, name, description, schedule, time.Now(), retentionPolicy}

	var workflow Workflow
	err = db.Query(ctx, &workflow, insertWorkflowStmt, args...)
	return &workflow, err
}

func (r *sqliteReaderImpl) GetWorkflowsWithLatestRunResult(
	ctx context.Context,
	organizationId string,
	db database.Database,
) ([]latestWorkflowResponse, error) {
	// Get workflow metadata (id, name, description, creation time, last run time, and last run status)
	// for all workflows whose `organization_id` is `organizationId` ordered by when the workflow was created.
	// Get the last run DAG by getting the max created_at timestamp for all workflow DAGs associated with each
	// workflow in the organization.
	query := `
		WITH workflow_results AS
		(
			SELECT
				wf.id AS id, wf.name AS name, 
		 		wf.description AS description, wf.created_at AS created_at, 
		 		wfdr.created_at AS last_run_at, wfdr.status as status
			FROM
				workflow AS wf
				INNER JOIN app_user ON wf.user_id = app_user.id
				INNER JOIN workflow_dag AS wfd ON wf.id = wfd.workflow_id
				LEFT JOIN workflow_dag_result AS wfdr ON wfd.id = wfdr.workflow_dag_id
			WHERE app_user.organization_id = $1
		),
		latest_result AS
		(
			SELECT 
				id, MAX(created_at) AS last_run_at 
	  		FROM 
				workflow_results
	  		GROUP BY wf.id
		),
		SELECT
			wf.id AS id, wf.name AS name, 
		 	wf.description AS description, wf.created_at AS created_at, 
		 	wfdr.created_at AS last_run_at, wfdr.status as status 
		 	FROM workflow AS wf, app_user, workflow_dag AS wfd, workflow_dag_result AS wfdr, 
		 		(
					SELECT 
						wf.id AS id, MAX(wfdr.created_at) AS last_run_at 
		  			FROM workflow AS wf, app_user, workflow_dag AS wfd, workflow_dag_result AS wfdr 
		  			WHERE app_user.organization_id = $1 
		  			AND wf.user_id = app_user.id AND wfd.workflow_id = wf.id AND wfdr.workflow_dag_id = wfd.id 
		  			GROUP BY wf.id
				) AS wflr 
		WHERE 
			app_user.organization_id = $1 
		 	AND wf.user_id = app_user.id 
			AND wfd.workflow_id = wf.id 
			AND wfdr.workflow_dag_id = wfd.id 
			AND wf.id = wflr.id AND wfdr.created_at = wflr.last_run_at 
		ORDER BY created_at DESC;`

	var latestWorkflowResponse []latestWorkflowResponse
	err := db.Query(ctx, &latestWorkflowResponse, query, organizationId)
	return latestWorkflowResponse, err
}
