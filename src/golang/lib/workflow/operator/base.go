package operator

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/check"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
	"github.com/aqueducthq/aqueduct/lib/collections/operator_result"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/preview_cache"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type baseOperator struct {
	dbOperator *operator.DBOperator

	// These fields are set to nil in the preview case.
	resultWriter operator_result.Writer
	resultID     uuid.UUID

	metadataPath string
	jobName      string

	inputs          []artifact.Artifact
	outputs         []artifact.Artifact
	inputExecPaths  []*utils.ExecPaths
	outputExecPaths []*utils.ExecPaths

	// The operator is cache-aware if this is non-nil.
	previewCacheManager preview_cache.CacheManager
	jobManager          job.JobManager
	vaultObject         vault.Vault
	storageConfig       *shared.StorageConfig
	db                  database.Database

	// This cannot be set if the operator is cache-aware, since this only happens in non-preview paths.
	resultsPersisted bool
	execMode         ExecutionMode
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

// A catch-all for execution states that are the system's fault.
// Logs an internal message so that we can debug.
func unknownSystemFailureExecState(err error, logMsg string) *shared.ExecutionState {
	log.Errorf("Execution had system failure: %s. %v", logMsg, err)

	failureType := shared.SystemFailure
	return &shared.ExecutionState{
		Status:      shared.FailedExecutionStatus,
		FailureType: &failureType,
		Error: &shared.Error{
			Context: fmt.Sprintf("%v", err),
			Tip:     shared.TipUnknownInternalError,
		},
	}
}

func (bo *baseOperator) launch(ctx context.Context, spec job.Spec) error {
	// Check if this operator can use previously cached results instead of computing for scratch.
	if bo.previewCacheManager != nil {
		outputArtifactSignatures := make([]uuid.UUID, 0, len(bo.outputs))
		for _, output := range bo.outputs {
			outputArtifactSignatures = append(outputArtifactSignatures, output.Signature())
		}

		allFound, cacheEntryByKey, err := bo.previewCacheManager.GetMulti(ctx, outputArtifactSignatures)
		if err != nil {
			log.Errorf("Unexpected error when querying the preview cache: %v", err)
		}

		if allFound {
			// Apply the cached results for each output artifact. This just means setting the output paths
			// to be the same as the cached ones.
			for i, outputArtifact := range bo.outputs {
				cacheEntry := cacheEntryByKey[outputArtifact.Signature()]
				bo.outputExecPaths[i].ArtifactMetadataPath = cacheEntry.ArtifactMetadataPath
				bo.outputExecPaths[i].ArtifactContentPath = cacheEntry.ArtifactContentPath

				// The op metadata path is updated for both for this operator and its output artifacts.
				bo.outputExecPaths[i].OpMetadataPath = cacheEntry.OpMetadataPath
				bo.metadataPath = cacheEntry.OpMetadataPath
			}
			return nil
		}
	}

	return bo.jobManager.Launch(ctx, spec.JobName(), spec)
}

// fetchExecState assumes that the operator has been computed already.
func (bo *baseOperator) fetchExecState(ctx context.Context) *shared.ExecutionState {
	var execState shared.ExecutionState
	err := utils.ReadFromStorage(
		ctx,
		bo.storageConfig,
		bo.metadataPath,
		&execState,
	)
	if err != nil {
		// Treat this as a system internal error since operator metadata was not found
		return unknownSystemFailureExecState(
			err,
			"Unable to read operator metadata from storage. Operator may have failed before writing metadata.",
		)
	}
	return &execState
}

// GetExecState takes a more wholelistic view of the operator's status than the job manager does,
// and can be called at any time. Because of this, the logic for figuring out the correct state is
// a little more involved.
func (bo *baseOperator) GetExecState(ctx context.Context) (*shared.ExecutionState, error) {
	if bo.jobName == "" {
		return nil, errors.Newf("Internal error: a job name was not set for this operator.")
	}

	status, err := bo.jobManager.Poll(ctx, bo.jobName)
	if err != nil {
		// If the job does not exist, this could mean that
		// 1) it is hasn't been run yet (pending),
		// 2) it has run already at sometime in the past, but has been garbage collected
		// 3) it has run already at sometime in the past, but did not go through the job manager.
		//    (this can happen when the output artifacts have already been cached).
		if err == job.ErrJobNotExist {
			// Check whether the operator has actually completed.
			if utils.ObjectExistsInStorage(ctx, bo.storageConfig, bo.metadataPath) {
				return bo.fetchExecState(ctx), nil
			}

			// Otherwise, this job has not run yet and is in a pending state.
			return &shared.ExecutionState{
				Status: shared.PendingExecutionStatus,
			}, nil
		} else {
			// This is just an internal polling error state.
			return unknownSystemFailureExecState(err, "Unable to poll job manager."), nil
		}
	} else {
		// The job just completed, so we know we can fetch the results (succeeded/failed).
		if status == shared.FailedExecutionStatus || status == shared.SucceededExecutionStatus {
			return bo.fetchExecState(ctx), nil
		}

		// The job must exist at this point, but it hasn't completed (running).
		return &shared.ExecutionState{
			Status: shared.RunningExecutionStatus,
		}, nil
	}
}

