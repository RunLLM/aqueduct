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

// This Operator interface allows a caller to manage and inspect the lifecycle of
// a single operator in a workflow run.
type Operator interface {
	Type() operator.Type
	Name() string
	ID() uuid.UUID

	// Lists immediate upstream and downstream dependencies.
	Inputs() []artifact.Artifact
	Outputs() []artifact.Artifact

	// Indicates whether this operator is can be scheduled. This means that all
	// dependencies to this operator have already been computed.
	Ready() bool

	// Kicks off the job that executes this operator.
	// Errors if the operator is not ready.
	Schedule() error

	// Performs a non-blocking fetch of the execution state of this operator.
	ExecState() (*shared.ExecutionState, error)

	// An additional hook that should be called after the operator has terminated execution,
	// regardless of whether it ran successfully or not. This allows the operator to perform
	// any final database writes or cleanup operations. This can only be called once.
	Finish() error
}

func initializeOperatorResultInDatabase(
	ctx context.Context,
	opID uuid.UUID,
	workflowDagResultID uuid.UUID,
	opResultWriter operator_result.Writer,
	db database.Database,
) (uuid.UUID, error) {
	operatorResult, err := opResultWriter.CreateOperatorResult(
		ctx,
		workflowDagResultID,
		opID,
		db,
	)
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "Failed to create operator result record.")
	}
	return operatorResult.Id, nil
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
	workflowDagResultID uuid.UUID,
	jobManager job.JobManager,
	vaultObject vault.Vault,
	storageConfig *shared.StorageConfig,
	db database.Database,
) (Operator, error) {
	var opResultID uuid.UUID
	if workflowDagResultID != uuid.Nil {
		var err error
		opResultID, err = initializeOperatorResultInDatabase(
			ctx,
			dbOperator.Id,
			workflowDagResultID,
			opResultWriter,
			db,
		)
		if err != nil {
			return nil, err
		}
	}

	baseFields := baseOperatorFields{
		dbOperator:          &dbOperator,
		opResultWriter:      opResultWriter,
		opResultID:          opResultID,
		opMetadataPath:      uuid.New().String(),
		inputs:              inputs,
		outputs:             outputs,
		inputContentPaths:   inputContentPaths,
		inputMetadataPaths:  inputMetadataPaths,
		outputContentPaths:  outputContentPaths,
		outputMetadataPaths: outputMetadataPaths,
		jobManager:          jobManager,
		vaultObject:         vaultObject,
		storageConfig:       storageConfig,
		db:                  db,
	}

	if dbOperator.Spec.IsFunction() {
		return newFunctionOperator(baseFields, inputs, outputs)
	}
	return nil, errors.Newf("Unsupported operator type %s", dbOperator.Spec.Type())
}
