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

	// These fields are set dynamically over the lifecycle of the operator:

	// Only set if the operator is scheduled.
	jobName string
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

func (bo *baseOperatorFields) ExecState() (*shared.ExecutionState, error) {
	status, err := bo.jobManager.Poll(bo.ctx, bo.jobName)
	if err != nil {
		return nil, err
	}
	if status == shared.SucceededExecutionStatus || status == shared.FailedExecutionStatus {
		var execState shared.ExecutionState
		err := utils.ReadFromStorage(
			bo.ctx,
			bo.storageConfig,
			bo.opMetadataPath,
			&execState,
		)
		if err != nil {
			// Treat this as a system internal error since operator metadata was not found
			log.Errorf(
				"Unable to read operator metadata from storage. Operator may have failed before writing metadata. %v",
				err,
			)

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

func (bo *baseOperatorFields) Finish() error {
	// No-op if this is a preview operator.
	if bo.isPreview {
		return nil
	}

	if len(bo.jobName) == 0 {
		return errors.Newf("Unable to finish operator %s. It was never scheduled.", bo.Name())
	}

	execState, err := bo.ExecState()
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

	for _, outputArtifact := range bo.outputs {
		err = outputArtifact.Persist(execState.Status)
		if err != nil {
			log.Errorf(fmt.Sprintf("Error occurred when persisting artifact %s.", outputArtifact.Name()))
		}
	}
	return nil
}
