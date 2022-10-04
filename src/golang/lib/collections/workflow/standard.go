package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateWorkflow(
	ctx context.Context,
	userId uuid.UUID,
	name string,
	description string,
	schedule *Schedule,
	retentionPolicy *RetentionPolicy,
	db database.Database,
) (*Workflow, error) {
	insertColumns := []string{UserIdColumn, NameColumn, DescriptionColumn, ScheduleColumn, CreatedAtColumn, RetentionColumn}
	insertWorkflowStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{userId, name, description, schedule, time.Now(), retentionPolicy}

	var workflow Workflow
	err := db.Query(ctx, &workflow, insertWorkflowStmt, args...)
	return &workflow, err
}

func (r *standardReaderImpl) Exists(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (bool, error) {
	return utils.IdExistsInTable(ctx, id, tableName, db)
}

func (r *standardReaderImpl) GetWorkflow(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*Workflow, error) {
	workflows, err := r.GetWorkflows(ctx, []uuid.UUID{id}, db)
	if err != nil {
		return nil, err
	}

	if len(workflows) != 1 {
		return nil, errors.Newf("Expected 1 workflow, but got %d workflows.", len(workflows))
	}

	return &workflows[0], nil
}

func (r *standardReaderImpl) GetWorkflowByWorkflowDagId(
	ctx context.Context,
	workflowDagId uuid.UUID,
	db database.Database,
) (*Workflow, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM workflow, workflow_dag 
		WHERE workflow.id = workflow_dag.workflow_id 
		AND workflow_dag.id = $1;`,
		allColumnsWithPrefix(),
	)

	var workflow Workflow
	err := db.Query(ctx, &workflow, query, workflowDagId)
	return &workflow, err
}

func (r *standardReaderImpl) GetWorkflows(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]Workflow, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	getWorkflowsQuery := fmt.Sprintf(
		"SELECT %s FROM workflow WHERE id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var workflows []Workflow
	err := db.Query(ctx, &workflows, getWorkflowsQuery, args...)
	return workflows, err
}

func (r *standardReaderImpl) GetWorkflowsByUser(
	ctx context.Context,
	userId uuid.UUID,
	db database.Database,
) ([]Workflow, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM workflow WHERE user_id = $1;",
		allColumns())
	var workflows []Workflow

	err := db.Query(ctx, &workflows, query, userId)
	return workflows, err
}

// Workflows are uniquely keyed by (user_id, name).
// Returns nil if workflow is not found.
func (r *standardReaderImpl) GetWorkflowByName(
	ctx context.Context,
	userId uuid.UUID,
	name string,
	db database.Database,
) (*Workflow, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM workflow WHERE user_id = $1 and name = $2;",
		allColumns())

	var workflow Workflow
	err := db.Query(ctx, &workflow, query, userId, name)
	if err == database.ErrNoRows {
		return nil, nil
	}
	return &workflow, err
}

func (r *standardReaderImpl) GetAllWorkflows(
	ctx context.Context,
	db database.Database,
) ([]Workflow, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM workflow;",
		allColumns())
	var workflows []Workflow

	err := db.Query(ctx, &workflows, query)
	return workflows, err
}

// This function returns `True` if the organization owns the workflow and `False` otherwise.
func (r *standardReaderImpl) ValidateWorkflowOwnership(
	ctx context.Context,
	id uuid.UUID,
	organizationId string,
	db database.Database,
) (bool, error) {
	// Get the count of all workflows with the given id and user whose organization_id=organizationId.
	validateWorkflowOwnershipQuery := `SELECT COUNT(*) AS count
		FROM workflow INNER JOIN app_user ON workflow.user_id = app_user.id
		WHERE workflow.id = $1 AND app_user.organization_id = $2;`
	var count utils.CountResult

	err := db.Query(ctx, &count, validateWorkflowOwnershipQuery, id, organizationId)
	if err != nil {
		return false, err
	}

	isOwned := count.Count == 1
	if !isOwned {
		log.Infof("Workflow ownership query returned count %v.", count.Count)
	}

	return isOwned, nil
}

func (w *standardWriterImpl) UpdateWorkflow(
	ctx context.Context,
	id uuid.UUID, changes map[string]interface{},
	db database.Database,
) (*Workflow, error) {
	var workflow Workflow
	err := utils.UpdateRecordToDest(ctx, &workflow, changes, tableName, IdColumn, id, allColumns(), db)
	return &workflow, err
}

func (w *standardWriterImpl) DeleteWorkflow(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	deleteWorkflowStmt := `DELETE FROM workflow WHERE id = $1;`
	return db.Execute(ctx, deleteWorkflowStmt, id)
}

// Use to associate a workflow.name, workflow.id with workflow_dag_result.created_at (ENG-625)
func (r *standardReaderImpl) GetNotificationWorkflowMetadata(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) (map[uuid.UUID]NotificationWorkflowMetadata, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	// Get all workflow ids, workflow names, and workflow DAG ids that has a workflow DAG result id in the ids list.
	workflowsMetadataQuery := fmt.Sprintf(`
		SELECT workflow.id, workflow.name, workflow_dag_result.id AS dag_result_id
		FROM workflow, workflow_dag, workflow_dag_result 
		WHERE workflow_dag_result.workflow_dag_id = workflow_dag.id 
		AND workflow.id = workflow_dag.workflow_id 
		AND workflow_dag_result.id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var workflowsMetadata []NotificationWorkflowMetadata
	err := db.Query(ctx, &workflowsMetadata, workflowsMetadataQuery, args...)

	workflowsMetadataMap := make(map[uuid.UUID]NotificationWorkflowMetadata)
	for _, workflowMetadata := range workflowsMetadata {
		workflowsMetadataMap[workflowMetadata.DagResultId] = workflowMetadata
	}
	return workflowsMetadataMap, err
}

func (r *standardReaderImpl) GetWatchersInBatch(
	ctx context.Context,
	workflowIds []uuid.UUID,
	db database.Database,
) ([]WorkflowWatcherInfo, error) {
	if len(workflowIds) == 0 {
		return nil, errors.New("Provided empty workflow IDs list.")
	}

	// Get all user `auth0_id`s that are watching workflows whose `workflow_id` is in `workflowIds``
	workflowWatchersQuery := fmt.Sprintf(`
		SELECT workflow_watcher.workflow_id AS workflow_id, app_user.auth0_id
		FROM workflow_watcher, app_user 
		WHERE workflow_watcher.user_id = app_user.id 
		AND workflow_watcher.workflow_id IN (%s);`,
		stmt_preparers.GenerateArgsList(len(workflowIds), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(workflowIds)

	var workflowWatchers []WorkflowWatcherInfo
	err := db.Query(ctx, &workflowWatchers, workflowWatchersQuery, args...)
	return workflowWatchers, err
}
