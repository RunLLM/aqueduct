package operator

import (
	"context"
	"fmt"
	db_artifact "github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type baseOperator struct {
	dbOperator *operator.DBOperator

	// These fields are set to nil in the preview case.
	opResultWriter operator_result.Writer
	opResultID     uuid.UUID

	isPreview      bool
	opMetadataPath string
	jobName        string

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

	resultsPersisted bool
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

func (bo *baseOperator) Ready(ctx context.Context) bool {
	for _, inputArtifact := range bo.inputs {
		if !inputArtifact.Computed(ctx) {
			return false
		}
	}
	return true
}

func (bo *baseOperator) GetExecState(ctx context.Context) (*shared.ExecutionState, error) {
	if bo.jobName == "" {
		return nil, errors.Newf("Internal error: a jobname was not set for this operator.")
	}

	status, err := bo.jobManager.Poll(ctx, bo.jobName)
	if err != nil {
		return nil, err
	}
	if status == shared.SucceededExecutionStatus || status == shared.FailedExecutionStatus {
		var execState shared.ExecutionState
		err = utils.ReadFromStorage(
			ctx,
			bo.storageConfig,
			bo.opMetadataPath,
			&execState,
		)

		if err != nil {
			if err != job.ErrJobNotExist {
				// The job already finished somehow and was garbage-collected.
				log.Errorf("Job %s does not exist for operator %s", bo.jobName, bo.Name())
			} else {
				// Treat this as a system internal error since operator metadata was not found
				log.Errorf(
					"Unable to read operator metadata from storage. Operator may have failed before writing metadata. %v",
					err,
				)
			}

			failureType := shared.SystemFailure
			return &shared.ExecutionState{
				Status:      shared.FailedExecutionStatus,
				FailureType: &failureType,
				Error: &shared.Error{
					Context: fmt.Sprintf("%v", err),
					Tip:     shared.TipUnknownInternalError,
				},
			}, nil
		}
	}

	// For pending and running operators.
	return &shared.ExecutionState{
		Status: status,
	}, nil

}

func (bo *baseOperator) PersistResult(ctx context.Context) error {
	if bo.resultsPersisted {
		return errors.Newf("Operator %s was already persisted!", bo.Name())
	}

	execState, err := bo.GetExecState(ctx)
	if err != nil {
		return err
	}
	if execState.Status != shared.FailedExecutionStatus && execState.Status != shared.SucceededExecutionStatus {
		return errors.Newf(fmt.Sprintf("Operator %s has neither succeeded or failed, so it does not have results that can be persisted.", bo.Name()))
	}

	// Best effort writes after this point.
	utils.UpdateOperatorResultAfterComputation(
		ctx,
		execState.Status,
		bo.storageConfig,
		bo.opMetadataPath,
		bo.opResultWriter,
		bo.opResultID,
		bo.db,
	)

	for _, outputArtifact := range bo.outputs {
		err = outputArtifact.PersistResult(ctx, execState.Status)
		if err != nil {
			log.Errorf(fmt.Sprintf("Error occurred when persisting artifact %s.", outputArtifact.Name()))
		}
	}
	bo.resultsPersisted = true
	return nil
}

func (bo *baseOperator) Finish(ctx context.Context) {
	utils.CleanupStorageFile(ctx, bo.storageConfig, bo.opMetadataPath)

	for _, outputArtifact := range bo.outputs {
		outputArtifact.Finish(ctx)
	}
}

// Any operator that runs a python function serialized from storage should use this instead of baseOperator.
type baseFunctionOperator struct {
	baseOperator
}

func (bfo *baseFunctionOperator) Finish(ctx context.Context) {
	// If the operator was not persisted to the DB, cleanup the serialized function.
	if !bfo.resultsPersisted {
		utils.CleanupStorageFile(ctx, bfo.storageConfig, bfo.dbOperator.Spec.Function().StoragePath)
	}

	bfo.baseOperator.Finish(ctx)
}

const (
	defaultFunctionEntryPointFile   = "model.py"
	defaultFunctionEntryPointClass  = "Function"
	defaultFunctionEntryPointMethod = "predict"
)

func (bfo *baseFunctionOperator) jobSpec(fn *function.Function) job.Spec {
	entryPoint := fn.EntryPoint
	if entryPoint == nil {
		entryPoint = &function.EntryPoint{
			File:      defaultFunctionEntryPointFile,
			ClassName: defaultFunctionEntryPointClass,
			Method:    defaultFunctionEntryPointMethod,
		}
	}

	inputArtifactTypes := make([]db_artifact.Type, 0, len(bfo.inputs))
	outputArtifactTypes := make([]db_artifact.Type, 0, len(bfo.outputs))
	for _, inputArtifact := range bfo.inputs {
		inputArtifactTypes = append(inputArtifactTypes, inputArtifact.Type())
	}
	for _, outputArtifact := range bfo.outputs {
		outputArtifactTypes = append(outputArtifactTypes, outputArtifact.Type())
	}

	return &job.FunctionSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.FunctionJobType,
			bfo.jobName,
			*bfo.storageConfig,
			bfo.opMetadataPath,
		),
		FunctionPath:        fn.StoragePath,
		EntryPointFile:      entryPoint.File,
		EntryPointClass:     entryPoint.ClassName,
		EntryPointMethod:    entryPoint.Method,
		CustomArgs:          fn.CustomArgs,
		InputContentPaths:   bfo.inputContentPaths,
		InputMetadataPaths:  bfo.inputMetadataPaths,
		OutputContentPaths:  bfo.outputContentPaths,
		OutputMetadataPaths: bfo.outputMetadataPaths,
		InputArtifactTypes:  inputArtifactTypes,
		OutputArtifactTypes: outputArtifactTypes,
	}

}
