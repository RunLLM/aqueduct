package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type workflowRepo struct {
	workflowReader
	workflowWriter
}

type workflowReader struct{}

type workflowWriter struct{}

func NewWorklowRepo() repos.Workflow {
	return &workflowRepo{
		workflowReader: workflowReader{},
		workflowWriter: workflowWriter{},
	}
}

func (*workflowReader) Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error) {
	return utils.IdExistsInTable(ctx, id, models.WorkflowTable, db)
}

func (*workflowReader) Get(ctx context.Context, id uuid.UUID, db database.Database) (*models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow WHERE id = $1;`,
		models.WorkflowCols(),
	)
	args := []interface{}{id}

	return getOneWorkflow(ctx, db, query, args...)
}

func (*workflowReader) GetByOwnerAndName(ctx context.Context, ownerID uuid.UUID, name string, db database.Database) (*models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow WHERE user_id = $1 and name = $2;`,
		models.WorkflowCols(),
	)
	args := []interface{}{ownerID, name}

	return getOneWorkflow(ctx, db, query, args...)
}

func (*workflowReader) GetLatestStatusesByOrg(ctx context.Context, orgID uuid.UUID, db database.Database) ([]views.LatestWorkflowStatus, error) {
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
		WITH workflow_results AS
		(
			SELECT 
				wf.id AS id, wf.name AS name,
		 		wf.description AS description, wf.created_at AS created_at,
		 		wfdr.created_at AS run_at, wfdr.status as status, 
				json_extract(wfd.engine_config, '$.type') as engine
			FROM 
				workflow AS wf
				INNER JOIN app_user ON wf.user_id = app_user.id
				INNER JOIN workflow_dag AS wfd ON wf.id = wfd.workflow_id
				LEFT JOIN workflow_dag_result AS wfdr ON wfd.id = wfdr.workflow_dag_id
			WHERE 
				app_user.organization_id = $1
		),
		latest_result AS
		(
			SELECT 
				id, MAX(run_at) AS last_run_at
	  		FROM 
				workflow_results
	  		GROUP BY 
				id
		)
		SELECT 
			wfr.id, wfr.name, wfr.description, wfr.created_at, 
			wfr.run_at AS last_run_at, wfr.status, wfr.engine
		FROM 
			workflow_results AS wfr, latest_result AS lr
		WHERE 
			wfr.id = lr.id
			AND 
			(	
				wfr.run_at = lr.last_run_at
				OR 
				(
					wfr.run_at IS NULL 
					AND lr.last_run_at IS NULL
				)
			)
		ORDER BY 
			created_at DESC;`
	args := []interface{}{orgID}

	var latestWorkflowResponse []views.LatestWorkflowStatus
	err := db.Query(ctx, &latestWorkflowResponse, query, args...)
	return latestWorkflowResponse, err
}

func (*workflowReader) List(ctx context.Context, db database.Database) ([]models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow;`,
		models.WorkflowCols(),
	)

	return getWorkflow(ctx, db, query)
}

func (*workflowReader) ValidateOrg(ctx context.Context, id uuid.UUID, orgID uuid.UUID, db database.Database) (bool, error) {
	query := `
	SELECT 
		COUNT(*) AS count 
	FROM 
		workflow INNER JOIN app_user ON workflow.user_id = app_user.id
	WHERE
		workflow.id = $1
		AND app_user.organization_id = $2;`
	args := []interface{}{id, orgID}

	var count utils.CountResult
	err := db.Query(ctx, &count, query, args...)
	if err != nil {
		return false, err
	}

	return count.Count == 1, nil
}

func (*workflowWriter) Create(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	description string,
	schedule *shared.Schedule,
	retentionPolicy *shared.RetentionPolicy,
	db database.Database,
) (*models.Workflow, error) {
	cols := []string{
		models.WorkflowID,
		models.WorkflowUserID,
		models.WorkflowName,
		models.WorkflowDescription,
		models.WorkflowSchedule,
		models.WorkflowCreatedAt,
		models.WorkflowRetention,
	}
	query := db.PrepareInsertWithReturnAllStmt(models.WorkflowTable, cols, models.WorkflowCols())

	id, err := utils.GenerateUniqueUUID(ctx, models.WorkflowTable, db)
	if err != nil {
		return nil, err
	}

	args := []interface{}{id, userID, name, description, schedule, time.Now(), retentionPolicy}
	return getOneWorkflow(ctx, db, query, args...)
}

func (*workflowWriter) Delete(ctx context.Context, id uuid.UUID, db database.Database) error {
	query := `DELETE FROM workflow WHERE id = $1;`
	args := []interface{}{id}
	return db.Execute(ctx, query, args...)
}

func (*workflowWriter) Update(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*models.Workflow, error) {
	var workflow models.Workflow
	err := utils.UpdateRecordToDest(
		ctx,
		&workflow,
		changes,
		models.WorkflowTable,
		models.WorkflowID,
		id,
		models.WorkflowCols(),
		db,
	)
	return &workflow, err
}

func getWorkflow(ctx context.Context, db database.Database, query string, args ...interface{}) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := db.Query(ctx, &workflows, query, args...)
	return workflows, err
}

func getOneWorkflow(ctx context.Context, db database.Database, query string, args ...interface{}) (*models.Workflow, error) {
	workflows, err := getWorkflow(ctx, db, query, args...)
	if err != nil {
		return nil, nil
	}

	if len(workflows) != 1 {
		return nil, errors.Newf("Expected 1 workflow but got %v", len(workflows))
	}

	return &workflows[0], nil
}
