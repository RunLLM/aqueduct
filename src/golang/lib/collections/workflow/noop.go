package workflow

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
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

func (r *noopReaderImpl) Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error) {
	return false, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (w *noopWriterImpl) CreateWorkflow(
	ctx context.Context,
	userId uuid.UUID,
	name string,
	description string,
	schedule *Schedule,
	retentionPolicy *RetentionPolicy,
	db database.Database,
) (*Workflow, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (r *noopReaderImpl) GetWorkflow(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*Workflow, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWorkflowByWorkflowDagId(
	ctx context.Context,
	workflowDagId uuid.UUID,
	db database.Database,
) (*Workflow, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWorkflows(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]Workflow, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetNotificationWorkflowMetadata(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) (map[uuid.UUID]NotificationWorkflowMetadata, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWorkflowsByUser(
	ctx context.Context,
	userId uuid.UUID,
	db database.Database,
) ([]Workflow, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWorkflowByName(
	ctx context.Context,
	userId uuid.UUID,
	name string,
	db database.Database,
) (*Workflow, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetAllWorkflows(
	ctx context.Context,
	db database.Database,
) ([]Workflow, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) ValidateWorkflowOwnership(
	ctx context.Context,
	id uuid.UUID,
	organizationId string,
	db database.Database,
) (bool, error) {
	return true, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (w *noopWriterImpl) UpdateWorkflow(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*Workflow, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteWorkflow(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}

func (r *noopReaderImpl) GetWorkflowsWithLatestRunResult(
	ctx context.Context,
	organizationId string,
	db database.Database,
) ([]latestWorkflowResponse, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWatchersInBatch(
	ctx context.Context,
	workflowIds []uuid.UUID,
	db database.Database,
) ([]WorkflowWatcherInfo, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetWorkflowsFromOperatorIds(
	ctx context.Context,
	operatorIds []uuid.UUID,
	db database.Database,
) (map[uuid.UUID][]uuid.UUID, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}
