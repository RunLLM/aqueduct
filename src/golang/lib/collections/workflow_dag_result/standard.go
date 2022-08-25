package workflow_dag_result

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type standardReaderImpl struct{}

type standardWriterImpl struct{}

func (w *standardWriterImpl) CreateWorkflowDagResult(
	ctx context.Context,
	workflowDagId uuid.UUID,
	db database.Database,
) (*WorkflowDagResult, error) {
	insertColumns := []string{WorkflowDagIdColumn, StatusColumn, CreatedAtColumn}
	insertWorkflowDagResultStmt := db.PrepareInsertWithReturnAllStmt(tableName, insertColumns, allColumns())

	args := []interface{}{workflowDagId, shared.PendingExecutionStatus, time.Now()}

	var workflowDagResult WorkflowDagResult
	err := db.Query(ctx, &workflowDagResult, insertWorkflowDagResultStmt, args...)
	return &workflowDagResult, err
}

func (r *standardReaderImpl) GetWorkflowDagResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*WorkflowDagResult, error) {
	workflowDagResults, err := r.GetWorkflowDagResults(ctx, []uuid.UUID{id}, db)
	if err != nil {
		return nil, err
	}

	if len(workflowDagResults) != 1 {
		return nil, errors.Newf(
			"Expected 1 workflow_dag_result, but got %d workflow_dag_results.",
			len(workflowDagResults),
		)
	}

	return &workflowDagResults[0], nil
}

func (r *standardReaderImpl) GetWorkflowDagResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]WorkflowDagResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("Provided empty IDs list.")
	}

	getWorkflowDagResultsQuery := fmt.Sprintf(
		"SELECT %s FROM workflow_dag_result WHERE id IN (%s);",
		allColumns(),
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)

	var workflowDagResults []WorkflowDagResult
	err := db.Query(ctx, &workflowDagResults, getWorkflowDagResultsQuery, args...)
	return workflowDagResults, err
}

func (r *standardReaderImpl) GetWorkflowDagResultsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	db database.Database,
) ([]WorkflowDagResult, error) {
	// Get all workflow DAGs for the workflow specified by `workflowId`
	query := fmt.Sprintf(`
		SELECT %s FROM workflow_dag_result, workflow_dag 
		WHERE workflow_dag_result.workflow_dag_id = workflow_dag.id AND workflow_dag.workflow_id = $1;`,
		allColumnsWithPrefix())

	var workflowDagResults []WorkflowDagResult
	err := db.Query(ctx, &workflowDagResults, query, workflowId)
	return workflowDagResults, err
}

func (r *standardReaderImpl) GetKOffsetWorkflowDagResultsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	k int,
	db database.Database,
) ([]WorkflowDagResult, error) {
	// Get all workflow DAGs for the workflow specified by `workflowId` except for the k latest.
	query := fmt.Sprintf(`
		SELECT %s FROM workflow_dag_result, workflow_dag 
		WHERE workflow_dag_result.workflow_dag_id = workflow_dag.id AND workflow_dag.workflow_id = $1
		ORDER BY workflow_dag_result.created_at DESC
		OFFSET $2;`,
		allColumnsWithPrefix())

	var workflowDagResults []WorkflowDagResult
	err := db.Query(ctx, &workflowDagResults, query, workflowId, k)
	return workflowDagResults, err
}

func workflowDagResultNotificationContent(
	workflowObject *workflow.Workflow,
	workflowDagResult *WorkflowDagResult,
) string {
	status := workflowDagResult.Status
	if status != shared.SucceededExecutionStatus && status != shared.FailedExecutionStatus {
		return ""
	}

	name := workflowObject.Name
	if status == shared.SucceededExecutionStatus {
		return fmt.Sprintf("Workflow %s has succeeded!", name)
	}

	return fmt.Sprintf("Workflow %s has failed.", name)
}

func createWorkflowDagResultNotification(
	ctx context.Context,
	workflowDagResult *WorkflowDagResult,
	notificationWriter notification.Writer,
	workflowReader workflow.Reader,
	userReader user.Reader,
	db database.Database,
) error {
	status := workflowDagResult.Status
	if status != shared.SucceededExecutionStatus && status != shared.FailedExecutionStatus {
		return nil
	}

	workflowObject, err := workflowReader.GetWorkflowByWorkflowDagId(ctx, workflowDagResult.WorkflowDagId, db)
	if err != nil {
		return err
	}

	notificationLevel := notification.SuccessLevel

	if status == shared.FailedExecutionStatus {
		notificationLevel = notification.ErrorLevel
	}

	notificationAssociation := notification.NotificationAssociation{
		Object: notification.WorkflowDagResultObject,
		Id:     workflowDagResult.Id,
	}

	workflow_watchers, err := userReader.GetWatchersByWorkflowId(
		ctx,
		workflowObject.Id,
		db,
	)
	for _, single_watcher := range workflow_watchers {
		if err != nil {
			return err
		}
		_, err = notificationWriter.CreateNotification(
			ctx,
			single_watcher.Id,
			workflowDagResultNotificationContent(workflowObject, workflowDagResult),
			notificationLevel,
			notificationAssociation,
			db,
		)
	}
	return err
}

func (w *standardWriterImpl) UpdateWorkflowDagResult(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	workflowReader workflow.Reader,
	notificationWriter notification.Writer,
	userReader user.Reader,
	db database.Database,
) (*WorkflowDagResult, error) {
	var workflowDagResult WorkflowDagResult
	err := utils.UpdateRecordToDest(ctx, &workflowDagResult, changes, tableName, IdColumn, id, allColumns(), db)
	if err != nil {
		return nil, err
	}

	err = createWorkflowDagResultNotification(ctx, &workflowDagResult, notificationWriter, workflowReader, userReader, db)
	if err != nil {
		// Only log the error and hide it to caller, since the dag result itself is successfully updated
		log.Errorf("Failed to create dag result notification: %s", err)
	}

	return &workflowDagResult, nil
}

func (w *standardWriterImpl) DeleteWorkflowDagResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return w.DeleteWorkflowDagResults(ctx, []uuid.UUID{id}, db)
}

func (w *standardWriterImpl) DeleteWorkflowDagResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	if len(ids) == 0 {
		return nil
	}

	deleteStmt := fmt.Sprintf(
		"DELETE FROM workflow_dag_result WHERE id IN (%s);",
		stmt_preparers.GenerateArgsList(len(ids), 1),
	)

	args := stmt_preparers.CastIdsListToInterfaceList(ids)
	return db.Execute(ctx, deleteStmt, args...)
}
