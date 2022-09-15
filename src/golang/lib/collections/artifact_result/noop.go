package artifact_result

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/shared"
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

func (w *noopWriterImpl) CreateArtifactResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	artifactId uuid.UUID,
	contentPath string,
	db database.Database,
) (*ArtifactResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) InsertArtifactResult(
	ctx context.Context,
	workflowDagResultId uuid.UUID,
	artifactId uuid.UUID,
	contentPath string,
	execState *shared.ExecutionState,
	metadata *Metadata,
	db database.Database,
) (*ArtifactResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (r *noopReaderImpl) GetArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) (*ArtifactResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetArtifactResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) ([]ArtifactResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetArtifactResultsByArtifactId(
	ctx context.Context,
	artifactId uuid.UUID,
	db database.Database,
) ([]ArtifactResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetArtifactResultByWorkflowDagResultIdAndArtifactId(
	ctx context.Context,
	workflowDagResultId,
	artifactId uuid.UUID,
	db database.Database,
) (*ArtifactResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (r *noopReaderImpl) GetArtifactResultsByWorkflowDagResultIds(
	ctx context.Context,
	workflowDagResultIds []uuid.UUID,
	db database.Database,
) ([]ArtifactResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(r.throwError)
}

func (w *noopWriterImpl) UpdateArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	changes map[string]interface{},
	db database.Database,
) (*ArtifactResult, error) {
	return nil, utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteArtifactResult(
	ctx context.Context,
	id uuid.UUID,
	db database.Database,
) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}

func (w *noopWriterImpl) DeleteArtifactResults(
	ctx context.Context,
	ids []uuid.UUID,
	db database.Database,
) error {
	return utils.NoopInterfaceErrorHandling(w.throwError)
}
