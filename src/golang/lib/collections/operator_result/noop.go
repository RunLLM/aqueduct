package operator_result

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

func (w *noopWriterImpl) CreateOperatorResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	operatorId uuid.UUID,
	db database.Database,
) (*OperatorResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (r *noopReaderImpl) GetOperatorResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*OperatorResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetOperatorResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]OperatorResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetOperatorResultByWorkflowDagResultIdAndOperatorId(
	ctx context.Context,
	workflowDagResultId,
	operatorId uuid.UUID,
	db database.Database,
) (*OperatorResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetOperatorResultsByWorkflowDagResultIds(
	ctx context.Context,
	workflowDagResultIds []uuid.UUID,
	db database.Database,
) ([]OperatorResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (w *noopWriterImpl) UpdateOperatorResult(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*OperatorResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteOperatorResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteOperatorResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}
