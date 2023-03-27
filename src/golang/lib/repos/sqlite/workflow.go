package sqlite

import (
	"context"
	"fmt"
	"time"

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

func (*workflowReader) Exists(ctx context.Context, ID uuid.UUID, DB database.Database) (bool, error) {
	return IDExistsInTable(ctx, ID, models.WorkflowTable, DB)
}

func (*workflowReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow WHERE id = $1;`,
		models.WorkflowCols(),
	)
	args := []interface{}{ID}

	return getWorkflow(ctx, DB, query, args...)
}

func (*workflowReader) GetByDAG(ctx context.Context, dagID uuid.UUID, DB database.Database) (*models.Workflow, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM workflow, workflow_dag 
		WHERE workflow.id = workflow_dag.workflow_id 
		AND workflow_dag.id = $1;`,
		models.WorkflowColsWithPrefix(),
	)
	args := []interface{}{dagID}

	return getWorkflow(ctx, DB, query, args...)
}

func (*workflowReader) GetByOwnerAndName(ctx context.Context, ownerID uuid.UUID, name string, DB database.Database) (*models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow WHERE user_id = $1 and name = $2;`,
		models.WorkflowCols(),
	)
	args := []interface{}{ownerID, name}

	return getWorkflow(ctx, DB, query, args...)
}

func (*workflowReader) GetByScheduleTrigger(
	ctx context.Context,
	trigger shared.UpdateTrigger,
	DB database.Database,
) ([]models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow WHERE
			json_extract(schedule, '$.trigger') = $1;
		`,
		models.WorkflowCols(),
	)
	args := []interface{}{trigger}

	return getWorkflows(ctx, DB, query, args...)
}

func (*workflowReader) GetTargets(ctx context.Context, ID uuid.UUID, DB database.Database) ([]uuid.UUID, error) {
	query := `
		SELECT id FROM workflow
		WHERE
			json_extract(schedule, '$.trigger') = $1
			AND json_extract(schedule, '$.source_id') = $2
		;`
	args := []interface{}{shared.CascadingUpdateTrigger, ID}

	var objectIDs []views.ObjectID
	err := DB.Query(ctx, &objectIDs, query, args...)
	if err != nil {
		return nil, err
	}

	IDs := make([]uuid.UUID, 0, len(objectIDs))
	for _, objectID := range objectIDs {
		IDs = append(IDs, objectID.ID)
	}

	return IDs, nil
}

func (*workflowReader) GetLastRunByEngine(
	ctx context.Context,
	engine shared.EngineType,
	DB database.Database,
) ([]views.WorkflowLastRun, error) {
	query := `
		SELECT 
			workflow.id AS workflow_id, 
			workflow.schedule, 
			workflow_dag_result.created_at AS last_run_at 
		FROM 
			workflow, 
			workflow_dag, 
			workflow_dag_result, 
			(
				SELECT 
					workflow.id, 
					MAX(workflow_dag_result.created_at) AS created_at 
				FROM 
					workflow, 
					workflow_dag, 
					workflow_dag_result 
				WHERE 
					workflow.id = workflow_dag.workflow_id 
					AND workflow_dag.id = workflow_dag_result.workflow_dag_id 
				GROUP BY workflow.id
			) AS workflow_latest_run 
		WHERE 
			workflow.id = workflow_dag.workflow_id 
			AND workflow_dag.id = workflow_dag_result.workflow_dag_id 
			AND workflow.id = workflow_latest_run.id 
			AND workflow_dag_result.created_at = workflow_latest_run.created_at
			AND json_extract(workflow_dag.engine_config, '$.type') = $1;`

	var lastRuns []views.WorkflowLastRun
	args := []interface{}{engine}

	err := DB.Query(ctx, &lastRuns, query, args...)
	return lastRuns, err
}

func (*workflowReader) GetLatestStatusesByOrg(ctx context.Context, orgID string, DB database.Database) ([]views.LatestWorkflowStatus, error) {
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
				wfdr.id as result_id, wfd.id AS dag_id,
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
			wfr.id,
			wfr.name,
			wfr.description,
			wfr.created_at,
			wfr.result_id,
			wfr.dag_id, 
			wfr.run_at AS last_run_at,
			wfr.status,
			wfr.engine
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
	err := DB.Query(ctx, &latestWorkflowResponse, query, args...)
	return latestWorkflowResponse, err
}

func (*workflowReader) List(ctx context.Context, DB database.Database) ([]models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow;`,
		models.WorkflowCols(),
	)

	return getWorkflows(ctx, DB, query)
}

func (*workflowReader) ValidateOrg(ctx context.Context, ID uuid.UUID, orgID string, DB database.Database) (bool, error) {
	query := `
	SELECT 
		COUNT(*) AS count 
	FROM 
		workflow INNER JOIN app_user ON workflow.user_id = app_user.id
	WHERE
		workflow.id = $1
		AND app_user.organization_id = $2;`
	args := []interface{}{ID, orgID}

	var count countResult
	err := DB.Query(ctx, &count, query, args...)
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
	notificationSettings *shared.NotificationSettings,
	DB database.Database,
) (*models.Workflow, error) {
	cols := []string{
		models.WorkflowID,
		models.WorkflowUserID,
		models.WorkflowName,
		models.WorkflowDescription,
		models.WorkflowSchedule,
		models.WorkflowCreatedAt,
		models.WorkflowRetentionPolicy,
		models.WorkflowNotificationSettings,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.WorkflowTable, cols, models.WorkflowCols())

	ID, err := GenerateUniqueUUID(ctx, models.WorkflowTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{ID, userID, name, description, schedule, time.Now(), retentionPolicy, notificationSettings}
	return getWorkflow(ctx, DB, query, args...)
}

func (*workflowWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	query := `DELETE FROM workflow WHERE id = $1;`
	args := []interface{}{ID}
	return DB.Execute(ctx, query, args...)
}

func (*workflowWriter) Update(
	ctx context.Context,
	ID uuid.UUID,
	changes map[string]interface{},
	DB database.Database,
) (*models.Workflow, error) {
	var workflow models.Workflow
	err := repos.UpdateRecordToDest(
		ctx,
		&workflow,
		changes,
		models.WorkflowTable,
		models.WorkflowID,
		ID,
		models.WorkflowCols(),
		DB,
	)
	return &workflow, err
}

func (*workflowWriter) RemoveNotificationFromSettings(ctx context.Context, notificationIntegrationID uuid.UUID, DB database.Database) error {
	query := `
	UPDATE workflow
	SET
		notification_settings=CAST(
			json_remove(
				notification_settings,
				$1
			) AS BLOB
		)
	WHERE
		notification_settings IS NOT NULL
		AND json_extract(
			notification_settings,
			$1
		) IS NOT NULL;`
	json_path := fmt.Sprintf("$.settings.%s", notificationIntegrationID)
	return DB.Execute(ctx, query, json_path)
}

func getWorkflows(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := DB.Query(ctx, &workflows, query, args...)
	return workflows, err
}

func getWorkflow(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.Workflow, error) {
	workflows, err := getWorkflows(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(workflows) == 0 {
		return nil, database.ErrNoRows()
	}

	if len(workflows) != 1 {
		return nil, errors.Newf("Expected 1 workflow but got %v", len(workflows))
	}

	return &workflows[0], nil
}
