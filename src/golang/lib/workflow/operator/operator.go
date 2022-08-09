package operator

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

// Operator is an interface for managing and inspecting the lifecycle of an operator
// used by a workflow run.
type Operator interface {
	Type() operator.Type
	Name() string
	ID() uuid.UUID
	JobSpec() job.Spec

	// GetExecState performs a non-blocking fetch for the execution state of this operator.
	GetExecState(ctx context.Context) (*shared.ExecutionState, error)

	// InitializeResult initializes the operator in the database.
	InitializeResult(ctx context.Context, dagResultID uuid.UUID) error

	// PersistResult writes the results of this operator execution to the database.
	// Errors if the artifact hasn ot yet been computed, or InitializeResult() hasn't been called yet.
	// *This method also persists any artifact results produced by this operator.*
	PersistResult(ctx context.Context) error

	// Finish is an end-of-lifecycle hook meant to do any final cleanup work.
	// Also calls Finish() on all the operator's output artifacts.
	Finish(ctx context.Context)
}

func NewOperator(
	ctx context.Context,
	dbOperator operator.DBOperator,
	inputs []artifact.Artifact,
	inputContentPaths []string,
	inputMetadataPaths []string,
	outputs []artifact.Artifact,
	outputContentPaths []string,
	outputMetadataPaths []string,
	opResultWriter operator_result.Writer, // A nil value means the operator is run in preview mode.
	jobManager job.JobManager,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	isPreview bool,
	db database.Database,
) (Operator, error) {
	if len(inputs) != len(inputContentPaths) || len(inputs) != len(inputMetadataPaths) {
		return nil, errors.New("Internal error: mismatched number of input arguments.")
	}
	if len(outputs) != len(outputContentPaths) || len(outputs) != len(outputMetadataPaths) {
		return nil, errors.New("Internal error: mismatched number of output arguments.")
	}

	baseOp := baseOperator{
		dbOperator:   &dbOperator,
		resultWriter: opResultWriter,
		resultID:     uuid.Nil,
		metadataPath: uuid.New().String(),
		jobName:      "", /* Must be set by the specific type constructors below. */

		inputs:              inputs,
		outputs:             outputs,
		inputContentPaths:   inputContentPaths,
		inputMetadataPaths:  inputMetadataPaths,
		outputContentPaths:  outputContentPaths,
		outputMetadataPaths: outputMetadataPaths,

		jobManager:    jobManager,
		vaultObject:   vaultObject,
		storageConfig: storageConfig,
		db:            db,

		resultsPersisted: false,
		isPreview:        isPreview,
	}

	if dbOperator.Spec.IsFunction() {
		return newFunctionOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsMetric() {
		return newMetricOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsCheck() {
		return newCheckOperator(baseFunctionOperator{baseOp})
	} else if dbOperator.Spec.IsExtract() {
		return newExtractOperator(ctx, baseOp)
	} else if dbOperator.Spec.IsLoad() {
		return newLoadOperator(ctx, baseOp)
	} else if dbOperator.Spec.IsParam() {
		return newParamOperator(baseOp)
	} else if dbOperator.Spec.IsSystemMetric() {
		return newSystemMetricOperator(baseOp)
	}

	return nil, errors.Newf("Unsupported operator type %s", dbOperator.Spec.Type())
}
