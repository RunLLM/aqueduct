package operator

import (
	"context"
	"fmt"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
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

type baseOperatorFields struct {
	ctx        context.Context
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
}

func (bo *baseOperatorFields) Type() operator.Type {
	return bo.dbOperator.Spec.Type()
}

func (bo *baseOperatorFields) Name() string {
	return bo.dbOperator.Name
}

func (bo *baseOperatorFields) ID() uuid.UUID {
	return bo.dbOperator.Id
}

func (bo *baseOperatorFields) Inputs() []artifact.Artifact {
	return bo.inputs
}

func (bo *baseOperatorFields) Outputs() []artifact.Artifact {
	return bo.outputs
}

func (bo *baseOperatorFields) Ready() bool {
	for _, inputArtifact := range bo.inputs {
		if !inputArtifact.Computed() {
			return false
		}
	}
	return true
}

func (bo *baseOperatorFields) GetExecState() (*shared.ExecutionState, error) {
	status, err := bo.jobManager.Poll(bo.ctx, bo.jobName)
	if err != nil {
		return nil, err
	}
	if status == shared.SucceededExecutionStatus || status == shared.FailedExecutionStatus {
		var execState shared.ExecutionState
		err = utils.ReadFromStorage(
			bo.ctx,
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

func (bo *baseOperatorFields) PersistResult() error {
	execState, err := bo.GetExecState()
	if err != nil {
		return err
	}
	if execState.Status != shared.FailedExecutionStatus && execState.Status != shared.SucceededExecutionStatus {
		return errors.Newf(fmt.Sprintf("Operator %s has neither succeeded or failed, so it does not have results that can be persisted.", bo.Name()))
	}

	// Best effort writes after this point.
	utils.UpdateOperatorResultAfterComputation(
		bo.ctx,
		execState.Status,
		bo.storageConfig,
		bo.opMetadataPath,
		bo.opResultWriter,
		bo.opResultID,
		bo.db,
	)

	// TODO: move this to artifact persist.
	for _, outputArtifact := range bo.outputs {
		err = outputArtifact.PersistResult(execState.Status)
		if err != nil {
			log.Errorf(fmt.Sprintf("Error occurred when persisting artifact %s.", outputArtifact.Name()))
		}
	}
	return nil
}
