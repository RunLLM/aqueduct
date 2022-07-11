package operator

import (
	"context"
	"fmt"
	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/scheduler"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// This Operator interface allows a caller to manage and inspect the lifecycle of
// a single operator in a workflow run.
type Operator interface {
	Type() operator.Type
	Name() string
	ID() uuid.UUID

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

	// Lists immediate upstream and downstream dependencies.
	Inputs() []artifact.Artifact
	Outputs() []artifact.Artifact
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

type baseOperator struct {
	ctx        context.Context
	dbOperator *operator.DBOperator

	// These fields are nil in the preview case.
	opResultWriter operator_result.Writer
	opResultID     uuid.UUID

	isPreview      bool
	opMetadataPath string

	inputs              []artifact.Artifact
	outputs             []artifact.Artifact
	inputContentPaths   []string
	inputMetadataPaths  []string
	outputContentPaths  []string
	outputMetadataPaths []string

	jobManager    job.JobManager
	vaultObject   vault.Vault
	storageConfig *shared.StorageConfig
	db            database.Database

	// This field is only set after this operator has launched.
	jobName string
}

func (bo *baseOperator) Type() operator.Type {
	return bo.dbOperator.Spec.Type()
}

func (bo *baseOperator) Name() string {
	return bo.dbOperator.Name
}

func (bo *baseOperator) ID() uuid.UUID {
	return bo.dbOperator.Id
}

func (bo *baseOperator) Inputs() []artifact.Artifact {
	return bo.inputs
}

func (bo *baseOperator) Outputs() []artifact.Artifact {
	return bo.outputs
}

func (bo *baseOperator) Ready() bool {
	for _, inputArtifact := range bo.inputs {
		if !inputArtifact.Computed() {
			return false
		}
	}
	return true
}

func (bo *baseOperator) Finish() error {
	if bo.isPreview {
		return errors.Newf(fmt.Sprintf("Cannot persist the results of operator %s in preview-mode.", bo.Name()))
	}

	status, err := bo.Status()
	if err != nil {
		return err
	}
	if status != shared.FailedExecutionStatus && status != shared.SucceededExecutionStatus {
		return errors.Newf(fmt.Sprintf("Operator %s has neither succeeded or failed, so it does not have results that can be persisted.", bo.Name()))
	}

	// Best effort writes after this point.
	utils.UpdateOperatorResultAfterComputation(
		bo.ctx,
		status,
		bo.storageConfig,
		bo.opMetadataPath,
		bo.opResultWriter,
		bo.opResultID,
		bo.db,
	)

	for _, outputArtifact := range bo.outputs {
		err = outputArtifact.Persist(status)
		if err != nil {
			log.Errorf(fmt.Sprintf("Error occurred when persisting artifact %s.", outputArtifact.Name()))
		}
	}
	return nil
}

func (bo *baseOperator) Status() (shared.ExecutionStatus, error) {
	// TODO(kenxu): I think the stuff in `CheckOperatorExecutionStatus()` belongs here.

	return bo.jobManager.Poll(bo.ctx, bo.jobName)
}

type functionOperatorImpl struct {
	baseOperator
}

func (fo *functionOperatorImpl) Schedule() error {
	inputArtifactTypes := make([]db_artifact.Type, 0, len(fo.inputs))
	outputArtifactTypes := make([]db_artifact.Type, 0, len(fo.outputs))
	for _, inputArtifact := range fo.inputs {
		inputArtifactTypes = append(inputArtifactTypes, inputArtifact.Type())
	}
	for _, outputArtifact := range fo.outputs {
		outputArtifactTypes = append(outputArtifactTypes, outputArtifact.Type())
	}

	jobName, err := scheduler.ScheduleFunction(
		fo.ctx,
		*fo.dbOperator.Spec.Function(),
		fo.opMetadataPath,
		fo.inputContentPaths,
		fo.inputMetadataPaths,
		fo.outputContentPaths,
		fo.outputMetadataPaths,
		inputArtifactTypes,
		outputArtifactTypes,
		fo.storageConfig,
		fo.jobManager,
	)
	if err != nil {
		return err
	}

	fo.jobName = jobName
	return nil
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
// A nil value here means the operator is not persisted.
	opResultWriter operator_result.Writer,
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

	base := baseOperator{
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
		for _, inputArtifact := range inputs {
			if inputArtifact.Type() != db_artifact.TableType && inputArtifact.Type() != db_artifact.JsonType {
				return nil, errors.New("Inputs to function operator must be Table or Parameter Artifacts.")
			}
		}
		for _, outputArtifact := range outputs {
			if outputArtifact.Type() != db_artifact.TableType {
				return nil, errors.New("Outputs of function operator must be Table Artifacts.")
			}
		}

		// TODO: Validate the number of inputs.
		return &functionOperatorImpl{
			base,
		}, nil
	}
	return nil, errors.Newf("Unsupported operator type %s", dbOperator.Spec.Type())
}
