package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/check"
	"github.com/aqueducthq/aqueduct/lib/collections/operator/function"
	"github.com/aqueducthq/aqueduct/lib/collections/shared"
	"github.com/aqueducthq/aqueduct/lib/database"
	exec_env "github.com/aqueducthq/aqueduct/lib/execution_environment"
	"github.com/aqueducthq/aqueduct/lib/job"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/aqueducthq/aqueduct/lib/vault"
	"github.com/aqueducthq/aqueduct/lib/workflow/artifact"
	"github.com/aqueducthq/aqueduct/lib/workflow/preview_cache"
	"github.com/aqueducthq/aqueduct/lib/workflow/utils"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type baseOperator struct {
	dbOperator *models.Operator

	// These fields are set to nil in the preview case.
	resultRepo repos.OperatorResult
	resultID   uuid.UUID

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
	execState        shared.ExecutionState

	// If set to nil, the job manager will run this operator in the server's default Python environment.
	// Otherwise, it will switch to the approppriate Conda environment before running the operator.
	// This only applies to operators running with the Aqueduct engine.
	execEnv *exec_env.ExecutionEnvironment
}

func (bo *baseOperator) Type() operator.Type {
	return bo.dbOperator.Spec.Type()
}

func (bo *baseOperator) Name() string {
	return bo.dbOperator.Name
}

func (bo *baseOperator) ID() uuid.UUID {
	return bo.dbOperator.ID
}

// A catch-all for execution states that are the system's fault.
// Logs an internal message so that we can debug.
func unknownSystemFailureExecState(err error, logMsg string) *shared.ExecutionState {
	log.Errorf("Execution had system failure: %s %v", logMsg, err)

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

func jobManagerUserFailureExecState(err job.JobError, logMsg string) *shared.ExecutionState {
	log.Errorf("Job execution had a user-facing issue: %s %v", logMsg, err)

	failureType := shared.UserFatalFailure
	return &shared.ExecutionState{
		Status:      shared.FailedExecutionStatus,
		FailureType: &failureType,

		// We only need to surface the user-facing message back the user,
		// since the stack trace is within our own code.
		Error: &shared.Error{
			Context: "",
			Tip:     err.GetMessage(),
		},
	}
}

func (bo *baseOperator) launch(ctx context.Context, spec job.Spec) error {
	if bo.execState.Status != shared.PendingExecutionStatus {
		return errors.Newf("Cannot launch operator with state %s", bo.execState.Status)
	}

	bo.updateExecState(&shared.ExecutionState{Status: shared.RunningExecutionStatus})

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

				// The op metadata path is updated for both this operator and its output artifacts.
				bo.outputExecPaths[i].OpMetadataPath = cacheEntry.OpMetadataPath
				bo.metadataPath = cacheEntry.OpMetadataPath
			}
			return nil
		}
	}

	return bo.jobManager.Launch(ctx, spec.JobName(), spec)
}

// fetchAndUpdateExecState assumes that the operator has been computed already.
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

// updateExecState and merge timestamps with current state based on the status of the new state.
// Other fields of bo.execState will be replaced.
func (bo *baseOperator) updateExecState(execState *shared.ExecutionState) {
	now := time.Now()
	// copy current timestamps to merge these time
	execTimestamps := bo.execState.Timestamps
	if execState.Terminated() {
		execTimestamps.FinishedAt = &now
	} else if execState.Status == shared.RunningExecutionStatus {
		execTimestamps.RunningAt = &now
	} else if execState.Status == shared.PendingExecutionStatus {
		execTimestamps.PendingAt = &now
	}

	execState.Timestamps = execTimestamps
	bo.execState = *execState
}

