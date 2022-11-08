package execution_environment

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

func (w *noopWriterImpl) CreateExecutionEnvironment(
	ctx context.Context,
	spec Spec, hash string,
	db database.Database,
) (*DBExecutionEnvironment, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (r *noopReaderImpl) GetExecutionEnvironment(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*DBExecutionEnvironment, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetExecutionEnvironments(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]DBExecutionEnvironment, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (w *noopWriterImpl) UpdateExecutionEnvironment(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*DBExecutionEnvironment, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteExecutionEnvironment(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteExecutionEnvironments(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}