func updateOperatorResultAfterComputation(
	ctx context.Context,
	execState *shared.ExecutionState,
	opResultWriter operator_result.Writer,
	opResultID uuid.UUID,
	db database.Database,
) {
	changes := map[string]interface{}{
		operator_result.StatusColumn:    execState.Status,
		operator_result.ExecStateColumn: execState,
	}

	_, err := opResultWriter.UpdateOperatorResult(
		ctx,
		opResultID,
		changes,
		db,
	)
	if err != nil {
		log.WithFields(
			log.Fields{
				"changes": changes,
			},
		).Errorf("Unable to update operator result metadata: %v", err)
	}
}

func (bo *baseOperator) InitializeResult(ctx context.Context, dagResultID uuid.UUID) error {
	if bo.resultWriter == nil {
		return errors.New("Operator's result writer cannot be nil.")
	}

	operatorResult, err := bo.resultWriter.CreateOperatorResult(
		ctx,
		dagResultID,
		bo.ID(),
		bo.db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create operator result record.")
	}
	bo.resultID = operatorResult.Id
	return nil
}

func (bo *baseOperator) PersistResult(ctx context.Context) error {
	if bo.execMode == Preview {
		// We should not be persisting any result for preview operators.
		return errors.Newf("Operator %s cannot be persisted, as it is being previewed.", bo.Name())
	}

	if bo.previewCacheManager != nil {
		return errors.Newf("Operator %s is cache-aware, so it cannot be persisted.", bo.Name())
	}

	if bo.resultsPersisted {
		return errors.Newf("Operator %s was already persisted!", bo.Name())
	}

	execState, err := bo.GetExecState(ctx)
	if err != nil {
		return err
	}
	if execState.Status != shared.FailedExecutionStatus && execState.Status != shared.SucceededExecutionStatus {
		return errors.Newf("Operator %s has neither succeeded or failed, so it does not have results that can be persisted.", bo.Name())
	}

	// Best effort writes after this point.
	updateOperatorResultAfterComputation(
		ctx,
		execState,
		bo.resultWriter,
		bo.resultID,
		bo.db,
	)

	for _, outputArtifact := range bo.outputs {
		err = outputArtifact.PersistResult(ctx, execState)
		if err != nil {
			log.Errorf("Error occurred when persisting artifact %s.", outputArtifact.Name())
		}
	}
	bo.resultsPersisted = true
	return nil
}

func (bo *baseOperator) Finish(ctx context.Context) {
	// Delete the operator's metadata path only if it was already copied into the operator_result's table.
	// Otherwise, the artifact preview cache manager will handle the deletion.
	if bo.resultsPersisted {
		utils.CleanupStorageFile(ctx, bo.storageConfig, bo.metadataPath)
	}

	for _, outputArtifact := range bo.outputs {
		outputArtifact.Finish(ctx)
	}
}

// Any operator that runs a python function serialized from storage should use this instead of baseOperator.
type baseFunctionOperator struct {
	baseOperator
}

func (bfo *baseFunctionOperator) Finish(ctx context.Context) {
	// Delete the serialized function only for previews.
	if bfo.execMode == Preview {
		utils.CleanupStorageFile(ctx, bfo.storageConfig, bfo.dbOperator.Spec.Function().StoragePath)
	}

	bfo.baseOperator.Finish(ctx)
}

const (
	defaultFunctionEntryPointFile   = "model.py"
	defaultFunctionEntryPointClass  = "Function"
	defaultFunctionEntryPointMethod = "predict"
)

func (bfo *baseFunctionOperator) jobSpec(
	fn *function.Function,
	checkSeverity *check.Level,
) job.Spec {
	entryPoint := fn.EntryPoint
	if entryPoint == nil {
		entryPoint = &function.EntryPoint{
			File:      defaultFunctionEntryPointFile,
			ClassName: defaultFunctionEntryPointClass,
			Method:    defaultFunctionEntryPointMethod,
		}
	}

	inputContentPaths, inputMetadataPaths := unzipExecPathsToRawPaths(bfo.inputExecPaths)
	outputContentPaths, outputMetadataPaths := unzipExecPathsToRawPaths(bfo.outputExecPaths)
	return &job.FunctionSpec{
		BasePythonSpec: job.NewBasePythonSpec(
			job.FunctionJobType,
			bfo.jobName,
			*bfo.storageConfig,
			bfo.metadataPath,
		),
		FunctionPath: fn.StoragePath,
		/* `FunctionExtractPath` is set by the job manager at launch time. */
		EntryPointFile:      entryPoint.File,
		EntryPointClass:     entryPoint.ClassName,
		EntryPointMethod:    entryPoint.Method,
		CustomArgs:          fn.CustomArgs,
		InputContentPaths:   inputContentPaths,
		InputMetadataPaths:  inputMetadataPaths,
		OutputContentPaths:  outputContentPaths,
		OutputMetadataPaths: outputMetadataPaths,
		OperatorType:        bfo.Type(),
		CheckSeverity:       checkSeverity,
	}
}