func updateOperatorResultAfterComputation(
	ctx context.Context,
	execState *shared.ExecutionState,
	opResultRepo repos.OperatorResult,
	opResultID uuid.UUID,
	db database.Database,
) {
	changes := map[string]interface{}{
		models.OperatorResultStatus:    execState.Status,
		models.OperatorResultExecState: execState,
	}

	_, err := opResultRepo.Update(
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
	if bo.resultRepo == nil {
		return errors.New("Operator's result writer cannot be nil.")
	}

	operatorResult, err := bo.resultRepo.Create(
		ctx,
		dagResultID,
		bo.ID(),
		&bo.execState,
		bo.db,
	)
	if err != nil {
		return errors.Wrap(err, "Failed to create operator result record.")
	}

	bo.resultID = operatorResult.ID

	return nil
}

func (bo *baseOperator) Poll(ctx context.Context) (*shared.ExecutionState, error) {
	if bo.jobName == "" {
		return nil, errors.Newf("Internal error: a job name was not set for this operator.")
	}

	// The operator is already terminated. No need to update status.
	if bo.execState.Terminated() {
		return bo.ExecState(), nil
	}

	status, err := bo.jobManager.Poll(ctx, bo.jobName)
	if err != nil {
		// If the job does not exist, this could mean that
		// 1) it is hasn't been run yet (pending),
		// 2) it has run already at sometime in the past, but has been garbage collected
		// 3) it has run already at sometime in the past, but did not go through the job manager.
		//    (this can happen when the output artifacts have already been cached).
		if err.Code() == job.JobMissing || err.Code() == job.Noop {
			// Check whether the operator has actually completed.
			if utils.ObjectExistsInStorage(ctx, bo.storageConfig, bo.metadataPath) {
				execState := bo.fetchExecState(ctx)
				bo.updateExecState(execState)
				return bo.ExecState(), nil
			}

			// Otherwise, return the current state of the operator (pending or running).
			return bo.ExecState(), nil

			// Catch any errors here that originate from within the JobManager, outside of the
			// python execution context, and that need a better user-facing message (eg. OOM issues).
		} else if err.Code() == job.User {
			bo.updateExecState(
				jobManagerUserFailureExecState(err, "Job manager failed due to user error."),
			)
			return bo.ExecState(), nil
		} else if err.Code() == job.System {
			bo.updateExecState(
				unknownSystemFailureExecState(err, "Unable to poll job manager."),
			)
			return bo.ExecState(), nil
		} else {
			return nil, errors.Newf("Unexpected JobErrorCode: %v", err.Code())
		}
	} else {
		// The job just completed, so we know we can fetch the results (succeeded/failed).
		if status == shared.FailedExecutionStatus || status == shared.SucceededExecutionStatus {
			execState := bo.fetchExecState(ctx)
			bo.updateExecState(execState)
			return bo.ExecState(), nil
		}

		// The job must exist at this point, but it hasn't completed (running).
		return bo.ExecState(), nil
	}
}

func (bo *baseOperator) ExecState() *shared.ExecutionState {
	return &bo.execState
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

	execState := bo.ExecState()
	if !execState.Terminated() {
		return errors.Newf("Operator %s is not terminated, so it does not have results that can be persisted.", bo.Name())
	}

	// Best effort writes after this point.
	updateOperatorResultAfterComputation(
		ctx,
		execState,
		bo.resultRepo,
		bo.resultID,
		bo.db,
	)

	for _, outputArtifact := range bo.outputs {
		// If the downstream artifact was never generated, we mark it as "cancelled", since the
		// operator either never ran or did run but hit a user-generated exception.
		// System-generated errors from things like the type checking of parameters will
		// still generate downstream artifacts, so those will continue to be marked as "failed".
		// Invariant: if an artifact is marked as failed, it's operator must also be marked failed,
		// with the same error message and context.
		artifactExecState := *execState
		if !outputArtifact.Computed(ctx) {
			artifactExecState.Status = shared.CanceledExecutionStatus
		}

		err := outputArtifact.PersistResult(ctx, &artifactExecState)
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

func (bo *baseOperator) Cancel() {
	bo.updateExecState(&shared.ExecutionState{
		Status: shared.CanceledExecutionStatus,
	})
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

	expectedOutputTypes := make([]string, 0, 1)
	for _, output := range bfo.outputs {
		expectedOutputTypes = append(expectedOutputTypes, string(output.Type()))
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
		EntryPointFile:              entryPoint.File,
		EntryPointClass:             entryPoint.ClassName,
		EntryPointMethod:            entryPoint.Method,
		CustomArgs:                  fn.CustomArgs,
		InputContentPaths:           inputContentPaths,
		InputMetadataPaths:          inputMetadataPaths,
		OutputContentPaths:          outputContentPaths,
		OutputMetadataPaths:         outputMetadataPaths,
		ExpectedOutputArtifactTypes: expectedOutputTypes,
		OperatorType:                bfo.Type(),
		CheckSeverity:               checkSeverity,
		ExecEnv:                     bfo.execEnv,
		Resources:                   bfo.dbOperator.Spec.Resources(),
	}
}
