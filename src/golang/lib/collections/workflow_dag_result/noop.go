package workflow_dag_result

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/notification"
	"github.com/aqueducthq/aqueduct/lib/collections/user"
	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/google/uuid"
)

type noopReaderImpl struct {
	throwError bool
}

type noopWriterImpl struct {
	throwError bool
}

func NewNoopReader(throwError bool) Reader {
	return &noopReaderImpl{throwError: throwError}
}

func NewNoopWriter(throwError bool) Writer {
	return &noopWriterImpl{throwError: throwError}
}

func (w *noopWriterImpl) CreateWorkflowDagResult(
	ctx context.Context,
	workflowDagId uuid.UUID,
	db database.Database,
) (*WorkflowDagResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (r *noopReaderImpl) GetWorkflowDagResult(ctx context.Context, id uuid.UUID, db database.Database) (*WorkflowDagResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWorkflowDagResults(ctx context.Context, ids []uuid.UUID, db database.Database) ([]WorkflowDagResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWorkflowDagResultsByWorkflowId(ctx context.Context, workflowId uuid.UUID, db database.Database) ([]WorkflowDagResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetKOffsetWorkflowDagResultsByWorkflowId(
	ctx context.Context,
	workflowId uuid.UUID,
	k int,
	db database.Database,
) ([]WorkflowDagResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (w *noopWriterImpl) UpdateWorkflowDagResult(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	workflowReader workflow.Reader,
	notificationWriter notification.Writer,
	userReader user.Reader,
	db database.Database,
) (*WorkflowDagResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteWorkflowDagResult(ctx context.Context, id uuid.UUID, db database.Database) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteWorkflowDagResults(ctx context.Context, ids []uuid.UUID, db database.Database) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}
